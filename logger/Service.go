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

const schema = "LogData5"

type Service struct {
	stats   LockedAdminStats
	logData LogData
	origin  string
}

func (service *Service) Init(sys *piazza.SystemConfig, esIndex elasticsearch.IIndex) error {
	var err error

	service.stats.CreatedOn = time.Now()
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
		log.Printf("Creating index: %s", esIndex.IndexName())
		err = esIndex.Create("")
		if err != nil {
			log.Fatal(err)
		}
	}

	if !esIndex.TypeExists(schema) {
		log.Printf("Creating type: %s", schema)

		mapping :=
			`{
			"LogData5":{
				"properties":{
					"service":{
						"type": "string",
						"store": true
					},
					"address":{
						"type": "string",
						"store": true
					},
					"createdOn":{
						"type": "date",
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
			log.Printf("LoggerService.Init: %s", err.Error())
			return err
		}
	}

	service.logData.esIndex = esIndex

	service.origin = string(sys.Name)

	return nil
}

func (service *Service) newInternalErrorResponse(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusInternalServerError,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

func (service *Service) newBadRequestResponse(err error) *piazza.JsonResponse {
	return &piazza.JsonResponse{
		StatusCode: http.StatusBadRequest,
		Message:    err.Error(),
		Origin:     service.origin,
	}
}

func (service *Service) GetRoot() *piazza.JsonResponse {
	resp := &piazza.JsonResponse{
		StatusCode: 200,
		Data:       "Hi. I'm pz-logger.",
	}

	err := resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	return resp
}

func (service *Service) PostMessage(mssg *Message) *piazza.JsonResponse {
	err := mssg.Validate()
	if err != nil {
		return newBadRequestResponse(err)
	}

	service.logData.Lock()
	idStr := strconv.Itoa(service.logData.id)
	service.logData.id++
	service.logData.Unlock()

	_, err = service.logData.esIndex.PostData(schema, idStr, mssg)
	if err != nil {
		//log.Printf("POST failed (1): %#v %#v", err, indexResult)
		return service.newInternalErrorResponse(err)
	}

	/*	if !indexResult.Created {
		log.Printf("POST failed (2): %#v", *indexResult)
		return NewInternalErrorResponse(err)
	}*/

	service.stats.LoggerAdminStats.NumMessages++

	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Data:       mssg,
	}

	err = resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	return resp
}

func (service *Service) GetStats() *piazza.JsonResponse {
	service.logData.Lock()
	t := service.stats.LoggerAdminStats
	service.logData.Unlock()

	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Data:       t,
	}

	err := resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	return resp
}

func (service *Service) GetMessage(params *piazza.HttpQueryParams) *piazza.JsonResponse {
	var err error

	var formalPagination *piazza.JsonPagination
	{
		defaults := &piazza.JsonPagination{
			PerPage: 10,
			Page:    0,
			Order:   piazza.PaginationOrderDescending,
			SortBy:  "createdOn",
		}
		formalPagination, err = piazza.NewJsonPagination(params, defaults)
		if err != nil {
			return newBadRequestResponse(err)
		}
	}

	filterParams := service.parseFilterParams(params)

	//log.Printf("size %d, from %d, key %s, format %v",
	//	format.Size, format.From, format.Key, format.Order)

	//log.Printf("filterParams: %v\n", filterParams)

	var searchResult *elasticsearch.SearchResult

	if len(filterParams) == 0 {
		searchResult, err = service.logData.esIndex.FilterByMatchAll(schema, formalPagination)
	} else {
		var jsonString = service.createQueryDslAsString(formalPagination, filterParams)
		searchResult, err = service.logData.esIndex.SearchByJSON(schema, jsonString)
	}

	if err != nil {
		return service.newInternalErrorResponse(err)
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
			return service.newInternalErrorResponse(err)
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

	formalPagination.Count = matched
	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Data:       bar,
		Pagination: formalPagination,
	}

	err = resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	return resp
}

func (service *Service) parseFilterParams(params *piazza.HttpQueryParams) map[string]interface{} {

	var filterParams = map[string]interface{}{}

	before := params.Get("before")

	if before != "" {
		num, err := strconv.Atoi(before)
		if err == nil {
			filterParams["before"] = num
		}
	}

	after := params.Get("after")

	if after != "" {
		num, err := strconv.Atoi(after)
		if err == nil {
			filterParams["after"] = num
		}
	}

	svc := params.Get("service")

	if svc != "" {
		filterParams["service"] = svc
	}

	contains := params.Get("contains")

	if contains != "" {
		filterParams["contains"] = contains
	}

	return filterParams
}

func (service *Service) createQueryDslAsString(
	format *piazza.JsonPagination,
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
				"fields": []string{"address", "message", "service", "severity"},
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
		"size": format.PerPage,
		"from": format.PerPage * format.Page,
	}

	dsl["sort"] = map[string]string{
		format.SortBy: string(format.Order),
	}

	output, _ := json.Marshal(dsl)
	return string(output)
}
