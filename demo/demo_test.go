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

package logger_demo

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/gocommon"
	"github.com/venicegeo/pz-logger/logger"
)

func sleep() {
	time.Sleep(1 * time.Second)
}

type LoggerTester struct {
	suite.Suite
	client *logger.Client
}

func (suite *LoggerTester) setupFixture() {
	t := suite.T()
	assert := assert.New(t)

	client, err := logger.NewClient2("https://pz-logger.int.geointservices.io")
	assert.NoError(err)
	suite.client = client
}

func (suite *LoggerTester) teardownFixture() {
}

func TestRunSuite(t *testing.T) {
	s := &LoggerTester{}
	suite.Run(t, s)
}

func (suite *LoggerTester) getMessages(key string) []logger.Message {
	t := suite.T()
	assert := assert.New(t)

	client := suite.client

	format := &piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   piazza.PaginationOrderAscending,
		SortBy:  "createdOn",
	}
	ms, _, err := client.GetMessages(format, nil)
	assert.NoError(err)
	assert.True(len(ms) > 0)

	ret := make([]logger.Message, 0)
	for _, m := range ms {
		if m.Message == key {
			ret = append(ret, m)
		}
	}

	return ret
}

func (suite *LoggerTester) verifyMessageExists(expected *logger.Message) bool {
	t := suite.T()
	assert := assert.New(t)

	client := suite.client

	format := &piazza.JsonPagination{
		PerPage: 10,
		Page:    0,
		Order:   piazza.PaginationOrderAscending,
		SortBy:  "createdOn",
	}
	ms, _, err := client.GetMessages(format, nil)
	assert.NoError(err)
	assert.Len(ms, format.PerPage)

	for _, m := range ms {
		if m.String() == expected.String() {
			return true
		}
	}
	return false
}

func (suite *LoggerTester) xTestRawGet() {
	t := suite.T()
	assert := assert.New(t)

	resp, err := http.Get("https://pz-logger.int.geointservices.io/message?perPage=13&page=&0")
	assert.NoError(err)
	assert.True(resp.ContentLength >= 0)
	if resp.ContentLength == -1 {
		assert.FailNow("bonk")
	}
	assert.True(resp.ContentLength > 0)

	raw := make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, raw)
	defer resp.Body.Close()
	if err != nil && err != io.EOF {
		assert.NoError(err)
	}

	//log.Printf("RAW GET: %s", string(raw))
	//err = json.Unmarshal(raw, output)
	//assett.NoError(err)

	assert.Equal(200, resp.StatusCode)
}

func (suite *LoggerTester) xTestRawPost() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	jsn := `
	{
		"address":"XXX",
		"createdOn":"2016-07-22T16:44:58.065583138-04:00",
		"message":"XXX",
		"service":"XXX",
		"severity":"XXX"
	}`
	reader := bytes.NewReader([]byte(jsn))

	resp, err := http.Post("https://pz-logger.int.geointservices.io/message",
		piazza.ContentTypeJSON, reader)
	assert.NoError(err)

	raw := make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, raw)
	defer resp.Body.Close()
	if err != nil && err != io.EOF {
		assert.NoError(err)
	}

	//err = json.Unmarshal(raw, output)
	//assett.NoError(err)

	assert.Equal(200, resp.StatusCode)
}

func (suite *LoggerTester) xTestGet() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	format := &piazza.JsonPagination{
		PerPage: 12,
		Page:    0,
		Order:   piazza.PaginationOrderAscending,
		SortBy:  "createdOn",
	}
	ms, _, err := client.GetMessages(format, nil)
	assert.NoError(err)
	assert.Len(ms, format.PerPage)
}

func (suite *LoggerTester) xTestPost() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	var err error

	key := time.Now().String()

	data := &logger.Message{
		Service:   "log-tester",
		Address:   "128.1.2.3",
		CreatedOn: time.Now(),
		Severity:  "Info",
		Message:   key,
	}

	err = client.PostMessage(data)
	assert.NoError(err, "PostToMessages")

	sleep()

	assert.True(suite.verifyMessageExists(data))
}

func (suite *LoggerTester) xTestAdmin() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	stats, err := client.GetStats()
	assert.NoError(err, "GetFromAdminStats")
	assert.NotZero(stats.NumMessages)
}

func (suite *LoggerTester) xTestPagination() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	format := &piazza.JsonPagination{
		PerPage: 1,
		Page:    0,
		SortBy:  "createdOn",
		Order:   piazza.PaginationOrderAscending,
	}
	params := &piazza.HttpQueryParams{}

	ms, _, err := client.GetMessages(format, params)
	assert.NoError(err)
	assert.Len(ms, format.PerPage)
	//	log.Printf("%#v ===", ms)
}

func (suite *LoggerTester) TestDateRange() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	uniq := "foo-" + time.Now().String() + "-bar"
	client.SetService(piazza.ServiceName(uniq), "1.2.3.4")

	tstart := time.Now()
	client.Debug("D")
	client.Error("E")
	client.Fatal("F")
	tend := time.Now()

	sleep()

	format := &piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   piazza.PaginationOrderDescending,
		SortBy:  "createdOn",
	}

	params := &piazza.HttpQueryParams{}
	params.AddTime("before", tstart)
	params.AddTime("after", tend)

	msgs, cnt, err := client.GetMessages(format, params)

	assert.NoError(err)
	assert.True(cnt >= 3)
	assert.Len(msgs, 3)
}
