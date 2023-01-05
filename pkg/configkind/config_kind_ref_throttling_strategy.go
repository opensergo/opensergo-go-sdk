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

// ConfigKindRefThrottlingStrategy  type of ConfigKind
type ConfigKindRefThrottlingStrategy struct {
}

// GetName returns the name of ConfigKindRefThrottlingStrategy
func (throttlingStrategy ConfigKindRefThrottlingStrategy) GetName() string {
	return "fault-tolerance.opensergo.io/v1alpha1/ThrottlingStrategy"
}

// GetSimpleName returns the crd name of ConfigKindRefThrottlingStrategy
func (throttlingStrategy ConfigKindRefThrottlingStrategy) GetSimpleName() string {
	return "ThrottlingStrategy"
}

// registry ConfigKindRefThrottlingStrategy to ConfigKindMetadataRegistry
func init() {
	GetConfigKindMetadataRegistry().RegisterConfigKind(ConfigKindRefThrottlingStrategy{}, new(faulttolerancePb.ThrottlingStrategy).ProtoReflect().Type())
}
