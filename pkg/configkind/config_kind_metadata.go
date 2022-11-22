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

package configkind

import "google.golang.org/protobuf/reflect/protoreflect"

// ConfigKindMetadata defined the match-relation between ConfigKind and proto-Entity
type ConfigKindMetadata struct {
	kind              ConfigKind
	kindPbMessageType protoreflect.MessageType
}

// NewConfigKindMetadata returns an instance of ConfigKindMetadata with construct parameters.
func NewConfigKindMetadata(kind ConfigKind, kindPbMessageType protoreflect.MessageType) *ConfigKindMetadata {

	return &ConfigKindMetadata{
		kind:              kind,
		kindPbMessageType: kindPbMessageType,
	}
}

// GetConfigKind returns the ConfigKind of ConfigKindMetadata
func (configKindMetadata *ConfigKindMetadata) GetConfigKind() ConfigKind {
	return configKindMetadata.kind
}

// GetKindPbMessageType returns the kindPbFullName of ConfigKindMetadata
func (configKindMetadata *ConfigKindMetadata) GetKindPbMessageType() protoreflect.MessageType {
	return configKindMetadata.kindPbMessageType
}
