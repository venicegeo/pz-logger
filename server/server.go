// Copyright 2016, RadiantBlue Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

type LockedAdminSettings struct {
	sync.Mutex
	client.LoggerAdminSettings
}

var settings LockedAdminSettings

type LockedAdminStats struct {
	sync.Mutex
	client.LoggerAdminStats
}

var stats LockedAdminStats

type LogData struct {
	sync.Mutex
	data []client.LogMessage
}

var logData LogData

func init() {
	stats.StartTime = time.Now()
}

func handleGetRoot(c *gin.Context) {
	log.Print("got health-check request")
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
	stats.LoggerAdminStats.NumMessages = len(logData.data)
	t := stats.LoggerAdminStats
	logData.Unlock()
	c.JSON(http.StatusOK, t)
}

func handleGetAdminSettings(c *gin.Context) {
	settings.Lock()
	t := settings.LoggerAdminSettings
	settings.Unlock()
	c.JSON(http.StatusOK, t)
}

func handlePostAdminSettings(c *gin.Context) {
	t := client.LoggerAdminSettings{}
	err := c.BindJSON(&t)
	if err != nil {
		c.Error(err)
		return
	}
	settings.Lock()
	settings.LoggerAdminSettings = t
	settings.Unlock()
	c.String(http.StatusOK, "")
}

func handlePostAdminShutdown(c *gin.Context) {
	piazza.HandlePostAdminShutdown(c)
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

func CreateHandlers(sys *piazza.System) http.Handler {

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

	return router
}
