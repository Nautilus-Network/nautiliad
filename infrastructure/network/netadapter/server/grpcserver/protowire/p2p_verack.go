package protowire

import (
	"github.com/Nautilus-Network/nautiliad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *NautiliadMessage_Verack) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NautiliadMessage_Verack is nil")
	}
	return &appmessage.MsgVerAck{}, nil
}

func (x *NautiliadMessage_Verack) fromAppMessage(_ *appmessage.MsgVerAck) error {
	return nil
}
