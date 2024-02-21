package main

import (
	"github.com/Nautilus-Network/nautiliad/infrastructure/logger"
	"github.com/Nautilus-Network/nautiliad/util/panics"
)

var (
	backendLog = logger.NewBackend()
	log        = backendLog.Logger("MNJS")
	spawn      = panics.GoroutineWrapperFunc(log)
)
