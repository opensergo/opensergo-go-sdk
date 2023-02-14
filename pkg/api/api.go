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

package api

import (
	"github.com/opensergo/opensergo-go/pkg/model"
	"github.com/opensergo/opensergo-go/pkg/transport/subscribe"
)

type ClientOptions struct {
	connectRetryTimes uint
}

type ClientOption func(*ClientOptions)

func (clientOptions *ClientOptions) ApplyClientOptions(opts ...ClientOption) {
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(clientOptions)
		}
	}
}

func NewDefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		connectRetryTimes: 3,
	}
}

func (opts *ClientOptions) ConnectRetryTimes() uint {
	return opts.connectRetryTimes
}

func WithConnectRetryTimes(connectRetryTimes uint) ClientOption {
	return func(opts *ClientOptions) {
		opts.connectRetryTimes = connectRetryTimes
	}
}

// SubscribeOptions represents the options of OpenSergo data subscription.
type SubscribeOptions struct {
	subscribers []subscribe.Subscriber
	attachments map[string]interface{}
}

func (opts *SubscribeOptions) Subscribers() []subscribe.Subscriber {
	return opts.subscribers
}

func (opts *SubscribeOptions) Attachments() map[string]interface{} {
	return opts.attachments
}

type SubscribeOption func(*SubscribeOptions)

// WithSubscriber provides a subscriber.
func WithSubscriber(subscriber subscribe.Subscriber) SubscribeOption {
	return func(opts *SubscribeOptions) {
		if opts.subscribers == nil {
			opts.subscribers = make([]subscribe.Subscriber, 0)
		}
		opts.subscribers = append(opts.subscribers, subscriber)
	}
}

// WithAttachment provides an attachment (key-value pair).
func WithAttachment(key string, value interface{}) SubscribeOption {
	return func(opts *SubscribeOptions) {
		if opts.attachments == nil {
			opts.attachments = make(map[string]interface{})
		}
		opts.attachments[key] = value
	}
}

// OpenSergoClient is the universal interface of OpenSergo client.
type OpenSergoClient interface {
	// Start the client.
	Start() error
	// Close the client.
	Close() error
	// SubscribeConfig subscribes data for given subscribe target.
	SubscribeConfig(key model.SubscribeKey, opts ...SubscribeOption) error
	// UnsubscribeConfig unsubscribes data for given subscribe target.
	UnsubscribeConfig(subscribeKey model.SubscribeKey) error
}
