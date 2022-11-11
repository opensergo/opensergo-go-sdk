# OpenSergo Go SDK


## How to use

### scene 1 : subscribe config-data

``` go
package main

func main() {
    // add console logger (optional)
    //logging.NewConsoleLogger(logging.InfoLevel, logging.SeparateFormat)
    // add file logger (optional)
    //logging.NewFileLogger("/logs/opensergo/opensergo-universal-transport-service.log", logging.InfoLevel, logging.JsonFormat)
    
    // instant OpenSergoClient
    openSergoClient := client.NewOpenSergoClient("33.1.33.1",10246)
    
    // register SubscribeInfo of FaultToleranceRule
    // 1. instant SubscribeKey
    faultToleranceSubscribeKey := subscribe.NewSubscribeKey("default", "foo-app", configkind.ConfigKindRefFaultToleranceRule{})
    // 2. construct SubscribeInfo
    faultToleranceSubscribeInfo := client.NewSubscribeInfo(faultToleranceSubscribeKey)
    // 3. do register
    openSergoClient.RegisterSubscribeInfo(faultToleranceSubscribeInfo)
    
    // start OpensergoClient
    openSergoClient.Start()
    
    // register after OpenSergoClient started
    // register SubscribeInfo of RateLimitStrategy
    rateLimitSubscribeKey := subscribe.NewSubscribeKey("default", "foo-app", configkind.ConfigKindRefRateLimitStrategy{})
    rateLimitSubscribeInfo := client.NewSubscribeInfo(rateLimitSubscribeKey)
    openSergoClient.RegisterSubscribeInfo(rateLimitSubscribeInfo)

    select {}
}
```

### scene 2 : subscribe config-data with custom-logic when config-data changed.
Add a subscriber by implementing the function in `subscribe.Subscriber`.  
There are some samples in `sample` directory : `sample/main/sample_faulttolerance_rule_subscriber.go` and `sample_ratelimit_strategy_subscriber.go`

``` go
    type SampleFaultToleranceRuleSubscriber struct {
    }
    
    func (sampleFaultToleranceRuleSubscriber SampleFaultToleranceRuleSubscriber) OnSubscribeDataUpdate(subscribeKey subscribe.SubscribeKey, dataSlice []protoreflect.ProtoMessage) bool {
        // TODO add custom-logic when config-data change
        // ......
        return true
    }
```

And then register it into `subscriber.SubscriberRegistry`.

``` go
package main

func main() {
    // add console logger (optional)
    //logging.NewConsoleLogger(logging.InfoLevel, logging.SeparateFormat)
    // add file logger (optional)
    //logging.NewFileLogger("/logs/opensergo/opensergo-universal-transport-service.log", "fileLogger", logging.InfoLevel, logging.JsonFormat)
    
    // instant OpenSergoClient
    openSergoClient := client.NewOpenSergoClient("33.1.33.1",10246)
    
    // register SubscribeInfo of FaultToleranceRule
    // 1. instant SubscribeKey
    faultToleranceSubscribeKey := subscribe.NewSubscribeKey("default", "foo-app", configkind.ConfigKindRefFaultToleranceRule{})
    // 2. instant Subscriber
    sampleFaultToleranceRuleSubscriber := new(SampleFaultToleranceRuleSubscriber)
    // 3. construct SubscribeInfo
    faultToleranceSubscribeInfo := client.NewSubscribeInfo(faultToleranceSubscribeKey)
    faultToleranceSubscribeInfo.AppendSubscriber(sampleFaultToleranceRuleSubscriber)
    // 4. do register
    openSergoClient.RegisterSubscribeInfo(faultToleranceSubscribeInfo)
    
    // register SubscribeInfo of RateLimitStrategy
    rateLimitSubscribeKey := subscribe.NewSubscribeKey("default", "foo-app", configkind.ConfigKindRefRateLimitStrategy{})
    sampleRateLimitStrategySubscriber := new(SampleRateLimitStrategySubscriber)
    rateLimitSubscribeInfo := client.NewSubscribeInfo(rateLimitSubscribeKey)
    rateLimitSubscribeInfo.AppendSubscriber(sampleRateLimitStrategySubscriber)
    openSergoClient.RegisterSubscribeInfo(rateLimitSubscribeInfo)
    
    // start OpensergoClient
    openSergoClient.Start()
    
    // register after OpenSergoClient started
    faultToleranceSubscribeInfo.AppendSubscriber(new(subscribe.DefaultSubscriber))
    openSergoClient.RegisterSubscribeInfo(faultToleranceSubscribeInfo)
    
    // register after OpenSergoClient started
    rateLimitSubscribeInfo.AppendSubscriber(new(subscribe.DefaultSubscriber))
    openSergoClient.RegisterSubscribeInfo(rateLimitSubscribeInfo)
    
    select {}
}
```

## Demo

For more demo detail, please refer to [samples](./samples)