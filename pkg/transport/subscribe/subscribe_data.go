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

// SubscribedData is the config-data that subscribed from opensergo-control-plane.
type SubscribedData struct {
	version int64
	data    interface{}
}

// NewSubscribedData returns a instance of SubscribedData by construct parameters.
func NewSubscribedData(version int64, data interface{}) *SubscribedData {
	return &SubscribedData{
		version: version,
		data:    data,
	}
}

func (subscribedData SubscribedData) GetVersion() int64 {
	return subscribedData.version
}

func (subscribedData SubscribedData) GetData() interface{} {
	return subscribedData.data
}
