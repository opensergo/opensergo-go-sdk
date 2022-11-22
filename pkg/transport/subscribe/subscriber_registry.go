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
)

// SubscriberRegistry is a local cache, witch stores all the Subscriber that we .
type SubscriberRegistry struct {
	// map[SubscribeKey] []*Subscriber
	registry sync.Map
}

// RegisterSubscriber registry a Subscriber to local cache
func (subscriberRegistry *SubscriberRegistry) RegisterSubscriber(subscribeKey SubscribeKey, subscriber Subscriber) {
	subscribersInRegistry, _ := subscriberRegistry.registry.Load(subscribeKey)
	if subscribersInRegistry == nil {
		subscriberRegistry.registry.Store(subscribeKey, []Subscriber{subscriber})
		return
	}

	for _, subscriberItem := range subscribersInRegistry.([]Subscriber) {
		if reflect.ValueOf(subscriberItem).Interface() != reflect.ValueOf(subscriber).Interface() {
			subscribers := append(subscribersInRegistry.([]Subscriber), subscriber)
			subscriberRegistry.registry.Store(subscribeKey, subscribers)
		}
	}
}

// RunWithRangeRegistry run func with range SubscriberRegistry, avoid copy sync.Map
func (subscriberRegistry *SubscriberRegistry) RunWithRangeRegistry(runner func(key, value interface{}) bool) {
	subscriberRegistry.registry.Range(func(key, value interface{}) bool {
		return runner(key, value)
	})
}

// GetSubscribersOf returns a Subscriber by SubscribeKey
func (subscriberRegistry *SubscriberRegistry) GetSubscribersOf(subscribeKey SubscribeKey) []Subscriber {
	value, _ := subscriberRegistry.registry.Load(subscribeKey)
	if value == nil {
		return nil
	}
	subscribers := value.([]Subscriber)
	return subscribers
}

// RemoveSubscribers remove Subscribers by SubscribeKey
func (subscriberRegistry *SubscriberRegistry) RemoveSubscribers(subscribeKey SubscribeKey) {
	subscriberRegistry.registry.Delete(subscribeKey)
}
