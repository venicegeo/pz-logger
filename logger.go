package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var pzService *piazza.PzService

var startTime = time.Now()

var logData []piazza.LogMessage

func handleHealthCheck(c *gin.Context) {
	c.String(http.StatusOK, "Hi. I'm pz-logger.")
}

func handleLoggerPost(c *gin.Context) {
	var mssg piazza.LogMessage
	err := c.BindJSON(&mssg)
	if err != nil {
		c.String(http.StatusBadRequest, "%v", err)
		return
	}

	err = mssg.Validate()
	if err != nil {
		c.String(http.StatusBadRequest, "%v", err)
		return
	}

	log.Printf("LOG: %s\n", mssg.ToString())

	logData = append(logData, mssg)
	//c.IndentedJSON(http.StatusOK,)
}

func handleAdminGet(c *gin.Context) {
	m := piazza.AdminResponse{StartTime: startTime, Logger: &piazza.AdminResponseLogger{NumMessages: len(logData)}}

	c.JSON(http.StatusOK, m)
}

func handleLoggerGet(c *gin.Context) {

	c.JSON(http.StatusOK, logData)
}

func runLoggerServer() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	//router.Use(gin.Logger())
	//router.Use(gin.Recovery())

	router.GET("/log/admin", func(c *gin.Context) {
		handleAdminGet(c)
	})
	router.POST("/log", func(c *gin.Context) {
		handleLoggerPost(c)
	})

	router.GET("/log", func(c *gin.Context) {
		handleLoggerGet(c)
	})
	router.GET("/", func(c *gin.Context) {
		handleHealthCheck(c)
	})

	return router.Run(pzService.Address)
}

func app() int {

	var err error

	// handles the command line flags, finds the discover service, registers us,
	// and figures out our own server address
	serviceAddress, discoverAddress, debug, err := piazza.NewDiscoverService("pz-logger", "localhost:12341", "localhost:3000")
	if err != nil {
		log.Print(err)
		return 1
	}

	pzService, err = piazza.NewPzService("pz-logger", serviceAddress, discoverAddress, debug)
	if err != nil {
		log.Fatal(err)
		return 1
	}

	err = runLoggerServer()
	if err != nil {
		log.Print(err)
		return 1
	}

	// not reached
	return 1
}

func main2(cmd string) int {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = strings.Fields("main_tester " + cmd)
	return app()
}

func main() {
	os.Exit(app())
}
