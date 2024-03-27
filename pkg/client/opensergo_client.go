// Copyright 2022, OpenSergo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/groupcache"
	"github.com/opensergo/opensergo-go/pkg/api"
	"github.com/opensergo/opensergo-go/pkg/common/logging"
	"github.com/opensergo/opensergo-go/pkg/model"
	transportPb "github.com/opensergo/opensergo-go/pkg/proto/transport/v1"
	"github.com/opensergo/opensergo-go/pkg/transport/subscribe"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OpenSergoClient is the client to communicate with opensergo-control-plane.
type OpenSergoClient struct {
	host                   string
	port                   uint32
	transportServiceClient transportPb.OpenSergoUniversalTransportServiceClient

	subscribeConfigStreamPtr         atomic.Value // type of value is *client.subscribeConfigStream
	subscribeConfigStreamObserverPtr atomic.Value // type of value is *client.SubscribeConfigStreamObserver
	subscribeConfigStreamStatus      atomic.Value // type of value is client.OpensergoClientStreamStatus

	subscribeDataCache *subscribe.SubscribeDataCache
	subscriberRegistry *subscribe.SubscriberRegistry

	requestId groupcache.AtomicInt
}

// NewOpenSergoClient returns an instance of OpenSergoClient, and init some properties.
func NewOpenSergoClient(host string, port uint32) (*OpenSergoClient, error) {
	address := host + ":" + strconv.FormatUint(uint64(port), 10)
	clientConn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	transportServiceClient := transportPb.NewOpenSergoUniversalTransportServiceClient(clientConn)

	openSergoClient := &OpenSergoClient{
		host:                   host,
		port:                   port,
		transportServiceClient: transportServiceClient,
		subscribeDataCache:     &subscribe.SubscribeDataCache{},
		subscriberRegistry:     &subscribe.SubscriberRegistry{},
		requestId:              groupcache.AtomicInt(0),
	}
	openSergoClient.subscribeConfigStreamPtr.Store(&subscribeConfigStream{})
	openSergoClient.subscribeConfigStreamObserverPtr.Store(&subscribeConfigStreamObserver{})
	openSergoClient.subscribeConfigStreamStatus.Store(initial)
	return openSergoClient, nil
}

var (
	clientMu   sync.Mutex
	clientPool = make(map[string]*OpenSergoClient)
)

// GetOpenSergoClientByPool returns an instance of OpenSergoClient from a pool, based on host:port,
// if it doesn't exist, it will be created and reused for the next call.
func GetOpenSergoClientByPool(host string, port uint32) (*OpenSergoClient, error) {
	address := host + ":" + strconv.FormatUint(uint64(port), 10)

	clientMu.Lock()
	defer clientMu.Unlock()

	if client, ok := clientPool[address]; ok {
		return client, nil
	}

	client, err := NewOpenSergoClient(host, port)
	if err != nil {
		return nil, err
	}
	clientPool[address] = client

	return client, nil
}

func (c *OpenSergoClient) SubscribeDataCache() *subscribe.SubscribeDataCache {
	return c.subscribeDataCache
}

func (c *OpenSergoClient) SubscriberRegistry() *subscribe.SubscriberRegistry {
	return c.subscriberRegistry
}

// KeepAlive
//
// keepalive OpenSergoClient
func (c *OpenSergoClient) keepAlive() {
	// TODO change to event-driven-model instead of for-loop
	for {
		status := c.CurrentStreamStatus()
		if status == interrupted {
			logging.Info("Try to restart OpenSergoClient...")
			// We do not handle error here because error has print in Start()
			_ = c.Start()
		}
		time.Sleep(time.Duration(10) * time.Second)
	}
}

func (c *OpenSergoClient) CurrentStreamStatus() OpensergoClientStreamStatus {
	return c.subscribeConfigStreamStatus.Load().(OpensergoClientStreamStatus)
}

// Start the OpenSergoClient.
func (c *OpenSergoClient) Start() error {
	logging.Info("OpenSergoClient is starting...")

	// keepalive OpensergoClient
	if c.CurrentStreamStatus() == initial {
		logging.Info("Start a keepalive() daemon goroutine to keep OpenSergoClient alive")
		go c.keepAlive()
	}

	c.subscribeConfigStreamStatus.Store(starting)

	stream, err := c.transportServiceClient.SubscribeConfig(context.Background())
	if err != nil {
		logging.Error(err, "[OpenSergo SDK] SubscribeConfigStream can not connect.")
		c.subscribeConfigStreamStatus.Store(interrupted)
		return err
	}
	c.subscribeConfigStreamPtr.Store(&subscribeConfigStream{stream: stream})

	if c.subscribeConfigStreamObserverPtr.Load().(*subscribeConfigStreamObserver).observer == nil {
		c.subscribeConfigStreamObserverPtr.Store(&subscribeConfigStreamObserver{observer: NewSubscribeConfigStreamObserver(c)})
		c.subscribeConfigStreamObserverPtr.Load().(*subscribeConfigStreamObserver).observer.Start()
	}

	logging.Info("[OpenSergo SDK] begin to subscribe config-data...")
	// TODO: handle error inside the ForEach here.
	c.subscriberRegistry.ForEachSubscribeKey(func(key model.SubscribeKey) bool {
		err := c.SubscribeConfig(key)
		if err != nil {
			logging.Error(err, "Failed to SubscribeConfig for key", "namespace", key.Namespace(), "app", key.App(), "kind", key.Kind())
		}
		return true
	})

	logging.Info("OpenSergoClient has been started", "host", c.host, "port", c.port)
	c.subscribeConfigStreamStatus.Store(started)
	return nil
}

// SubscribeConfig sends a subscribe request to opensergo-control-plane,
// and return the result of subscribe config-data from opensergo-control-plane.
func (c *OpenSergoClient) SubscribeConfig(subscribeKey model.SubscribeKey, opts ...api.SubscribeOption) error {
	configStream := c.subscribeConfigStreamPtr.Load().(*subscribeConfigStream)
	if configStream == nil || configStream.stream == nil {
		logging.Warn("gRPC stream in OpenSergoClient is nil! Cannot subscribe config-data. waiting for keepalive goroutine to restart...", "namespace", subscribeKey.Namespace(), "app", subscribeKey.App(), "kinds", subscribeKey.Kind().GetName())
		c.subscribeConfigStreamStatus.Store(interrupted)
		return errors.New("configStream not ready (nil)")
	}

	options := &api.SubscribeOptions{}
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(options)
		}
	}

	// Register subscribers.
	if len(options.Subscribers) > 0 {
		for _, subscriber := range options.Subscribers {
			c.subscriberRegistry.RegisterSubscriber(subscribeKey, subscriber)
		}
	}

	subscribeRequestTarget := transportPb.SubscribeRequestTarget{
		Namespace: subscribeKey.Namespace(),
		App:       subscribeKey.App(),
		Kinds:     []string{subscribeKey.Kind().GetName()},
	}

	subscribeRequest := &transportPb.SubscribeRequest{
		Target: &subscribeRequestTarget,
		OpType: transportPb.SubscribeOpType_SUBSCRIBE,
	}

	err := c.subscribeConfigStreamPtr.Load().(*subscribeConfigStream).stream.Send(subscribeRequest)
	if err != nil {
		logging.Error(err, "Failed to send or handle SubscribeConfig request", "namespace", subscribeKey.Namespace(), "app", subscribeKey.App(), "kinds", subscribeKey.Kind().GetName())
		return err
	}

	logging.Info("Subscribe success", "namespace", subscribeKey.Namespace(), "app", subscribeKey.App(), "kinds", subscribeKey.Kind().GetName())

	return nil
}

// UnsubscribeConfig sends an un-subscribe request to opensergo-control-plane
// and remove all the subscribers by subscribeKey.
func (c *OpenSergoClient) UnsubscribeConfig(subscribeKey model.SubscribeKey) error {

	if c.subscribeConfigStreamPtr.Load().(*subscribeConfigStream).stream == nil {
		logging.Warn("gRPC stream in OpenSergoClient is nil! Cannot unsubscribe config-data. waiting for keepalive goroutine to restart...", "namespace", subscribeKey.Namespace(), "app", subscribeKey.App(), "kinds", subscribeKey.Kind().GetName())
		c.subscribeConfigStreamStatus.Store(interrupted)
		return errors.New("configStream not ready (nil)")
	}

	subscribeRequestTarget := transportPb.SubscribeRequestTarget{
		Namespace: subscribeKey.Namespace(),
		App:       subscribeKey.App(),
	}
	subscribeRequestTarget.Kinds = append(subscribeRequestTarget.Kinds, subscribeKey.Kind().GetName())

	subscribeRequest := &transportPb.SubscribeRequest{
		Target: &subscribeRequestTarget,
		OpType: transportPb.SubscribeOpType_UNSUBSCRIBE,
	}

	// Send SubscribeRequest (unsubscribe command)
	err := c.subscribeConfigStreamPtr.Load().(*subscribeConfigStream).stream.Send(subscribeRequest)
	if err != nil {
		logging.Error(err, "Failed to send or handle UnsubscribeConfig request", "namespace", subscribeKey.Namespace(), "app", subscribeKey.App(), "kinds", subscribeKey.Kind().GetName())
		return err
	}

	logging.Info("Unsubscribe success", "namespace", subscribeKey.Namespace(), "app", subscribeKey.App(), "kinds", subscribeKey.Kind().GetName())

	// Remove subscribers of the subscribe target.
	c.subscriberRegistry.RemoveSubscribers(subscribeKey)

	return nil
}
