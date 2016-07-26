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
	"fmt"
	"log"
	"time"

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
)

// implements Logger
type MockClient struct {
	serviceName    piazza.ServiceName
	serviceAddress string
	messages       []Message
	stats          LoggerAdminStats
}

func NewMockClient(sys *piazza.SystemConfig) (IClient, error) {
	var _ IClient = new(MockClient)

	service := &MockClient{}

	service.stats.CreatedOn = time.Now()

	return service, nil
}

func (logger *MockClient) GetMessages(
	format *piazza.JsonPagination,
	params *piazza.HttpQueryParams) ([]Message, int, error) {

	if params.ToParamString() != "" {
		log.Fatalf("parameters are not supported in mock client")
	}

	startIndex := format.Page * format.PerPage
	endIndex := startIndex + format.PerPage
	totalCount := len(logger.messages)

	if format.SortBy != "createdOn" {
		log.Fatalf("unsupported sort key in mock client: %s", format.SortBy)
	}

	if startIndex > totalCount {
		m := make([]Message, 0)
		return m, totalCount, nil
	}

	if endIndex > totalCount {
		// clip!
		endIndex = totalCount
	}

	resultCount := endIndex - startIndex

	//log.Printf("====")
	//log.Printf("Size=%d, From=%d", format.Size, format.From)
	//log.Printf("StartIndex=%d, EndIndex=%d, ResultCount=%d", startIndex, endIndex, resultCount)
	//for i, v := range logger.messages {
	//log.Printf("%d: %s", i, v)
	//}

	// we return exactly one page
	// first we get the right page, *then* we sort that subset

	buf := make([]Message, resultCount)
	for i := 0; i < resultCount; i++ {
		//log.Printf("== %d %d", i, startIndex+i)
		buf[i] = logger.messages[startIndex+i]
	}

	if format.Order == piazza.PaginationOrderDescending {
		buf2 := make([]Message, resultCount)
		for i := 0; i < resultCount; i++ {
			buf2[i] = buf[resultCount-1-i]
		}
		buf = buf2
	}

	//log.Printf("----")
	//for i, v := range buf {
	//log.Printf("%d: %s", i, v)
	//}

	//log.Printf("====")

	return buf, totalCount, nil
}

func (logger *MockClient) GetStats() (*LoggerAdminStats, error) {
	return &logger.stats, nil
}

func (logger *MockClient) PostMessage(mssg *Message) error {
	logger.messages = append(logger.messages, *mssg)
	logger.stats.NumMessages++
	return nil
}

func (mock *MockClient) PostLog(service piazza.ServiceName, address string, severity Severity, t time.Time, message string, v ...interface{}) error {
	mssg := Message{Service: service, Address: address, Severity: severity, Message: message, CreatedOn: t}
	return mock.PostMessage(&mssg)
}

func (logger *MockClient) SetService(name piazza.ServiceName, address string) {
	logger.serviceName = name
	logger.serviceAddress = address
}

func (logger *MockClient) post(severity Severity, message string, v ...interface{}) error {
	str := fmt.Sprintf(message, v...)
	return logger.PostLog(logger.serviceName, logger.serviceAddress, severity, time.Now(), str)
}

// Debug sends a Debug-level message to the logger.
func (logger *MockClient) Debug(message string, v ...interface{}) error {
	return logger.post(SeverityDebug, message, v...)
}

// Info sends an Info-level message to the logger.
func (logger *MockClient) Info(message string, v ...interface{}) error {
	return logger.post(SeverityInfo, message, v...)
}

// Warn sends a Waring-level message to the logger.
func (logger *MockClient) Warn(message string, v ...interface{}) error {
	return logger.post(SeverityWarning, message, v...)
}

// Error sends a Error-level message to the logger.
func (logger *MockClient) Error(message string, v ...interface{}) error {
	return logger.post(SeverityError, message, v...)
}

// Fatal sends a Fatal-level message to the logger.
func (logger *MockClient) Fatal(message string, v ...interface{}) error {
	return logger.post(SeverityFatal, message, v...)
}
