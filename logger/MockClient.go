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

// MockClient implements Logger
type MockClient struct {
	serviceName    piazza.ServiceName
	serviceAddress string
	messages       []Message
	stats          Stats
}

func NewMockClient() (IClient, error) {
	var _ IClient = new(MockClient)

	service := &MockClient{}

	service.stats.CreatedOn = time.Now()

	return service, nil
}

func (c *MockClient) GetVersion() (*piazza.Version, error) {
	version := piazza.Version{Version: Version}
	return &version, nil
}

func (c *MockClient) GetMessages(
	format *piazza.JsonPagination,
	params *piazza.HttpQueryParams) ([]Message, int, error) {

	if params.String() != "" {
		log.Fatalf("parameters are not supported in mock client")
	}

	startIndex := format.Page * format.PerPage
	endIndex := startIndex + format.PerPage
	totalCount := len(c.messages)

	if format.SortBy != "createdOn" {
		log.Fatalf("unsupported sort key in mock client: %s", format.SortBy)
	}

	if startIndex > totalCount {
		return []Message{}, totalCount, nil
	}

	if endIndex > totalCount {
		// clip!
		endIndex = totalCount
	}

	resultCount := endIndex - startIndex

	// we return exactly one page
	// first we get the right page, *then* we sort that subset

	buf := make([]Message, resultCount)
	for i := 0; i < resultCount; i++ {
		buf[i] = c.messages[startIndex+i]
	}

	if format.Order == piazza.SortOrderDescending {
		buf2 := make([]Message, resultCount)
		for i := 0; i < resultCount; i++ {
			buf2[i] = buf[resultCount-1-i]
		}
		buf = buf2
	}

	return buf, totalCount, nil
}

func (c *MockClient) GetStats() (*Stats, error) {
	return &c.stats, nil
}

func (c *MockClient) PostMessage(mssg *Message) error {
	c.messages = append(c.messages, *mssg)
	c.stats.NumMessages++
	return nil
}

func (c *MockClient) PostLog(service piazza.ServiceName, address string, severity Severity, t time.Time, message string, v ...interface{}) error {
	mssg := Message{Service: service, Address: address, Severity: severity, Message: message, CreatedOn: t}
	return c.PostMessage(&mssg)
}

func (c *MockClient) SetService(name piazza.ServiceName, address string) {
	c.serviceName = name
	c.serviceAddress = address
}

func (c *MockClient) post(severity Severity, message string, v ...interface{}) error {
	str := fmt.Sprintf(message, v...)
	return c.PostLog(c.serviceName, c.serviceAddress, severity, time.Now(), str)
}

// Debug sends a Debug-level message to the logger.
func (c *MockClient) Debug(message string, v ...interface{}) {
	err := c.post(SeverityDebug, message, v...)
	if err != nil {
		panic(err)
	}
}

// Info sends an Info-level message to the logger.
func (c *MockClient) Info(message string, v ...interface{}) {
	err := c.post(SeverityInfo, message, v...)
	if err != nil {
		panic(err)
	}
}

// Warn sends a Waring-level message to the logger.
func (c *MockClient) Warn(message string, v ...interface{}) {
	err := c.post(SeverityWarning, message, v...)
	if err != nil {
		panic(err)
	}
}

// Error sends a Error-level message to the logger.
func (c *MockClient) Error(message string, v ...interface{}) {
	err := c.post(SeverityError, message, v...)
	if err != nil {
		panic(err)
	}
}

// Fatal sends a Fatal-level message to the logger.
func (c *MockClient) Fatal(message string, v ...interface{}) {
	err := c.post(SeverityFatal, message, v...)
	if err != nil {
		panic(err)
	}
}
