package main

import (
	"github.com/jotitan/dico-definition-game/src/logger"
	"github.com/jotitan/dico-definition-game/src/server"
	"os"
)

func main(){
	//loadDicoFromDatasource()
	// Only one parameter, folder where resources are
	if len(os.Args) != 2 {
		logger.GetLogger2().Fatal("Need one parameter, resources folder")
	}
	server.NewGameServer(os.Args[1]).Run()
}

