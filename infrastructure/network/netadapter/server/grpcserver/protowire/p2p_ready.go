package protowire

import (
	"github.com/Nautilus-Network/nautiliad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *NautiliadMessage_Ready) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "NautiliadMessage_Ready is nil")
	}
	return &appmessage.MsgReady{}, nil
}

func (x *NautiliadMessage_Ready) fromAppMessage(_ *appmessage.MsgReady) error {
	return nil
}
