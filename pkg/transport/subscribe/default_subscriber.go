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
	"encoding/json"
	"github.com/opensergo/opensergo-go/pkg/common/logging"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type DefaultSubscriber struct {
}

func (defaultSubscriber DefaultSubscriber) OnSubscribeDataUpdate(subscribeKey SubscribeKey, dataSlice []protoreflect.ProtoMessage) (bool, error) {
	// TODO  implement the custom-logic OnSubscribeDataUpdate
	jsonBytes, _ := json.Marshal(dataSlice)
	logging.Info("[OpenSergo SDK] receive data in DefaultSubscriber.", "data", string(jsonBytes))
	return true, nil
}
