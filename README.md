<img src="https://user-images.githubusercontent.com/9434884/197435179-bbda0a82-6bae-485e-ac1a-490fee91a002.png" alt="OpenSergo Logo" width="50%">

# OpenSergo Go SDK

[![OpenSergo Go SDK CI](https://github.com/opensergo/opensergo-go-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/opensergo/opensergo-go-sdk/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

## Introduction

## Documentation

See the [OpenSergo Website](https://opensergo.io/) for the official website of OpenSergo.

See the [中文文档](https://opensergo.io/zh-cn/) for document in Chinese.

## Quick Start

``` go
func StartAndSubscribeOpenSergoConfig() error {
	// Set OpenSergo console logger (optional)
	logging.NewConsoleLogger(logging.InfoLevel, logging.SeparateFormat, true)
	// Set OpenSergo file logger (optional)
	// logging.NewFileLogger("./opensergo-universal-transport-service.log", logging.InfoLevel, logging.JsonFormat, true)

	// Create an OpenSergoClient.
	openSergoClient, err := client.NewOpenSergoClient("127.0.0.1", 10246)
	if err != nil {
		return err
	}

	// Start the OpenSergoClient.
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
	// Subscribe data with the key, but without a subscriber.
	// We may pull the data from the data cache later.
	err = openSergoClient.SubscribeConfig(*rateLimitSubscribeKey))

	return err
}
```

## Samples

For more samples, please refer to [here](./samples).

## Components

### logger mechanism

In `OpenSergo Go SDK`, we provide a fundamental logger mechanism, which has the following features:

- provider a universal Logger interface. Users may register customized logger adapters.
- provider a default Logger implementation, which can be used directory.
- provider a default Logger format assembler, which can assemble two formatters of log message.

Detail info refers to [pkg/common/logging](./pkg/common/logging)