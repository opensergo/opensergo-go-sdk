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
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type ConfigKindTrafficRouterStrategy struct {
}

// GetName returns the name of ConfigKindRefTrafficRouterStrategy
func (trafficRouterStrategy ConfigKindTrafficRouterStrategy) GetName() string {
	return "traffic.opensergo.io/v1alpha1/TrafficRouter"
}

// GetSimpleName returns the crd name of ConfigKindRefTrafficRouterStrategy
func (trafficRouterStrategy ConfigKindTrafficRouterStrategy) GetSimpleName() string {
	return "TrafficRouter"
}

// registry ConfigKindRefTrafficRouterStrategy to ConfigKindMetadataRegistry
func init() {
	GetConfigKindMetadataRegistry().RegisterConfigKind(ConfigKindTrafficRouterStrategy{}, new(routev3.RouteConfiguration).ProtoReflect().Type())
}
