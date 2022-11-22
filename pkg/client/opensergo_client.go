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
	"github.com/golang/groupcache"
	"github.com/opensergo/opensergo-go/pkg/common/logging"
	transportPb "github.com/opensergo/opensergo-go/pkg/proto/transport/v1"
	"github.com/opensergo/opensergo-go/pkg/transport/subscribe"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"strconv"
	"sync/atomic"
	"time"
)

// OpenSergoClient is the client to communication with opensergo-control-plane.
type OpenSergoClient struct {
	transportServiceClient transportPb.OpenSergoUniversalTransportServiceClient

	subscribeConfigStreamPtr         atomic.Value // type of value is *client.subscribeConfigStream
	subscribeConfigStreamObserverPtr atomic.Value // type of value is *client.SubscribeConfigStreamObserver
	subscribeConfigStreamStatus      atomic.Value // type of value is client.OpensergoClientStreamStatus

	subscribeDataCache *subscribe.SubscribeDataCache
	subscriberRegistry *subscribe.SubscriberRegistry

	requestId groupcache.AtomicInt
}

// NewOpenSergoClient returns an instance of OpenSergoClient, and init some properties.
func NewOpenSergoClient(host string, port int) *OpenSergoClient {
	address := host + ":" + strconv.Itoa(port)
	clientConn, _ := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	transportServiceClient := transportPb.NewOpenSergoUniversalTransportServiceClient(clientConn)

	openSergoClient := &OpenSergoClient{
		transportServiceClient: transportServiceClient,
		subscribeDataCache:     &subscribe.SubscribeDataCache{},
		subscriberRegistry:     &subscribe.SubscriberRegistry{},
		requestId:              groupcache.AtomicInt(0),
	}
	openSergoClient.subscribeConfigStreamPtr.Store(&subscribeConfigStream{})
	openSergoClient.subscribeConfigStreamObserverPtr.Store(&subscribeConfigStreamObserver{})
	openSergoClient.subscribeConfigStreamStatus.Store(initial)
	return openSergoClient
}

func (openSergoClient *OpenSergoClient) SubscribeDataCache() *subscribe.SubscribeDataCache {
	return openSergoClient.subscribeDataCache
}

func (openSergoClient *OpenSergoClient) SubscriberRegistry() *subscribe.SubscriberRegistry {
	return openSergoClient.subscriberRegistry
}

// RegisterSubscribeInfo
//
// registry Subscriber
// if subscribeConfigClient in OpenSergoClient is initialized then subscribe config-data by subscribeKey
func (openSergoClient *OpenSergoClient) RegisterSubscribeInfo(subscribeInfo *SubscribeInfo) *OpenSergoClient {
	subscribers := openSergoClient.subscriberRegistry.GetSubscribersOf(subscribeInfo.subscribeKey)
	for _, subscriber := range subscribeInfo.subscribers {
		openSergoClient.subscriberRegistry.RegisterSubscriber(subscribeInfo.subscribeKey, subscriber)
	}

	if openSergoClient.subscribeConfigStreamStatus.Load().(OpensergoClientStreamStatus) == started && len(subscribers) == 0 {
		openSergoClient.SubscribeConfig(subscribeInfo.subscribeKey)
	}
	return openSergoClient
}

// KeepAlive
//
// keepalive OpenSergoClient
func (openSergoClient *OpenSergoClient) keepAlive() {
	// TODO change to event-driven-model instead of for-loop
	for {
		status := openSergoClient.subscribeConfigStreamStatus.Load().(OpensergoClientStreamStatus)
		if status == interrupted {
			logging.Info("[OpenSergo SDK] try to restart openSergoClient...")
			if err := openSergoClient.Start(); err != nil {
				// nothing to do because error has print in Start()
			}
		}
		time.Sleep(time.Duration(10) * time.Second)
	}
}

// Start
//
// to start the OpenSergoClient,
// and start the opensergoClientObserver for OpenSergoClient
func (openSergoClient *OpenSergoClient) Start() error {

	logging.Info("[OpenSergo SDK] OpenSergoClient is starting...")

	// keepalive OpensergoClient
	if openSergoClient.subscribeConfigStreamStatus.Load().(OpensergoClientStreamStatus) == initial {
		logging.Info("[OpenSergo SDK] open keepalive() goroutine to keep OpensergoClient alive")
		go openSergoClient.keepAlive()
	}

	openSergoClient.subscribeConfigStreamStatus.Store(starting)

	stream, err := openSergoClient.transportServiceClient.SubscribeConfig(context.Background())
	if err != nil {
		logging.Error(err, "[OpenSergo SDK] SubscribeConfigStream can not connect.")
		openSergoClient.subscribeConfigStreamStatus.Store(interrupted)
		return err
	}
	openSergoClient.subscribeConfigStreamPtr.Store(&subscribeConfigStream{stream: stream})

	if openSergoClient.subscribeConfigStreamObserverPtr.Load().(*subscribeConfigStreamObserver).observer == nil {
		openSergoClient.subscribeConfigStreamObserverPtr.Store(&subscribeConfigStreamObserver{observer: NewSubscribeConfigStreamObserver(openSergoClient)})
		openSergoClient.subscribeConfigStreamObserverPtr.Load().(*subscribeConfigStreamObserver).observer.Start()
	}

	logging.Info("[OpenSergo SDK] begin to subscribe config-data...")
	openSergoClient.subscriberRegistry.RunWithRangeRegistry(func(key, value interface{}) bool {
		subscribeKey := key.(subscribe.SubscribeKey)
		openSergoClient.SubscribeConfig(subscribeKey)
		return true
	})

	logging.Info("[OpenSergo SDK] OpenSergoClient is started")
	openSergoClient.subscribeConfigStreamStatus.Store(started)
	return nil
}

// SubscribeConfig
//
// send a subscribe request to opensergo-control-plane,
// and return the result of subscribe config-data from opensergo-control-plane.
func (openSergoClient *OpenSergoClient) SubscribeConfig(subscribeKey subscribe.SubscribeKey) bool {
	if openSergoClient.subscribeConfigStreamPtr.Load().(*subscribeConfigStream).stream == nil {
		logging.Warn("[OpenSergo SDK] gRpc Stream (SubscribeConfigStream) in openSergoClient is nil! can't subscribe config-data. waiting for keepalive goroutine to restart...")
		openSergoClient.subscribeConfigStreamStatus.Store(interrupted)
		return false
	}

	subscribeRequestTarget := transportPb.SubscribeRequestTarget{
		Namespace: subscribeKey.GetNamespace(),
		App:       subscribeKey.GetApp(),
		Kinds:     []string{subscribeKey.GetKind().GetName()},
	}

	subscribeRequest := &transportPb.SubscribeRequest{
		Target: &subscribeRequestTarget,
		OpType: transportPb.SubscribeOpType_SUBSCRIBE,
	}

	err := openSergoClient.subscribeConfigStreamPtr.Load().(*subscribeConfigStream).stream.Send(subscribeRequest)
	if err != nil {
		logging.Error(err, "[OpenSergo SDK] something occured when sending subscribe request.")
		return false
	}

	return true
}

// UnsubscribeConfig
// send an un-subscribe request to opensergo-control-plane
// and remove all the subscribers by subscribeKey
func (openSergoClient *OpenSergoClient) UnsubscribeConfig(subscribeKey subscribe.SubscribeKey) bool {

	if openSergoClient.subscribeConfigStreamPtr.Load().(*subscribeConfigStream).stream == nil {
		logging.Warn("[OpenSergo SDK] gRpc Stream (SubscribeConfigStream) in openSergoClient is nil! can't unsubscribe config-data. waiting for keepalive goroutine to restart...")
		openSergoClient.subscribeConfigStreamStatus.Store(interrupted)
		return false
	}

	subscribeRequestTarget := transportPb.SubscribeRequestTarget{
		Namespace: subscribeKey.GetNamespace(),
		App:       subscribeKey.GetApp(),
	}
	subscribeRequestTarget.Kinds = append(subscribeRequestTarget.Kinds, subscribeKey.GetKind().GetName())

	subscribeRequest := &transportPb.SubscribeRequest{
		Target: &subscribeRequestTarget,
		OpType: transportPb.SubscribeOpType_UNSUBSCRIBE,
	}

	// Send SubscribeRequest (unsubscribe command)
	err := openSergoClient.subscribeConfigStreamPtr.Load().(*subscribeConfigStream).stream.Send(subscribeRequest)
	if err != nil {
		logging.Error(err, "[OpenSergo SDK] something occured when sending unsubscribe request.")
		return false
	}

	// Remove subscribers of the subscribe target.
	openSergoClient.subscriberRegistry.RemoveSubscribers(subscribeKey)

	return true
}
