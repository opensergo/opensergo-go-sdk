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

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"sync"
)

// ConfigKindMetadataRegistry is a local cache, witch stores all the ConfigKindMetadata supported by OpenSergo.
type ConfigKindMetadataRegistry struct {

	// TODO consider for ConfigKind maybe has several Properties,
	// TODO should change the struct 'map[kindName] ConfigKindMetadata' to 'map[kindName] []ConfigKindMetadata',
	// TODO for example, ConfigKind maybe has version, so the same kindName would have several ConfigKindMetadata
	// map[kindName] ConfigKindMetadata
	registry sync.Map
}

// configKindMetadataRegistry singleton instance
var configKindMetadataRegistry *ConfigKindMetadataRegistry

var configKindMetadataRegistryOnce sync.Once

// GetConfigKindMetadataRegistry returns a singleton instance of ConfigKindMetadataRegistry.
func GetConfigKindMetadataRegistry() *ConfigKindMetadataRegistry {
	configKindMetadataRegistryOnce.Do(func() {
		configKindMetadataRegistry = new(ConfigKindMetadataRegistry)
	})
	return configKindMetadataRegistry
}

// GetKindMetadataByInstance returns a ConfigKindMetadata by ConfigKind
func (configKindMetadataRegistry ConfigKindMetadataRegistry) GetKindMetadataByInstance(kind ConfigKind) ConfigKindMetadata {
	if kind == nil {
		return ConfigKindMetadata{}
	}
	if value, ok := configKindMetadataRegistry.registry.Load(kind.GetName()); ok {
		return value.(ConfigKindMetadata)
	}
	return ConfigKindMetadata{}
}

// GetKindMetadataByName returns a ConfigKindMetadata by kindName of ConfigKind
func (configKindMetadataRegistry ConfigKindMetadataRegistry) GetKindMetadataByName(kindName string) ConfigKindMetadata {
	if kindName == "" {
		return ConfigKindMetadata{}
	}
	if value, ok := configKindMetadataRegistry.registry.Load(kindName); ok {
		return value.(ConfigKindMetadata)
	}
	return ConfigKindMetadata{}
}

// RegisterConfigKind caches a KindMetadata
func (configKindMetadataRegistry *ConfigKindMetadataRegistry) RegisterConfigKind(kind ConfigKind, pbMessageType protoreflect.MessageType) {
	configKindMetadataRegistry.registry.Store(kind.GetName(), *NewConfigKindMetadata(kind, pbMessageType))
}
