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
	transportServiceClient *transportPb.OpenSergoUniversalTransportServiceClient
	subscribeConfigStream  transportPb.OpenSergoUniversalTransportService_SubscribeConfigClient

	subscribeDataCache            *subscribe.SubscribeDataCache
	subscriberRegistry            *subscribe.SubscriberRegistry
	subscribeConfigStreamObserver *SubscribeConfigStreamObserver

	status    *atomic.Value
	requestId groupcache.AtomicInt
}

type OpensergoClientStatus uint8

const (
	initial OpensergoClientStatus = iota
	starting
	started
	interrupted
)

// NewOpenSergoClient returns an instance of OpenSergoClient, and init some properties.
func NewOpenSergoClient(host string, port int) *OpenSergoClient {
	address := host + ":" + strconv.Itoa(port)
	clientConn, _ := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	transportServiceClient := transportPb.NewOpenSergoUniversalTransportServiceClient(clientConn)

	status := new(atomic.Value)
	status.Store(initial)
	subscribeDataCache := new(subscribe.SubscribeDataCache)
	subscriberRegistry := new(subscribe.SubscriberRegistry)

	return &OpenSergoClient{
		transportServiceClient: &transportServiceClient,
		subscribeDataCache:     subscribeDataCache,
		subscriberRegistry:     subscriberRegistry,
		status:                 status,
		requestId:              groupcache.AtomicInt(0),
	}
}

func (openSergoClient OpenSergoClient) GetSubscribeDataCache() *subscribe.SubscribeDataCache {
	return openSergoClient.subscribeDataCache
}

func (openSergoClient OpenSergoClient) GetSubscriberRegistry() *subscribe.SubscriberRegistry {
	return openSergoClient.subscriberRegistry
}

// RegisterSubscribeInfo
//
// registry Subscriber
// if subscribeConfigClient in OpenSergoClient is initialized then subscribe config-data by subscribeKey
func (openSergoClient *OpenSergoClient) RegisterSubscribeInfo(subscribeInfo *SubscribeInfo) *OpenSergoClient {
	subscribers := openSergoClient.GetSubscriberRegistry().GetSubscribersOf(subscribeInfo.subscribeKey)
	for _, subscriber := range subscribeInfo.subscribers {
		openSergoClient.GetSubscriberRegistry().RegisterSubscriber(subscribeInfo.subscribeKey, subscriber)
	}

	if openSergoClient.status.Load().(OpensergoClientStatus) == started && len(subscribers) == 0 {
		openSergoClient.SubscribeConfig(subscribeInfo.subscribeKey)
	}
	return openSergoClient
}

// KeepAlive
//
// keepalive OpenSergoClient
func (openSergoClient *OpenSergoClient) keepAlive() {
	status := openSergoClient.status.Load().(OpensergoClientStatus)
	if status == initial || status == interrupted {
		logging.Info("[OpenSergo SDK] try to restart openSergoClient...")
		openSergoClient.Start()
	}
	time.Sleep(time.Duration(10) * time.Second)
	openSergoClient.keepAlive()
}

// Start
//
// to start the OpenSergoClient,
// and start the opensergoClientObserver for OpenSergoClient
func (openSergoClient *OpenSergoClient) Start() {

	logging.Info("[OpenSergo SDK] OpenSergoClient is starting...")

	// keepalive OpensergoClient
	if openSergoClient.status.Load().(OpensergoClientStatus) == initial {
		logging.Info("[OpenSergo SDK] open keepalive() goroutine to keep OpensergoClient alive")
		go openSergoClient.keepAlive()
	}

	openSergoClient.status.Store(starting)

	subscribeConfigStream, err := (*openSergoClient.transportServiceClient).SubscribeConfig(context.Background())
	if err != nil {
		logging.Error(err, "[OpenSergo SDK] SubscribeConfigStream can not connect.")
		openSergoClient.status.Store(interrupted)
		return
	}
	openSergoClient.subscribeConfigStream = subscribeConfigStream

	if openSergoClient.subscribeConfigStreamObserver == nil {
		openSergoClient.subscribeConfigStreamObserver = NewSubscribeConfigStreamObserver(openSergoClient)
	}
	openSergoClient.subscribeConfigStreamObserver.Start()

	logging.Info("[OpenSergo SDK] begin to subscribe config-data...")
	subscribersInRegistry := openSergoClient.GetSubscriberRegistry().GetSubscribersAll()
	subscribersInRegistry.Range(func(key, value interface{}) bool {
		subscribeKey := key.(subscribe.SubscribeKey)
		openSergoClient.SubscribeConfig(subscribeKey)
		return true
	})

	logging.Info("[OpenSergo SDK] OpenSergoClient is started")
	openSergoClient.status.Store(started)
}

// SubscribeConfig
//
// send a subscribe request to opensergo-control-plane,
// and return the result of subscribe config-data from opensergo-control-plane.
func (openSergoClient *OpenSergoClient) SubscribeConfig(subscribeKey subscribe.SubscribeKey) bool {
	if openSergoClient.subscribeConfigStream == nil {
		logging.Warn("[OpenSergo SDK] gRpc Stream (SubscribeConfigStream) in openSergoClient is nil! can't subscribe config-data. waiting for keepalive goroutine to restart...")
		openSergoClient.status.Store(interrupted)
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

	err := openSergoClient.subscribeConfigStream.Send(subscribeRequest)
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

	if openSergoClient.subscribeConfigStream == nil {
		logging.Warn("[OpenSergo SDK] gRpc Stream (SubscribeConfigStream) in openSergoClient is nil! can't unsubscribe config-data. waiting for keepalive goroutine to restart...")
		openSergoClient.status.Store(interrupted)
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
	err := openSergoClient.subscribeConfigStream.Send(subscribeRequest)
	if err != nil {
		logging.Error(err, "[OpenSergo SDK] something occured when sending unsubscribe request.")
		return false
	}

	// Remove subscribers of the subscribe target.
	openSergoClient.GetSubscriberRegistry().RemoveSubscribers(subscribeKey)

	return true
}
