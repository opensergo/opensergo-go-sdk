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
	"github.com/opensergo/opensergo-go/pkg/common/logging"

	"github.com/opensergo/opensergo-go/pkg/api"
	"github.com/opensergo/opensergo-go/pkg/configkind"
	"github.com/opensergo/opensergo-go/pkg/model"
	"github.com/opensergo/opensergo-go/samples"
)

// main
//
// a simple example.
func main() {
	err := StartAndSubscribeOpenSergoConfig()
	if err != nil {
		// Handle error here.
		logging.Error(err, "Failed to StartAndSubscribeOpenSergoConfig: %s\n")
	}

	select {}
}

func StartAndSubscribeOpenSergoConfig() error {
	// Set OpenSergo console logger (optional)
	consoleLogger := logging.NewConsoleLogger(logging.InfoLevel, logging.JsonFormat, true)
	logging.AddLogger(consoleLogger)
	// Set OpenSergo file logger (optional)
	// fileLogger := logging.NewFileLogger("./opensergo-universal-transport-service.log", logging.InfoLevel, logging.JsonFormat, true)
	//logging.AddLogger(fileLogger)

	// Create a OpenSergoClient.
	openSergoClient, err := client.NewOpenSergoClient("127.0.0.1", 10246)
	if err != nil {
		return err
	}

	// Start OpenSergoClient
	err = openSergoClient.Start()
	if err != nil {
		return err
	}

	// Create a SubscribeKey for FaultToleranceRule.
	faultToleranceSubscribeKey := model.NewSubscribeKey("default", "foo-app", configkind.ConfigKindRefFaultToleranceRule{})
	// Create a Subscriber.
	sampleFaultToleranceRuleSubscriber := &samples.SampleFaultToleranceRuleSubscriber{}
	// Subscribe data with the key and subscriber.
	err = openSergoClient.SubscribeConfig(*faultToleranceSubscribeKey, api.WithSubscriber(sampleFaultToleranceRuleSubscriber))
	if err != nil {
		return err
	}

	// Create a SubscribeKey for RateLimitStrategy.
	rateLimitSubscribeKey := model.NewSubscribeKey("default", "foo-app", configkind.ConfigKindRefRateLimitStrategy{})
	// Create another Subscriber.
	sampleRateLimitStrategySubscriber := &samples.SampleRateLimitStrategySubscriber{}
	// Subscribe data with the key and subscriber.
	err = openSergoClient.SubscribeConfig(*rateLimitSubscribeKey, api.WithSubscriber(sampleRateLimitStrategySubscriber))

	if err != nil {
		return err
	}
	// Create a SubscribeKey for TrafficRouter
	trafficRouterSubscribeKey := model.NewSubscribeKey("default", "service-provider", configkind.ConfigKindTrafficRouterStrategy{})
	// Create another Subscriber
	sampleTrafficRouterSubscriber := &samples.SampleTrafficRouterSubscriber{}
	// Subscribe data with the key and subscriber
	err = openSergoClient.SubscribeConfig(*trafficRouterSubscribeKey, api.WithSubscriber(sampleTrafficRouterSubscriber))

	return err
}
