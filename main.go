package main

import (
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-logger/server"
	"log"
	"os"
)

var pzService *piazza.PzService

func Main(done chan bool, local bool) int {

	var err error

	// handles the command line flags, finds the discover service, registers us,
	// and figures out our own server address
	config, err := piazza.GetConfig("pz-logger", local)
	if err != nil {
		log.Fatal(err)
		return 1
	}

	err = config.RegisterServiceWithDiscover()
	if err != nil {
		log.Fatal(err)
		return 1
	}

	pzService, err = piazza.NewPzService(config, false)
	if err != nil {
		log.Fatal(err)
		return 1
	}

	if done != nil {
		done <- true
	}

	err = server.RunLoggerServer(config.BindTo, pzService)
	if err != nil {
		log.Print(err)
		return 1
	}

	// not reached
	return 1
}

func main() {
	local := piazza.IsLocalConfig()
	os.Exit(Main(nil, local))
}
