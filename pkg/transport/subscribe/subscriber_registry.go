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

package subscribe

import (
	"reflect"
	"sync"

	"github.com/opensergo/opensergo-go/pkg/model"
)

// SubscriberRegistry is a local cache, which stores all the Subscriber.
type SubscriberRegistry struct {
	// map[SubscribeKey] []subscribe.Subscriber
	subscriberMap sync.Map
}

// RegisterSubscriber subscriberMap a Subscriber to local cache
func (subscriberRegistry *SubscriberRegistry) RegisterSubscriber(subscribeKey model.SubscribeKey, subscriber Subscriber) {
	if subscriber == nil {
		return
	}
	subscribersInRegistry, _ := subscriberRegistry.subscriberMap.Load(subscribeKey)
	if subscribersInRegistry == nil {
		subscriberRegistry.subscriberMap.Store(subscribeKey, []Subscriber{subscriber})
		return
	}

	for _, subscriberItem := range subscribersInRegistry.([]Subscriber) {
		if reflect.ValueOf(subscriberItem).Interface() != reflect.ValueOf(subscriber).Interface() {
			subscribers := append(subscribersInRegistry.([]Subscriber), subscriber)
			subscriberRegistry.subscriberMap.Store(subscribeKey, subscribers)
		}
	}
}

// ForEachSubscribeKey run func with range SubscriberRegistry.
func (subscriberRegistry *SubscriberRegistry) ForEachSubscribeKey(runner func(key model.SubscribeKey) bool) {
	subscriberRegistry.subscriberMap.Range(func(key, value interface{}) bool {
		return runner(key.(model.SubscribeKey))
	})
}

// GetSubscribersOf returns a Subscriber by SubscribeKey
func (subscriberRegistry *SubscriberRegistry) GetSubscribersOf(subscribeKey model.SubscribeKey) []Subscriber {
	value, _ := subscriberRegistry.subscriberMap.Load(subscribeKey)
	if value == nil {
		return nil
	}
	subscribers := value.([]Subscriber)
	return subscribers
}

// RemoveSubscribers remove Subscribers by SubscribeKey
func (subscriberRegistry *SubscriberRegistry) RemoveSubscribers(subscribeKey model.SubscribeKey) {
	subscriberRegistry.subscriberMap.Delete(subscribeKey)
}
