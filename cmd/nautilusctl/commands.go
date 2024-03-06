package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Nautilus-Network/nautiliad/infrastructure/network/netadapter/server/grpcserver/protowire"
)

var commandTypes = []reflect.Type{
	reflect.TypeOf(protowire.NautiliadMessage_AddPeerRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetConnectedPeerInfoRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetPeerAddressesRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetCurrentNetworkRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetInfoRequest{}),

	reflect.TypeOf(protowire.NautiliadMessage_GetBlockRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetBlocksRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetHeadersRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetBlockCountRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetBlockDagInfoRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetSelectedTipHashRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetVirtualSelectedParentBlueScoreRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetVirtualSelectedParentChainFromBlockRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_ResolveFinalityConflictRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_EstimateNetworkHashesPerSecondRequest{}),

	reflect.TypeOf(protowire.NautiliadMessage_GetBlockTemplateRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_SubmitBlockRequest{}),

	reflect.TypeOf(protowire.NautiliadMessage_GetMempoolEntryRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetMempoolEntriesRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetMempoolEntriesByAddressesRequest{}),

	reflect.TypeOf(protowire.NautiliadMessage_SubmitTransactionRequest{}),

	reflect.TypeOf(protowire.NautiliadMessage_GetUtxosByAddressesRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetBalanceByAddressRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_GetCoinSupplyRequest{}),

	reflect.TypeOf(protowire.NautiliadMessage_BanRequest{}),
	reflect.TypeOf(protowire.NautiliadMessage_UnbanRequest{}),
}

type commandDescription struct {
	name       string
	parameters []*parameterDescription
	typeof     reflect.Type
}

type parameterDescription struct {
	name   string
	typeof reflect.Type
}

func commandDescriptions() []*commandDescription {
	commandDescriptions := make([]*commandDescription, len(commandTypes))

	for i, commandTypeWrapped := range commandTypes {
		commandType := unwrapCommandType(commandTypeWrapped)

		name := strings.TrimSuffix(commandType.Name(), "RequestMessage")
		numFields := commandType.NumField()

		var parameters []*parameterDescription
		for i := 0; i < numFields; i++ {
			field := commandType.Field(i)

			if !isFieldExported(field) {
				continue
			}

			parameters = append(parameters, &parameterDescription{
				name:   field.Name,
				typeof: field.Type,
			})
		}
		commandDescriptions[i] = &commandDescription{
			name:       name,
			parameters: parameters,
			typeof:     commandTypeWrapped,
		}
	}

	return commandDescriptions
}

func (cd *commandDescription) help() string {
	sb := &strings.Builder{}
	sb.WriteString(cd.name)
	for _, parameter := range cd.parameters {
		_, _ = fmt.Fprintf(sb, " [%s]", parameter.name)
	}
	return sb.String()
}
