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

package transport

const (
	CODE_SUCCESS                       int32 = 1
	CODE_ERROR_UNKNOWN                 int32 = 4000
	CODE_ERROR_SUBSCRIBE_HANDLER_ERROR int32 = 4007
	CODE_ERROR_VERSION_OUTDATED        int32 = 4010

	CODE_ERROR_UPPER_BOUND int32 = 4100

	FLAG_ACK  string = "ACK"
	FLAG_NACK string = "NACK"
)
