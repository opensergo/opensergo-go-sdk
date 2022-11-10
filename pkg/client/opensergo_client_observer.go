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

package client

import (
	"fmt"
	"github.com/golang/groupcache"
	"github.com/opensergo/opensergo-go/pkg/common/logging"
	"github.com/opensergo/opensergo-go/pkg/configkind"
	transportPb "github.com/opensergo/opensergo-go/pkg/proto/transport/v1"
	"github.com/opensergo/opensergo-go/pkg/transport"
	"github.com/opensergo/opensergo-go/pkg/transport/subscribe"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"reflect"
)

// SubscribeConfigStreamObserver is a struct as an observer role to listen the response of SubscribeConfig() from opensergo-control-plane.
type SubscribeConfigStreamObserver struct {
	openSergoClient *OpenSergoClient

	subscribeResponseChan  chan transportPb.SubscribeResponse
	receiveGoroutineOpened groupcache.AtomicInt
	handleGoroutineOpened  groupcache.AtomicInt
}

func NewSubscribeConfigStreamObserver(openSergoClient *OpenSergoClient) *SubscribeConfigStreamObserver {
	return &SubscribeConfigStreamObserver{
		subscribeResponseChan: make(chan transportPb.SubscribeResponse, 1024),
		openSergoClient:       openSergoClient,
	}
}

func (scObserver *SubscribeConfigStreamObserver) Start() {
	if scObserver.receiveGoroutineOpened.Get() == 0 {
		// keepalive OpensergoClient
		logging.Info("[OpenSergo SDK] open observeReceive() goroutine to observe msg from opensergo-control-plane to subscribeResponseChannel.")
		go scObserver.observeReceive()
	}
	if scObserver.handleGoroutineOpened.Get() == 0 {
		// keepalive OpensergoClient
		logging.Info("[OpenSergo SDK] open doHandle() goroutine to handle msg in subscribeResponseChannel.")
		go scObserver.doHandle()
	}
}

// observeReceive receive message from opensergo-control-plane.
//
// Contains a recursive invoke.
// Push the subscribeResponse into the subscribeResponseChan channel in SubscribeConfigStreamObserver
func (scObserver *SubscribeConfigStreamObserver) observeReceive() {
	// handle the panic of observeReceive()
	defer func() {
		if r := recover(); r != nil {
			errRecover := errors.Errorf("%+v", r)
			logging.Error(errRecover, "[OpenSergo SDK] error when receive config-data.")
			scObserver.openSergoClient.subscribeConfigStream = nil
			scObserver.openSergoClient.status.Store(interrupted)
			scObserver.receiveGoroutineOpened = groupcache.AtomicInt(0)
		}
	}()

	scObserver.receiveGoroutineOpened = groupcache.AtomicInt(1)
	var errorLocal error
	if scObserver.openSergoClient.subscribeConfigStream == nil {
		errorLocal = errors.New("can not doReceive(), grpc Stream subscribeConfigStream is nil")
	}
	subscribeResponse, err := scObserver.openSergoClient.subscribeConfigStream.Recv()
	// TODO add handles for different errors from grpc response.
	if err != nil {
		errorLocal = err
	}
	if errorLocal != nil {
		logging.Error(errorLocal, "error when receive config-data.")
		scObserver.openSergoClient.subscribeConfigStream = nil
		scObserver.openSergoClient.status.Store(interrupted)
		scObserver.receiveGoroutineOpened = groupcache.AtomicInt(0)
		logging.Warn("[OpenSergo SDK] end observeReceive()")
		return
	}

	scObserver.subscribeResponseChan <- *subscribeResponse

	scObserver.observeReceive()
}

// doHandle handle the received messages from opensergo-control-plane.
//
// Contains a recursive invoke.
// Handle the subscribeResponse from the subscribeResponseChan channel in SubscribeConfigStreamObserver
func (scObserver *SubscribeConfigStreamObserver) doHandle() {
	scObserver.handleGoroutineOpened = groupcache.AtomicInt(1)

	subscribeResponse := <-scObserver.subscribeResponseChan
	scObserver.handleReceive(subscribeResponse)

	scObserver.doHandle()
}

// handleReceive handle the received response of SubscribeConfig() from opensergo-control-plane.
func (scObserver *SubscribeConfigStreamObserver) handleReceive(subscribeResponse transportPb.SubscribeResponse) {

	ack := subscribeResponse.GetAck()
	// response of client-send
	if ack != "" {
		code := subscribeResponse.GetStatus().GetCode()
		if transport.CODE_SUCCESS == code {
			return
		}

		if code >= transport.CODE_ERROR_UNKNOWN && code < transport.CODE_ERROR_UPPER_BOUND {
			// TODO log info code out of bound
			return
		}
	}

	// handle message of server-push
	defer func() {
		if r := recover(); r != nil {
			errRecover := errors.Errorf("%+v", r)
			logging.Error(errRecover, fmt.Sprintf("[OpenSergo SDK] [subscribeResponseId:%v] panic occurred when invoking doHandleReceive() for subscribe.", subscribeResponse.ResponseId))
			if scObserver.openSergoClient.status.Load() == started {
				subscribeRequest := &transportPb.SubscribeRequest{
					Status:      &transportPb.Status{Code: transport.CODE_ERROR_UNKNOWN},
					ResponseAck: transport.FLAG_NACK,
					RequestId:   subscribeResponse.ResponseId,
				}
				scObserver.openSergoClient.subscribeConfigStream.Send(subscribeRequest)
			}
		}
	}()

	if err := scObserver.doHandleReceive(subscribeResponse); err != nil {
		logging.Error(err, "[OpenSergo SDK] err occurred when invoke doHandleReceive().")
	}

}

func (scObserver *SubscribeConfigStreamObserver) doHandleReceive(subscribeResponse transportPb.SubscribeResponse) error {
	kindName := subscribeResponse.GetKind()
	kindMetadata := configkind.GetConfigKindMetadataRegistry().GetKindMetadataByName(kindName)
	if (kindMetadata == configkind.ConfigKindMetadata{}) {
		// TODO log unrecognized config kind
		return errors.New("[OpenSergo SDK] unrecognized config kind: " + kindName)
	}

	subscribeKey := subscribe.NewSubscribeKey(subscribeResponse.GetNamespace(), subscribeResponse.GetApp(), kindMetadata.GetConfigKind())
	dataWithVersion := subscribeResponse.GetDataWithVersion()
	subscribeDataNotifyResult := scObserver.onSubscribeDataNotify(*subscribeKey, *dataWithVersion)

	code := subscribeDataNotifyResult.Code

	// TODO: handle partial-success (i.e. the data has been updated to cache, but error occurred in subscribers)
	var status *transportPb.Status
	switch code {
	case transport.CODE_SUCCESS:
		status = &transportPb.Status{Code: transport.CODE_SUCCESS}
		break
	case transport.CODE_ERROR_SUBSCRIBE_HANDLER_ERROR:
		var message string
		for _, notifyError := range subscribeDataNotifyResult.NotifyErrors {
			message = notifyError.Error() + "|"
		}

		status = &transportPb.Status{Message: message, Code: transport.CODE_ERROR_SUBSCRIBE_HANDLER_ERROR}
		break
	case transport.CODE_ERROR_VERSION_OUTDATED:
		status = &transportPb.Status{Code: transport.CODE_ERROR_VERSION_OUTDATED}
		break
	default:
		status = &transportPb.Status{Code: subscribeDataNotifyResult.Code}
	}

	subscribeRequest := &transportPb.SubscribeRequest{
		Status:      status,
		ResponseAck: transport.FLAG_ACK,
		RequestId:   subscribeResponse.ResponseId,
	}
	scObserver.openSergoClient.subscribeConfigStream.Send(subscribeRequest)
	return nil
}

// onSubscribeDataNotify to notify all the Subscriber matched the SubscribeKey to update data
func (scObserver *SubscribeConfigStreamObserver) onSubscribeDataNotify(subscribeKey subscribe.SubscribeKey, dataWithVersion transportPb.DataWithVersion) *subscribe.SubscribeDataNotifyResult {
	receivedVersion := dataWithVersion.GetVersion()
	subscribeDataCache := scObserver.openSergoClient.subscribeDataCache
	cachedData := subscribeDataCache.GetSubscribedData(subscribeKey)
	if (cachedData != subscribe.SubscribedData{}) && cachedData.GetVersion() > receivedVersion {
		// The upcoming data is out-dated, so we'll not resolve the push request.
		return &subscribe.SubscribeDataNotifyResult{
			Code: transport.CODE_ERROR_VERSION_OUTDATED,
		}
	}

	decodeData, err := decodeSubscribeData(subscribeKey.GetKind().GetName(), dataWithVersion.Data)
	if err != nil {
		return &subscribe.SubscribeDataNotifyResult{
			Code:         transport.CODE_ERROR_SUBSCRIBE_HANDLER_ERROR,
			DecodedData:  decodeData,
			NotifyErrors: []error{err},
		}
	}

	subscribeDataCache.UpdateSubscribedData(subscribeKey, decodeData, receivedVersion)

	subscribers := scObserver.openSergoClient.subscriberRegistry.GetSubscribersOf(subscribeKey)
	if subscribers == nil || len(subscribers) == 0 {
		return subscribe.WithSuccessNotifyResult(decodeData)
	}

	var errorSlice []error
	for _, subscriber := range subscribers {
		if _, err := subscriber.OnSubscribeDataUpdate(subscribeKey, decodeData); err != nil {
			errorSlice = append(errorSlice, err)
		}
	}

	if len(errorSlice) == 0 {
		return subscribe.WithSuccessNotifyResult(decodeData)
	} else {
		return &subscribe.SubscribeDataNotifyResult{
			Code:         transport.CODE_ERROR_SUBSCRIBE_HANDLER_ERROR,
			DecodedData:  decodeData,
			NotifyErrors: errorSlice,
		}
	}
}

// decode the subscribed-data from protobuf
func decodeSubscribeData(kindName string, dataSlice []*anypb.Any) ([]protoreflect.ProtoMessage, error) {
	configKindMetadataRegistry := configkind.GetConfigKindMetadataRegistry()
	configKindMetadata := configKindMetadataRegistry.GetKindMetadataByName(kindName)

	if reflect.DeepEqual(configKindMetadata, configkind.ConfigKindMetadata{}) {
		return nil, errors.New("unrecognized config kind: " + kindName)
	}
	var decodeDataSlice []protoreflect.ProtoMessage
	protoMessage := configKindMetadata.GetKindPbMessageType().New().Interface()
	for _, data := range dataSlice {
		anypb.UnmarshalTo(data, protoMessage, proto.UnmarshalOptions{})
		decodeDataSlice = append(decodeDataSlice, protoMessage)
	}

	return decodeDataSlice, nil
}
