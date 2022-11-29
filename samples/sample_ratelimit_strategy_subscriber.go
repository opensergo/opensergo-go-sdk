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

package samples

import (
	"encoding/json"

	"github.com/opensergo/opensergo-go/pkg/common/logging"
	"github.com/opensergo/opensergo-go/pkg/transport/subscribe"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type SampleRateLimitStrategySubscriber struct {
}

func (sampleRateLimitStrategySubscriber SampleRateLimitStrategySubscriber) OnSubscribeDataUpdate(subscribeKey subscribe.SubscribeKey, data interface{}) (bool, error) {
	// TODO  implement the custom-logic OnSubscribeDataUpdate
	messages := data.([]protoreflect.ProtoMessage)
	jsonBytes, _ := json.Marshal(messages)
	logging.Info(string(jsonBytes))
	return true, nil
}
