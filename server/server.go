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
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
)

type LockedAdminSettings struct {
	sync.Mutex
	LoggerAdminSettings
}

var settings LockedAdminSettings

type LockedAdminStats struct {
	sync.Mutex
	LoggerAdminStats
}

var stats LockedAdminStats

type LogData struct {
	sync.Mutex
	esIndex elasticsearch.IIndex
	id      int
}

var logData LogData

func initServer(sys *piazza.SystemConfig, esIndex elasticsearch.IIndex) {
	var err error

	stats.StartTime = time.Now()

	if !esIndex.IndexExists() {
		err = esIndex.Create()
		if err != nil {
			log.Fatal(err)
		}
		mapping :=
			`{
		    "LogData":{
			    "properties":{
				    "service":{
					    "type": "string",
                        "store": true
    			    },
				    "address":{
					    "type": "string",
                        "store": true
    			    },
				    "time":{
					    "type": "string",
                        "store": true
    			    },
				    "severity":{
					    "type": "string",
                        "store": true
    			    },
				    "message":{
					    "type": "string",
                        "store": true
    			    }
	    	    }
	        }
        }`

		err = esIndex.SetMapping("LogData", piazza.JsonString(mapping))
		if err != nil {
			log.Fatal(err)
		}
	}

	logData.esIndex = esIndex
}

func handleGetRoot(c *gin.Context) {
	//log.Print("got health-check request")
	c.String(http.StatusOK, "Hi. I'm pz-logger.")
}

func handlePostMessages(c *gin.Context) {
	var mssg LogMessage
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

	log.Printf("PZLOG: %s\n", mssg.String())

	logData.Lock()
	idStr := strconv.Itoa(logData.id)
	logData.id++
	logData.Unlock()
	indexResult, err := logData.esIndex.PostData("LogData", idStr, mssg)
	if err != nil {
		c.String(http.StatusBadRequest, "%v", err)
		return
	}
	if !indexResult.Created {
		c.String(http.StatusBadRequest, "POST of log data failed")
		return
	}

	err = logData.esIndex.Flush()
	if err != nil {
		c.String(http.StatusBadRequest, "%v", err)
		return
	}

	stats.LoggerAdminStats.NumMessages++

	c.JSON(http.StatusOK, nil)
}

func handleGetAdminStats(c *gin.Context) {
	logData.Lock()
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
	t := LoggerAdminSettings{}
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

	searchResult, err := logData.esIndex.FilterByMatchAll("LogData")
	if err != nil {
		c.String(http.StatusBadRequest, "query failed: %s", err)
		return
	}

	// TODO: unsafe truncation
	l := int(searchResult.TotalHits())
	if count > l {
		count = l
	}
	lines := make([]LogMessage, count)

	i := 0
	for _, hit := range *searchResult.GetHits() {
		var tmp LogMessage
		err = json.Unmarshal(*hit.Source, &tmp)
		if err != nil {
			c.String(http.StatusBadRequest, "query unmarshal failed: %s", err)
			return
		}
		lines[i] = tmp
		i++
	}

	c.JSON(http.StatusOK, lines)
}

func CreateHandlers(sys *piazza.SystemConfig, esi elasticsearch.IIndex) http.Handler {
	initServer(sys, esi)

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
