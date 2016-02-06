package main

import (
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-logger/server"
	"log"
	"github.com/venicegeo/pz-logger/client"
)

func main() {

	var mode piazza.ConfigMode = piazza.ConfigModeCloud
	if piazza.IsLocalConfig() {
		mode = piazza.ConfigModeLocal
	}

	config, err := piazza.NewConfig(piazza.PzLogger, mode)
	if err != nil {
		log.Fatal(err)
	}

	sys, err := piazza.NewSystem(config)
	if err != nil {
		log.Fatal(err)
	}

	done := sys.StartServer(server.CreateHandlers(sys))

	err = <- done
	if err != nil {
		log.Fatal(err)
	}
}
