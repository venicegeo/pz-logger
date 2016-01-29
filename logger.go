package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var pzService *piazza.PzService

var startTime = time.Now()

type LogData struct {
	data []piazza.LogMessage
	sync.Mutex
}

var logData LogData

var debugMode bool

func handleGetRoot(c *gin.Context) {
	c.String(http.StatusOK, "Hi. I'm pz-logger.")
}

func handlePostMessages(c *gin.Context) {
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

	log.Printf("PZLOG: %s\n", mssg.ToString())

	logData.Lock()
	logData.data = append(logData.data, mssg)
	logData.Unlock()
}

func handleGetAdminStats(c *gin.Context) {
	logData.Lock()
	n := len(logData.data)
	logData.Unlock()
	m := piazza.AdminResponse{StartTime: startTime, Logger: &piazza.AdminResponseLogger{NumMessages: n}}
	c.JSON(http.StatusOK, m)
}

func handleGetAdminSettings(c *gin.Context) {
	s := "false"
	if debugMode {
		s = "true"
	}
	m := map[string]string{"debug": s}
	c.JSON(http.StatusOK, m)
}

func handlePostAdminSettings(c *gin.Context) {
	m := map[string]string{}
	err := c.BindJSON(&m)
	if err != nil {
		c.Error(err)
		return
	}
	for k, v := range m {
		switch k {
		case "debug":
			switch v {
			case "true":
				debugMode = true
				break
			case "false":
				debugMode = false
			default:
				c.String(http.StatusBadRequest, "Illegal value for 'debug': %s", v)
				return
			}
		default:
			c.String(http.StatusBadRequest, "Unknown parameter: %s", k)
			return
		}
	}
	c.JSON(http.StatusOK, m)
}

func handlePostAdminShutdown(c *gin.Context) {
	var reason string
	err := c.BindJSON(&reason)
	if err != nil {
		c.String(http.StatusBadRequest, "no reason supplied")
		return
	}
	pzService.Log(piazza.SeverityFatal, "Shutdown requested: "+reason)

	// TODO: need a graceful shutdown method
	os.Exit(0)
}

func handleGetMessages(c *gin.Context) {
	var err error
	count := 128
	key := c.Query("count")
	if key != "" {
		count, err = strconv.Atoi(key)
		if err != nil {
			c.String(http.StatusBadRequest, "query argument invalid: %s", key)
			return
		}
	}

	// copy up to count elements from the end of the log array
	logData.Lock()
	l := len(logData.data)
	if count > l {
		count = l
	}
	lines := make([]piazza.LogMessage, count)
	j := l - count
	for i := 0; i < count; i++ {
		lines[i] = logData.data[j]
		j++
	}
	logData.Unlock()

	c.JSON(http.StatusOK, lines)
}

func runLoggerServer() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	//router.Use(gin.Logger())
	//router.Use(gin.Recovery())

	router.GET("/", func(c *gin.Context) { handleGetRoot(c) })

	router.POST("/v1/messages", func(c *gin.Context) { handlePostMessages(c) })
	router.GET("/v1/messages", func(c *gin.Context) { handleGetMessages(c) })

	router.GET("/v1/admin/stats", func(c *gin.Context) { handleGetAdminStats(c) })

	router.GET("/v1/admin/settings", func(c *gin.Context) { handleGetAdminSettings(c) })
	router.POST("/v1/admin/settings", func(c *gin.Context) { handlePostAdminSettings(c) })

	router.POST("/v1/admin/shutdown", func(c *gin.Context) { handlePostAdminShutdown(c) })

	return router.Run(pzService.Address)
}

func app(done chan bool) int {

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

	if done != nil {
		done <- true
	}

	err = runLoggerServer()
	if err != nil {
		log.Print(err)
		return 1
	}

	// not reached
	return 1
}

func main2(cmd string, done chan bool) int {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = strings.Fields("main_tester " + cmd)
	return app(done)
}

func main() {
	os.Exit(app(nil))
}
