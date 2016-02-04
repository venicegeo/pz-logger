package main

import (
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-logger/server"
	"log"
)

func main() {

	local := piazza.IsLocalConfig()

	config, err := piazza.GetConfig("pz-logger", local)
	if err != nil {
		log.Fatal(err)
	}

	discover, err := piazza.NewDiscoverClient(config)
	if err != nil {
		log.Fatal(err)
	}

	err = discover.RegisterServiceWithDiscover(config.ServiceName, config.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}

	err = server.RunLoggerServer(config)
	if err != nil {
		log.Fatal(err)
	}

	// not reached
	log.Fatal("not reached")
}
