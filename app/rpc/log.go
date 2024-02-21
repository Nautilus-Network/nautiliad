package rpc

import (
	"github.com/Nautilus-Network/nautiliad/infrastructure/logger"
	"github.com/Nautilus-Network/nautiliad/util/panics"
)

var log = logger.RegisterSubSystem("RPCS")
var spawn = panics.GoroutineWrapperFunc(log)
