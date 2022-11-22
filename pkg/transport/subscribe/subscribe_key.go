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

import "github.com/opensergo/opensergo-go/pkg/configkind"

// SubscribeKey is the unique identification for SubscribeData or Subscriber or for
type SubscribeKey struct {
	namespace string
	app       string
	kind      configkind.ConfigKind
}

// NewSubscribeKey returns an instance of SubscribeKey which is constructed by parameters.
func NewSubscribeKey(namespace string, app string, configKind configkind.ConfigKind) *SubscribeKey {
	return &SubscribeKey{
		namespace: namespace,
		app:       app,
		kind:      configKind,
	}
}

// Namespace returns the namespace of SubscribeKey
func (subscribeKey SubscribeKey) Namespace() string {
	return subscribeKey.namespace
}

// App returns the app of SubscribeKey
func (subscribeKey SubscribeKey) App() string {
	return subscribeKey.app
}

// Kind returns the kind of SubscribeKey
func (subscribeKey SubscribeKey) Kind() configkind.ConfigKind {
	return subscribeKey.kind
}
