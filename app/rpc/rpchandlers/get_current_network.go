package rpchandlers

import (
	"github.com/Nautilus-Network/nautiliad/app/appmessage"
	"github.com/Nautilus-Network/nautiliad/app/rpc/rpccontext"
	"github.com/Nautilus-Network/nautiliad/infrastructure/network/netadapter/router"
)

// HandleGetCurrentNetwork handles the respectively named RPC command
func HandleGetCurrentNetwork(context *rpccontext.Context, _ *router.Router, _ appmessage.Message) (appmessage.Message, error) {
	response := appmessage.NewGetCurrentNetworkResponseMessage(context.Config.ActiveNetParams.Net.String())
	return response, nil
}
