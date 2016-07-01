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

package lib

import (
	"encoding/json"
	_ "fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
)

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

var schema = "LogData"

func initServer(sys *piazza.SystemConfig, esIndex elasticsearch.IIndex) {
	var err error

	stats.StartTime = time.Now()

	/***
	err = esIndex.Delete()
	if err != nil {
		log.Fatal(err)
	}
	if esIndex.IndexExists() {
		log.Fatal("index still exists")
	}
	err = esIndex.Create()
	if err != nil {
		log.Fatal(err)
	}
	***/

	if !esIndex.IndexExists() {
		err = esIndex.Create("")
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
				    "stamp":{
					    "type": "long",
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

		err = esIndex.SetMapping(schema, piazza.JsonString(mapping))
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
	var mssg Message
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
	indexResult, err := logData.esIndex.PostData(schema, idStr, mssg)
	if err != nil {
		c.String(http.StatusBadRequest, "%v", err)
		return
	}
	if !indexResult.Created {
		c.String(http.StatusBadRequest, "POST of log data failed")
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

func handleGetMessages(c *gin.Context) {
	var err error

	format := elasticsearch.GetFormatParams(c, 10, 0, "stamp", elasticsearch.SortDescending)
	filterParams := parseFilterParams(c)

	//log.Printf("size %d, from %d, key %s, format %v",
	//	format.Size, format.From, format.Key, format.Order)

	log.Printf("filterParams: %v\n", filterParams)

	var searchResult *elasticsearch.SearchResult

	if len(filterParams) == 0 {
		searchResult, err = logData.esIndex.FilterByMatchAll(schema, format)
	} else {
		var jsonString = createQueryDslAsString(format, filterParams)
		searchResult, err = logData.esIndex.SearchByJSON(schema, jsonString)
	}

	if err != nil {
		c.String(http.StatusBadRequest, "query failed: %s", err)
		return
	}

	// TODO: unsafe truncation
	count := int(searchResult.TotalHits())
	lines := make([]Message, count)

	i := 0
	for _, hit := range *searchResult.GetHits() {
		if hit == nil {
			log.Printf("null source hit")
			continue
		}
		src := *hit.Source
		//log.Printf("source hit: %s", string(src))

		tmp := &Message{}
		err = json.Unmarshal(src, tmp)
		if err != nil {
			log.Printf("UNABLE TO PARSE: %s", string(*hit.Source))
			c.String(http.StatusBadRequest, "query unmarshal failed: %s", err)
			return
		}
		err = tmp.Validate()
		if err != nil {
			log.Printf("UNABLE TO VALIDATE: %s", string(*hit.Source))
			//c.String(http.StatusBadRequest, "query unmarshal failed to validate: %s", err)
			//return
			continue
		}
		lines[i] = *tmp
		i++
	}

	c.JSON(http.StatusOK, lines)
}

func handleGetMessagesV2(c *gin.Context) {
	var err error

	format := elasticsearch.GetFormatParamsV2(c, 10, 0, "stamp", elasticsearch.SortDescending)
	filterParams := parseFilterParams(c)

	//log.Printf("size %d, from %d, key %s, format %v",
	//	format.Size, format.From, format.Key, format.Order)

	log.Printf("filterParams: %v\n", filterParams)

	var searchResult *elasticsearch.SearchResult

	if len(filterParams) == 0 {
		searchResult, err = logData.esIndex.FilterByMatchAll(schema, format)
	} else {
		var jsonString = createQueryDslAsString(format, filterParams)
		searchResult, err = logData.esIndex.SearchByJSON(schema, jsonString)
	}

	if err != nil {
		c.String(http.StatusBadRequest, "query failed: %s", err)
		return
	}

	// TODO: unsafe truncation
	count := searchResult.TotalHits()
	matched := searchResult.NumberMatched()
	lines := make([]Message, count)

	i := 0
	for _, hit := range *searchResult.GetHits() {
		if hit == nil {
			log.Printf("null source hit")
			continue
		}
		src := *hit.Source
		//log.Printf("source hit: %s", string(src))

		tmp := &Message{}
		err = json.Unmarshal(src, tmp)
		if err != nil {
			log.Printf("UNABLE TO PARSE: %s", string(*hit.Source))
			c.String(http.StatusBadRequest, "query unmarshal failed: %s", err)
			return
		}
		err = tmp.Validate()
		if err != nil {
			log.Printf("UNABLE TO VALIDATE: %s", string(*hit.Source))
			//c.String(http.StatusBadRequest, "query unmarshal failed to validate: %s", err)
			//return
			continue
		}
		lines[i] = *tmp
		i++
	}

	bar := make([]interface{}, len(lines))

	for i, e := range lines {
		bar[i] = e
	}

	var order string

	if format.Order {
		order = "desc"
	} else {
		order = "asc"
	}

	foo := &piazza.Common18FListResponse{
		Data: bar,
		Pagination: piazza.Pagination{
			Page:    format.From,
			PerPage: format.Size,
			Count:   matched,
			SortBy:  format.Key,
			Order:   order,
		},
	}

	// c.JSON(http.StatusOK, lines)
	c.JSON(http.StatusOK, foo)
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

	router.POST("/v2/message", func(c *gin.Context) { handlePostMessages(c) })
	router.GET("/v2/message", func(c *gin.Context) { handleGetMessagesV2(c) })

	router.GET("/v1/admin/stats", func(c *gin.Context) { handleGetAdminStats(c) })

	return router
}

func parseFilterParams(c *gin.Context) map[string]interface{} {

	var filterParams = map[string]interface{}{}

	before, beforeExists := c.GetQuery("before")

	if beforeExists && before != "" {
		num, err := strconv.Atoi(before)
		if err == nil {
			filterParams["before"] = num
		}
	}

	after, afterExists := c.GetQuery("after")

	if afterExists && after != "" {
		num, err := strconv.Atoi(after)
		if err == nil {
			filterParams["after"] = num
		}
	}

	service, serviceExists := c.GetQuery("service")

	if serviceExists && service != "" {
		filterParams["service"] = service
	}

	contains, containsExists := c.GetQuery("contains")

	if containsExists && contains != "" {
		filterParams["contains"] = contains
	}

	return filterParams
}

func createQueryDslAsString(
	format elasticsearch.QueryFormat,
	params map[string]interface{},
) string {
	// fmt.Printf("%d\n", len(params))

	must := []map[string]interface{}{}

	if params["service"] != nil {
		must = append(must, map[string]interface{}{
			"match": map[string]interface{}{
				"service": params["service"],
			},
		})
	}

	if params["contains"] != nil {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  params["contains"],
				"fields": []string{"address", "message", "service", "serverity"},
			},
		})
	}

	if params["after"] != nil || params["before"] != nil {
		rangeParams := map[string]int{}

		if params["after"] != nil {
			rangeParams["gte"] = params["after"].(int)
		}

		if params["before"] != nil {
			rangeParams["lte"] = params["before"].(int)
		}

		must = append(must, map[string]interface{}{
			"range": map[string]interface{}{
				"stamp": rangeParams,
			},
		})
	}

	dsl := map[string]interface{}{
		"query": map[string]interface{}{
			"filtered": map[string]interface{}{
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": must,
					},
				},
			},
		},
		"size": format.Size,
		"from": format.From,
	}

	var sortOrder string

	if format.Order {
		sortOrder = "desc"
	} else {
		sortOrder = "asc"
	}

	dsl["sort"] = map[string]string{
		format.Key: sortOrder,
	}

	output, _ := json.Marshal(dsl)
	return string(output)
}
