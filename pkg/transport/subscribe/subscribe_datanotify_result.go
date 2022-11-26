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
	"github.com/opensergo/opensergo-go/pkg/transport"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// SubscribeDataNotifyResult is a struct for the result from opensergo-control-plane
type SubscribeDataNotifyResult struct {
	Code         int32
	DecodedData  []protoreflect.ProtoMessage
	NotifyErrors []error
}

func WithSuccessNotifyResult(decodedData []protoreflect.ProtoMessage) *SubscribeDataNotifyResult {
	return &SubscribeDataNotifyResult{
		Code:        transport.CodeSuccess,
		DecodedData: decodedData,
	}
}
