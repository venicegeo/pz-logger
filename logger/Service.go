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
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	syslogger "github.com/venicegeo/pz-gocommon/syslog"
)

type Service struct {
	sync.Mutex

	stats  Stats
	origin string

	esIndex elasticsearch.IIndex
	id      int

	writer syslogger.Writer
}

const (
	OldFormat  = "old"
	RfcFormat  = "rfc"
	JsonFormat = "json"
)

func (service *Service) Init(sys *piazza.SystemConfig, esIndex elasticsearch.IIndex) error {
	var err error

	service.stats.CreatedOn = time.Now()

	service.writer, err = NewElasticsearchWriter(esIndex)
	if err != nil {
		return err
	}

	service.esIndex = esIndex

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

func (service *Service) GetStats() *piazza.JsonResponse {
	service.Lock()
	t := service.stats
	service.Unlock()

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
	format, err := params.GetAsString("format", OldFormat)
	if err != nil {
		return service.newBadRequestResponse(err)
	}
	return service.getMessages(params, format)
}

func (service *Service) getMessages(params *piazza.HttpQueryParams, format string) *piazza.JsonResponse {
	var err error

	pagination, err := piazza.NewJsonPagination(params)
	if err != nil {
		return service.newBadRequestResponse(err)
	}

	dsl, err := createQueryDslAsString(pagination, params)
	if err != nil {
		return service.newBadRequestResponse(err)
	}

	var searchResult *elasticsearch.SearchResult

	if dsl == "" {
		searchResult, err = service.esIndex.FilterByMatchAll(logSchema, pagination)
	} else {
		searchResult, err = service.esIndex.SearchByJSON(logSchema, dsl)
	}

	if err != nil {
		return service.newInternalErrorResponse(err)
	}

	var lines = make([]syslogger.Message, 0)

	if searchResult != nil && searchResult.GetHits() != nil {
		for _, hit := range *searchResult.GetHits() {
			if hit.Source == nil {
				log.Printf("null source hit")
				continue
			}

			var msg syslogger.Message
			err = json.Unmarshal(*hit.Source, &msg)
			if err != nil {
				log.Printf("UNABLE TO PARSE: %s", string(*hit.Source))
				return service.newInternalErrorResponse(err)
			}

			// just in case
			err = msg.Validate()
			if err != nil {
				log.Printf("UNABLE TO VALIDATE: %s", string(*hit.Source))
				continue
			}

			lines = append(lines, msg)
		}
	}

	var formattedLines interface{}
	switch format {
	case JsonFormat:
		// do nothing
		formattedLines = lines
	case OldFormat:
		formattedLines, err = toOldFormat(lines)
		if err != nil {
			return service.newInternalErrorResponse(err)
		}
	case RfcFormat:
		formattedLines, err = toRfcFormat(lines)
		if err != nil {
			return service.newInternalErrorResponse(err)
		}
	default:
		return service.newInternalErrorResponse(fmt.Errorf("unrecognized format: %s", format))
	}

	pagination.Count = int(searchResult.TotalHits())
	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
		Data:       formattedLines,
		Pagination: pagination,
	}

	err = resp.SetType()
	if err != nil {
		return service.newInternalErrorResponse(err)
	}
	return resp
}

func toOldFormat(lines []syslogger.Message) ([]Message, error) {
	var lines2 = make([]Message, len(lines))

	for i, newMssg := range lines {
		oldMssg, err := convertNewMessageToOld(&newMssg)
		if err != nil {
			return nil, err
		}
		lines2[i] = *oldMssg
	}

	return lines2, nil
}

func toRfcFormat(lines []syslogger.Message) ([]string, error) {
	var lines2 = make([]string, len(lines))

	for i, newMssg := range lines {
		lines2[i] = newMssg.String()
	}

	return lines2, nil
}

func createQueryDslAsString(
	pagination *piazza.JsonPagination,
	params *piazza.HttpQueryParams) (string, error) {

	must := []map[string]interface{}{}

	service, err := params.GetAsString("service", "")
	if err != nil {
		return "", err
	}

	contains, err := params.GetAsString("contains", "")
	if err != nil {
		return "", err
	}

	before, err := params.GetBefore(time.Time{})
	if err != nil {
		return "", err
	}

	after, err := params.GetAfter(time.Time{})
	if err != nil {
		return "", err
	}

	if service != "" {
		must = append(must, map[string]interface{}{
			"match": map[string]interface{}{
				"service": service,
			},
		})
	}

	if contains != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  contains,
				"fields": []string{"address", "message", "service", "severity"},
			},
		})
	}

	if !after.IsZero() || !before.IsZero() {
		rangeParams := map[string]time.Time{}

		if !after.IsZero() {
			rangeParams["gte"] = after
		}

		if !before.IsZero() {
			rangeParams["lte"] = before
		}

		must = append(must, map[string]interface{}{
			"range": map[string]interface{}{
				"createdOn": rangeParams,
			},
		})
	}

	if len(must) == 0 {
		return "", nil
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
		"size": pagination.PerPage,
		"from": pagination.PerPage * pagination.Page,
	}

	dsl["sort"] = map[string]string{
		pagination.SortBy: string(pagination.Order),
	}

	output, err := json.Marshal(dsl)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (service *Service) PostSyslog(mNew *syslogger.Message) *piazza.JsonResponse {
	err := mNew.Validate()
	if err != nil {
		return service.newBadRequestResponse(err)
	}

	go service.postSyslog(mNew)

	resp := &piazza.JsonResponse{
		StatusCode: http.StatusOK,
	}
	return resp
}

// postSyslog does not return anything. Any errors go to the local log.
func (service *Service) postSyslog(mssgNew *syslogger.Message) error {
	var err error

	service.Lock()
	idStr := strconv.Itoa(service.id)
	service.id++
	service.Unlock()

	mssgOld, err := convertNewMessageToOld(mssgNew)
	if err != nil {
		return err
	}

	_, err = service.esIndex.PostData(logSchema, idStr, mssgOld)
	if err != nil {
		log.Printf("old message post: %s", err.Error())
		// don't return yet, the audit post might still work
	}

	if mssgNew.AuditData != nil {
		_, err = service.esIndex.PostData(auditSchema, idStr, mssgOld)
		if err != nil {
			log.Printf("old message audit post: %s", err.Error())
		}
	}

	return nil
}

func (service *Service) GetSyslog(params *piazza.HttpQueryParams, newStyle bool) *piazza.JsonResponse {
	defaultStyle := JsonFormat
	if !newStyle {
		defaultStyle = OldFormat
	}

	fmt, err := params.GetAsString("format", defaultStyle)
	if err != nil {
		return service.newBadRequestResponse(err)
	}

	return service.getMessages(params, fmt)
}
