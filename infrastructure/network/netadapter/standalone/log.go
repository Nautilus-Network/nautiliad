package standalone

import (
	"github.com/Nautilus-Network/nautiliad/infrastructure/logger"
	"github.com/Nautilus-Network/nautiliad/util/panics"
)

var log = logger.RegisterSubSystem("NTAR")
var spawn = panics.GoroutineWrapperFunc(log)
