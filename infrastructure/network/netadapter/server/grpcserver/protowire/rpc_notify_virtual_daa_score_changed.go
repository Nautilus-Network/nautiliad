package protowire

import (
	"github.com/Nautilus-Network/nautiliad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *NautiliadMessage_NotifyVirtualDaaScoreChangedRequest) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NautiliadMessage_NotifyVirtualDaaScoreChangedRequest is nil")
	}
	return &appmessage.NotifyVirtualDaaScoreChangedRequestMessage{}, nil
}

func (x *NautiliadMessage_NotifyVirtualDaaScoreChangedRequest) fromAppMessage(_ *appmessage.NotifyVirtualDaaScoreChangedRequestMessage) error {
	x.NotifyVirtualDaaScoreChangedRequest = &NotifyVirtualDaaScoreChangedRequestMessage{}
	return nil
}

func (x *NautiliadMessage_NotifyVirtualDaaScoreChangedResponse) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NautiliadMessage_NotifyVirtualDaaScoreChangedResponse is nil")
	}
	return x.NotifyVirtualDaaScoreChangedResponse.toAppMessage()
}

func (x *NautiliadMessage_NotifyVirtualDaaScoreChangedResponse) fromAppMessage(message *appmessage.NotifyVirtualDaaScoreChangedResponseMessage) error {
	var err *RPCError
	if message.Error != nil {
		err = &RPCError{Message: message.Error.Message}
	}
	x.NotifyVirtualDaaScoreChangedResponse = &NotifyVirtualDaaScoreChangedResponseMessage{
		Error: err,
	}
	return nil
}

func (x *NotifyVirtualDaaScoreChangedResponseMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NotifyVirtualDaaScoreChangedResponseMessage is nil")
	}
	rpcErr, err := x.Error.toAppMessage()
	// Error is an optional field
	if err != nil && !errors.Is(err, errorNil) {
		return nil, err
	}
	return &appmessage.NotifyVirtualDaaScoreChangedResponseMessage{
		Error: rpcErr,
	}, nil
}

func (x *NautiliadMessage_VirtualDaaScoreChangedNotification) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NautiliadMessage_VirtualDaaScoreChangedNotification is nil")
	}
	return x.VirtualDaaScoreChangedNotification.toAppMessage()
}

func (x *NautiliadMessage_VirtualDaaScoreChangedNotification) fromAppMessage(message *appmessage.VirtualDaaScoreChangedNotificationMessage) error {
	x.VirtualDaaScoreChangedNotification = &VirtualDaaScoreChangedNotificationMessage{
		VirtualDaaScore: message.VirtualDaaScore,
	}
	return nil
}

func (x *VirtualDaaScoreChangedNotificationMessage) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "VirtualDaaScoreChangedNotificationMessage is nil")
	}
	return &appmessage.VirtualDaaScoreChangedNotificationMessage{
		VirtualDaaScore: x.VirtualDaaScore,
	}, nil
}
