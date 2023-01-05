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
	faulttolerancePb "github.com/opensergo/opensergo-go/pkg/proto/fault_tolerance/v1"
)

// ConfigKindRefConcurrencyLimitStrategy  type of ConfigKind
type ConfigKindRefConcurrencyLimitStrategy struct {
}

// GetName returns the name of ConfigKindRefConcurrencyLimitStrategy
func (concurrencyLimitStrategy ConfigKindRefConcurrencyLimitStrategy) GetName() string {
	return "fault-tolerance.opensergo.io/v1alpha1/ConcurrencyLimitStrategy"
}

// GetSimpleName returns the crd name of ConfigKindRefConcurrencyLimitStrategy
func (concurrencyLimitStrategy ConfigKindRefConcurrencyLimitStrategy) GetSimpleName() string {
	return "ConcurrencyLimitStrategy"
}

// registry ConfigKindRefConcurrencyLimitStrategy to ConfigKindMetadataRegistry
func init() {
	GetConfigKindMetadataRegistry().RegisterConfigKind(ConfigKindRefConcurrencyLimitStrategy{}, new(faulttolerancePb.ConcurrencyLimitStrategy).ProtoReflect().Type())
}
