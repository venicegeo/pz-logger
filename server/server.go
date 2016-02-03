package server

import (
	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-logger/client"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var pzService *piazza.PzService // TODO

var startTime = time.Now()

type LogData struct {
	data []client.LogMessage
	sync.Mutex
}

var logData LogData

var debugMode bool

func handleGetRoot(c *gin.Context) {
	c.String(http.StatusOK, "Hi. I'm pz-logger.")
}

func handlePostMessages(c *gin.Context) {
	var mssg client.LogMessage
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
	m := client.LoggerAdminStats{StartTime: startTime, NumMessages: n}
	c.JSON(http.StatusOK, m)
}

func handleGetAdminSettings(c *gin.Context) {
	s := client.LoggerAdminSettings{Debug: debugMode}
	c.JSON(http.StatusOK, s)
}

func handlePostAdminSettings(c *gin.Context) {
	settings := client.LoggerAdminSettings{}
	err := c.BindJSON(&settings)
	if err != nil {
		c.Error(err)
		return
	}
	debugMode = settings.Debug
	c.String(http.StatusOK, "")
}

func handlePostAdminShutdown(c *gin.Context) {
	piazza.HandlePostAdminShutdown(pzService, c)
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
	lines := make([]client.LogMessage, count)
	j := l - count
	for i := 0; i < count; i++ {
		lines[i] = logData.data[j]
		j++
	}
	logData.Unlock()

	c.JSON(http.StatusOK, lines)
}


func RunLoggerServer(bindTo string, pzServiceParam *piazza.PzService) error {
	pzService = pzServiceParam

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

	return router.Run(bindTo)
}

