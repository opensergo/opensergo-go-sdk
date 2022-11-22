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
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// SubscribeDataCache is a local-cache which cached all the config-data subscribed from opensergo-control-plane.
type SubscribeDataCache struct {
	// map[SubscribeKey] *SubscribedData
	cache sync.Map
}

// UpdateSubscribedData to update the config-data by subscribeKey
func (subscribeDataCache *SubscribeDataCache) UpdateSubscribedData(subscribeKey SubscribeKey, data []protoreflect.ProtoMessage, version int64) {
	// TODO: guarantee the latest version
	subscribeDataCache.cache.Store(subscribeKey, NewSubscribedData(version, data))
}

// GetSubscribedData to get the config-data from cache by subscribeKey
func (subscribeDataCache *SubscribeDataCache) GetSubscribedData(subscribeKey SubscribeKey) *SubscribedData {
	load, _ := subscribeDataCache.cache.Load(subscribeKey)
	if load == nil {
		return nil
	}
	return load.(*SubscribedData)
}
