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

package main

import (
	"github.com/opensergo/opensergo-go/pkg/client"
	"github.com/opensergo/opensergo-go/pkg/configkind"
	"github.com/opensergo/opensergo-go/pkg/transport/subscribe"
	"github.com/opensergo/opensergo-go/samples"
	"time"
)

// main
//
// a simple example.
func main() {
	// add console logger (optional)
	//logging.NewConsoleLogger(logging.InfoLevel, logging.SeparateFormat, true)
	// add file logger (optional)
	//logging.NewFileLogger("/Users/J/logs/opensergo/opensergo-universal-transport-service.log", logging.InfoLevel, logging.JsonFormat, true)

	// instant OpenSergoClient
	openSergoClient := client.NewOpenSergoClient("127.0.0.1", 10246)

	// registry SubscribeInfo of FaultToleranceRule
	// 1. instant SubscribeKey
	faultToleranceSubscribeKey := subscribe.NewSubscribeKey("default", "foo-app", configkind.ConfigKindRefFaultToleranceRule{})
	// 2. instant Subscriber
	sampleFaultToleranceRuleSubscriber := new(samples.SampleFaultToleranceRuleSubscriber)
	// 3. construct SubscribeInfo
	faultToleranceSubscribeInfo := client.NewSubscribeInfo(faultToleranceSubscribeKey)
	faultToleranceSubscribeInfo.AppendSubscriber(sampleFaultToleranceRuleSubscriber)
	// 4. registry
	openSergoClient.RegisterSubscribeInfo(faultToleranceSubscribeInfo)

	// registry SubscribeInfo of RateLimitStrategy
	rateLimitSubscribeKey := subscribe.NewSubscribeKey("default", "foo-app", configkind.ConfigKindRefRateLimitStrategy{})
	sampleRateLimitStrategySubscriber := new(samples.SampleRateLimitStrategySubscriber)
	rateLimitSubscribeInfo := client.NewSubscribeInfo(rateLimitSubscribeKey)
	rateLimitSubscribeInfo.AppendSubscriber(sampleRateLimitStrategySubscriber)
	openSergoClient.RegisterSubscribeInfo(rateLimitSubscribeInfo)

	// start OpensergoClient
	openSergoClient.Start()

	//registry after OpenSergoClient started
	//faultToleranceSubscribeInfo.AppendSubscriber(new(subscribe.DefaultSubscriber))
	//openSergoClient.RegisterSubscribeInfo(faultToleranceSubscribeInfo)

	// registry after OpenSergoClient started
	//rateLimitSubscribeInfo.AppendSubscriber(new(subscribe.DefaultSubscriber))
	//openSergoClient.RegisterSubscribeInfo(rateLimitSubscribeInfo)

	// unsubscribeConfig
	go unsubscribeConfig(openSergoClient, *rateLimitSubscribeKey)

	select {}
}

func unsubscribeConfig(openSergoClient *client.OpenSergoClient, key subscribe.SubscribeKey) {

	for {
		time.Sleep(time.Duration(60) * time.Second)
		openSergoClient.UnsubscribeConfig(key)
	}

}
