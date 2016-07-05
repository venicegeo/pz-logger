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

package logger

import (
	"encoding/json"
	_ "fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

type LockedAdminStats struct {
	sync.Mutex
	LoggerAdminStats
}

type LogData struct {
	sync.Mutex
	esIndex elasticsearch.IIndex
	id      int
}

const schema = "LogData"

type LoggerService struct {
	stats   LockedAdminStats
	logData LogData
}

func (logger *LoggerService) Init(sys *piazza.SystemConfig, esIndex elasticsearch.IIndex) {
	var err error

	logger.stats.StartTime = time.Now()

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

	logger.logData.esIndex = esIndex
}

func (logger *LoggerService) GetRoot() *piazza.JsonResponse {
	resp := &piazza.JsonResponse{
		StatusCode: 200,
		Data:       "Hi. I'm pz-logger.",
	}
	return resp
}

func (logger *LoggerService) PostMessage(mssg *Message) *piazza.JsonResponse {
	err := mssg.Validate()
	if err != nil {
		return &piazza.JsonResponse{StatusCode: http.StatusBadRequest, Message: err.Error()}
	}

	log.Printf("PZLOG: %s\n", mssg.String())

	logger.logData.Lock()
	idStr := strconv.Itoa(logger.logData.id)
	logger.logData.id++
	logger.logData.Unlock()
	indexResult, err := logger.logData.esIndex.PostData(schema, idStr, mssg)
	if err != nil {
		return &piazza.JsonResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		}
	}
	if !indexResult.Created {
		resp := &piazza.JsonResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "POST of log data failed",
		}
		return resp
	}

	logger.stats.LoggerAdminStats.NumMessages++

	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Data:       mssg,
	}

	return resp
}

func (logger *LoggerService) GetAdminStats() *piazza.JsonResponse {
	logger.logData.Lock()
	t := logger.stats.LoggerAdminStats
	logger.logData.Unlock()
	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Data:       t,
	}
	return resp
}

func (logger *LoggerService) GetMessages(queryFunc piazza.QueryFunc,
	getQueryFunc piazza.GetQueryFunc) *piazza.JsonResponse {
	var err error

	format, err := elasticsearch.GetFormatParamsV2(queryFunc, 10, 0, "stamp", elasticsearch.SortDescending)
	if err != nil {
		return &piazza.JsonResponse{StatusCode: http.StatusBadRequest, Message: err.Error()}
	}
	filterParams := logger.parseFilterParams(getQueryFunc)

	//log.Printf("size %d, from %d, key %s, format %v",
	//	format.Size, format.From, format.Key, format.Order)

	log.Printf("filterParams: %v\n", filterParams)

	var searchResult *elasticsearch.SearchResult

	if len(filterParams) == 0 {
		searchResult, err = logger.logData.esIndex.FilterByMatchAll(schema, format)
	} else {
		var jsonString = logger.createQueryDslAsString(format, filterParams)
		searchResult, err = logger.logData.esIndex.SearchByJSON(schema, jsonString)
	}

	if err != nil {
		resp := &piazza.JsonResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "query failed: " + err.Error(),
		}
		return resp
	}

	// TODO: unsafe truncation
	count := int(searchResult.TotalHits())
	matched := int(searchResult.NumberMatched())
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
			resp := &piazza.JsonResponse{
				StatusCode: http.StatusInternalServerError,
				Message:    "unmarshall failed: " + err.Error(),
			}
			return resp
		}

		// still needed?
		err = tmp.Validate()
		if err != nil {
			log.Printf("UNABLE TO VALIDATE: %s", string(*hit.Source))
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

	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Data:       bar,
		Pagination: &piazza.JsonPaginationResponse{
			Page:    format.From,
			PerPage: format.Size,
			Count:   matched,
			SortBy:  format.Key,
			Order:   order,
		},
	}

	return resp
}

func (logger *LoggerService) parseFilterParams(getQueryFunc piazza.GetQueryFunc) map[string]interface{} {

	var filterParams = map[string]interface{}{}

	before, beforeExists := getQueryFunc("before")

	if beforeExists && before != "" {
		num, err := strconv.Atoi(before)
		if err == nil {
			filterParams["before"] = num
		}
	}

	after, afterExists := getQueryFunc("after")

	if afterExists && after != "" {
		num, err := strconv.Atoi(after)
		if err == nil {
			filterParams["after"] = num
		}
	}

	service, serviceExists := getQueryFunc("service")

	if serviceExists && service != "" {
		filterParams["service"] = service
	}

	contains, containsExists := getQueryFunc("contains")

	if containsExists && contains != "" {
		filterParams["contains"] = contains
	}

	return filterParams
}

func (logger *LoggerService) createQueryDslAsString(
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