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
	"reflect"

	"github.com/opensergo/opensergo-go/pkg/common/logging"
	"github.com/opensergo/opensergo-go/pkg/configkind"
	"github.com/opensergo/opensergo-go/pkg/model"
	transportPb "github.com/opensergo/opensergo-go/pkg/proto/transport/v1"
	"github.com/opensergo/opensergo-go/pkg/transport"
	"github.com/opensergo/opensergo-go/pkg/transport/subscribe"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

// SubscribeConfigStreamObserver is a struct as an observer role to listen the response of SubscribeConfig() from opensergo-control-plane.
type SubscribeConfigStreamObserver struct {
	openSergoClient *OpenSergoClient

	subscribeResponseChan chan *transportPb.SubscribeResponse
}

func NewSubscribeConfigStreamObserver(openSergoClient *OpenSergoClient) *SubscribeConfigStreamObserver {
	scObserver := &SubscribeConfigStreamObserver{
		subscribeResponseChan: make(chan *transportPb.SubscribeResponse, 1024),
		openSergoClient:       openSergoClient,
	}
	return scObserver
}

func (scObserver *SubscribeConfigStreamObserver) SetOpenSergoClient(openSergoClient *OpenSergoClient) {
	scObserver.openSergoClient = openSergoClient
}

func (scObserver *SubscribeConfigStreamObserver) Start() {
	logging.Info("[OpenSergo SDK] starting SubscribeConfigStreamObserver.")

	logging.Info("[OpenSergo SDK] open observeReceive() goroutine to observe msg from opensergo-control-plane to subscribeResponseChannel.")
	go scObserver.doObserve()

	logging.Info("[OpenSergo SDK] open doHandle() goroutine to handle msg in subscribeResponseChannel.")
	go scObserver.doHandle()

	logging.Info("[OpenSergo SDK] started SubscribeConfigStreamObserver with goroutines of doObserve() and doHandle().")
}

func (scObserver *SubscribeConfigStreamObserver) doObserve() {
	for {
		if scObserver.openSergoClient.subscribeConfigStreamPtr.Load().(*subscribeConfigStream).stream != nil {
			subscribeResponse, err := scObserver.observeReceive()
			if err == nil && subscribeResponse != nil {
				scObserver.subscribeResponseChan <- subscribeResponse
			}
		}
	}
}

// observeReceive receive message from opensergo-control-plane.
//
// Push the subscribeResponse into the subscribeResponseChan channel in SubscribeConfigStreamObserver
func (scObserver *SubscribeConfigStreamObserver) observeReceive() (resp *transportPb.SubscribeResponse, err error) {
	// handle the panic of observeReceive()
	defer func() {
		if r := recover(); r != nil {
			errRecover := errors.Errorf("%+v", r)
			scObserver.openSergoClient.subscribeConfigStreamPtr.Store(&subscribeConfigStream{})
			scObserver.openSergoClient.subscribeConfigStreamStatus.Store(interrupted)
			logging.Error(errRecover, "[OpenSergo SDK] interrupted gRpc Stream (SubscribeConfigStream) because of panic occurring when receive data from opensergo-control-plane.")
			resp = nil
			err = errRecover
		}
	}()

	var errorLocal error
	subscribeResponse, err := scObserver.openSergoClient.subscribeConfigStreamPtr.Load().(*subscribeConfigStream).stream.Recv()
	// TODO add handles for different errors from grpc response.
	if err != nil {
		errorLocal = err
	}

	if errorLocal != nil {
		logging.Error(errorLocal, "error when receive config-data.")
		scObserver.openSergoClient.subscribeConfigStreamPtr.Store(&subscribeConfigStream{})
		scObserver.openSergoClient.subscribeConfigStreamStatus.Store(interrupted)
		logging.Warn("[OpenSergo SDK] interrupted gRpc Stream (SubscribeConfigStream) because of error occurring when receive data from opensergo-control-plane.")
		return nil, errorLocal
	}

	return subscribeResponse, nil
}

// doHandle handle the received messages from opensergo-control-plane.
//
// Handle the subscribeResponse from the subscribeResponseChan channel in SubscribeConfigStreamObserver
func (scObserver *SubscribeConfigStreamObserver) doHandle() {
	for {
		if scObserver.openSergoClient.subscribeConfigStreamPtr.Load().(*subscribeConfigStream).stream != nil {
			select {
			case subscribeResponse := <-scObserver.subscribeResponseChan:
				scObserver.handleReceive(subscribeResponse)
				break
			}
		}
	}
}

// handleReceive handle the received response of SubscribeConfig() from opensergo-control-plane.
func (scObserver *SubscribeConfigStreamObserver) handleReceive(subscribeResponse *transportPb.SubscribeResponse) {

	ack := subscribeResponse.GetAck()
	// response of client-send
	if ack != "" {
		code := subscribeResponse.GetStatus().GetCode()
		if transport.CodeSuccess == code {
			return
		}

		if code >= transport.CodeErrorUnknown && code < transport.CodeErrorUpperBound {
			// TODO log info code out of bound
			return
		}
	}

	// handle message of server-push
	defer func() {
		if r := recover(); r != nil {
			errRecover := errors.Errorf("%+v", r)
			logging.Error(errRecover, fmt.Sprintf("[OpenSergo SDK] [subscribeResponseId:%v] panic occurred when invoking doHandleReceive() for subscribe.", subscribeResponse.ResponseId))
			if scObserver.openSergoClient.subscribeConfigStreamStatus.Load() == started {
				subscribeRequest := &transportPb.SubscribeRequest{
					Status:      &transportPb.Status{Code: transport.CodeErrorUnknown},
					ResponseAck: transport.FlagNack,
					RequestId:   subscribeResponse.ResponseId,
				}
				if err := scObserver.openSergoClient.subscribeConfigStreamPtr.Load().(*subscribeConfigStream).stream.Send(subscribeRequest); err != nil {
					logging.Error(err, "error occurred when send NACK-request for SubscribeConfig")
				}
			}
		}
	}()

	if err := scObserver.doHandleReceive(subscribeResponse); err != nil {
		logging.Error(err, "[OpenSergo SDK] err occurred when invoke doHandleReceive().")
	}
}

func (scObserver *SubscribeConfigStreamObserver) doHandleReceive(subscribeResponse *transportPb.SubscribeResponse) error {
	kindName := subscribeResponse.GetKind()
	kindMetadata := configkind.GetConfigKindMetadataRegistry().GetKindMetadataByName(kindName)
	if (kindMetadata == configkind.ConfigKindMetadata{}) {
		// TODO log unrecognized config kind
		return errors.New("[OpenSergo SDK] unrecognized config kind: " + kindName)
	}

	subscribeKey := model.NewSubscribeKey(subscribeResponse.GetNamespace(), subscribeResponse.GetApp(), kindMetadata.GetConfigKind())
	dataWithVersion := subscribeResponse.GetDataWithVersion()
	subscribeDataNotifyResult := scObserver.onSubscribeDataNotify(*subscribeKey, dataWithVersion)

	code := subscribeDataNotifyResult.Code

	// TODO: handle partial-success (i.e. the data has been updated to cache, but error occurred in subscribers)
	var status *transportPb.Status
	switch code {
	case transport.CodeSuccess:
		status = &transportPb.Status{Code: transport.CodeSuccess}
		break
	case transport.CodeErrorSubscribeHandlerError:
		var message string
		for _, notifyError := range subscribeDataNotifyResult.NotifyErrors {
			message = notifyError.Error() + "|"
		}

		status = &transportPb.Status{Message: message, Code: transport.CodeErrorSubscribeHandlerError}
		break
	case transport.CodeErrorVersionOutdated:
		status = &transportPb.Status{Code: transport.CodeErrorVersionOutdated}
		break
	default:
		status = &transportPb.Status{Code: subscribeDataNotifyResult.Code}
	}

	subscribeRequest := &transportPb.SubscribeRequest{
		Status:      status,
		ResponseAck: transport.FlagAck,
		RequestId:   subscribeResponse.ResponseId,
	}
	if err := scObserver.openSergoClient.subscribeConfigStreamPtr.Load().(*subscribeConfigStream).stream.Send(subscribeRequest); err != nil {
		logging.Error(err, "error occurred when send ACK-request for SubscribeConfig")
		return err
	}
	return nil
}

// onSubscribeDataNotify to notify all the Subscriber matched the SubscribeKey to update data
func (scObserver *SubscribeConfigStreamObserver) onSubscribeDataNotify(subscribeKey model.SubscribeKey, dataWithVersion *transportPb.DataWithVersion) *subscribe.SubscribeDataNotifyResult {
	receivedVersion := dataWithVersion.GetVersion()
	subscribeDataCache := scObserver.openSergoClient.subscribeDataCache
	cachedData := subscribeDataCache.GetSubscribedData(subscribeKey)
	if (cachedData != nil) && cachedData.Version() > receivedVersion {
		// The upcoming data is out-dated, so we'll not resolve the push request.
		return &subscribe.SubscribeDataNotifyResult{
			Code: transport.CodeErrorVersionOutdated,
		}
	}

	decodeData, err := decodeSubscribeData(subscribeKey.Kind().GetName(), dataWithVersion.Data)
	if err != nil {
		return &subscribe.SubscribeDataNotifyResult{
			Code:         transport.CodeErrorSubscribeHandlerError,
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
			Code:         transport.CodeErrorSubscribeHandlerError,
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
	for _, data := range dataSlice {
		protoMessage := configKindMetadata.GetKindPbMessageType().New().Interface()
		if err := anypb.UnmarshalTo(data, protoMessage, proto.UnmarshalOptions{}); err != nil {
			return nil, err
		}
		decodeDataSlice = append(decodeDataSlice, protoMessage)
	}

	return decodeDataSlice, nil
}
