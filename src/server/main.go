package main

import (
	"github.com/name5566/leaf"
	leafConf "github.com/name5566/leaf/conf"
	"proj_bcbm/src/server/conf"
	"proj_bcbm/src/server/game"
	"proj_bcbm/src/server/gate"
	"proj_bcbm/src/server/log"
	"proj_bcbm/src/server/login"
)

func main() {
	logger, err := log.New(conf.Server.LogLevel, conf.Server.LogPath, conf.LogFlag, conf.Server.LogServer)
	if err != nil {
		panic(err)
	}
	log.Export(logger)
	defer logger.Close()

	leafConf.ConsolePort = conf.Server.ConsolePort
	leafConf.ProfilePath = conf.Server.ProfilePath

	leaf.Run(
		game.Module,
		gate.Module,
		login.Module,
	)
}
