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
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	piazza "github.com/venicegeo/pz-gocommon/gocommon"
)

const MOCKING = true

func sleep() {
	time.Sleep(1 * time.Second)
}

type LoggerTester struct {
	suite.Suite

	esi    elasticsearch.IIndex
	sys    *piazza.SystemConfig
	client IClient

	genericServer *piazza.GenericServer
}

func (suite *LoggerTester) setupFixture() {
	t := suite.T()
	assert := assert.New(t)

	var required []piazza.ServiceName
	if MOCKING {
		required = []piazza.ServiceName{}
	} else {
		required = []piazza.ServiceName{piazza.PzElasticSearch}
	}
	sys, err := piazza.NewSystemConfig(piazza.PzLogger, required)
	assert.NoError(err)
	suite.sys = sys

	esi, err := elasticsearch.NewIndexInterface(sys, "loggertest$", "", MOCKING)
	assert.NoError(err)
	suite.esi = esi

	if MOCKING {
		client, err := NewMockClient(sys)
		assert.NoError(err)
		suite.client = client
	} else {
		client, err := NewClient(sys)
		assert.NoError(err)
		suite.client = client
	}

	service := &Service{}
	err = service.Init(sys, esi)
	assert.NoError(err)

	server := &Server{}
	server.Init(service)

	suite.genericServer = &piazza.GenericServer{Sys: sys}

	err = suite.genericServer.Configure(server.Routes)
	if err != nil {
		log.Fatal(err)
	}

	_, err = suite.genericServer.Start()
	assert.NoError(err)
}

func (suite *LoggerTester) teardownFixture() {
	suite.genericServer.Stop()

	suite.esi.Close()
	suite.esi.Delete()
}

func TestRunSuite(t *testing.T) {
	s := &LoggerTester{}
	suite.Run(t, s)
}

func (suite *LoggerTester) getLastMessage() string {
	t := suite.T()
	assert := assert.New(t)

	client := suite.client

	format := piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   piazza.PaginationOrderAscending,
		SortBy:  "createdOn",
	}
	ms, count, err := client.GetMessages(&format, nil)
	assert.NoError(err)
	assert.True(len(ms) > 0)
	assert.True(count >= len(ms))

	return ms[len(ms)-1].String()
}

func (suite *LoggerTester) Test00Time() {
	t := suite.T()
	assert := assert.New(t)

	a := "2006-01-02T15:04:05+07:00"
	b, err := time.Parse(time.RFC3339, a)
	assert.NoError(err)
	c := b.Format(time.RFC3339)
	assert.EqualValues(a, c)
}

func (suite *LoggerTester) Test01Elasticsearch() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	version := suite.esi.GetVersion()
	assert.Contains("2.2.0", version)
}

func (suite *LoggerTester) Test02One() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	var err error

	data1 := Message{
		Service:   "log-tester",
		Address:   "128.1.2.3",
		CreatedOn: time.Now(),
		Severity:  "Info",
		Message:   "The quick brown fox",
	}

	data2 := Message{
		Service:   "log-tester",
		Address:   "128.0.0.0",
		CreatedOn: time.Now(),
		Severity:  "Fatal",
		Message:   "The quick brown fox",
	}

	{
		err = client.PostMessage(&data1)
		assert.NoError(err, "PostToMessages")
	}

	sleep()

	{
		actualMssg := suite.getLastMessage()
		expectedMssg := data1.String()
		assert.EqualValues(actualMssg, expectedMssg)
	}

	{
		err = client.PostMessage(&data2)
		assert.NoError(err, "PostToMessages")
	}

	sleep()

	{
		actualMssg := suite.getLastMessage()
		expectedMssg := data2.String()
		assert.EqualValues(actualMssg, expectedMssg)
	}

	{
		stats, err := client.GetStats()
		assert.NoError(err, "GetFromAdminStats")
		assert.Equal(2, stats.NumMessages, "stats check")
	}
}

func (suite *LoggerTester) Test03Help() {
	t := suite.T()
	assert := assert.New(t)

	err := suite.client.PostLog("mocktest", "0.0.0.0", SeverityInfo, time.Now(), "message from logger unit test via piazza.Log()")
	assert.NoError(err, "pzService.Log()")
}

func (suite *LoggerTester) Test04Admin() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	_, err := client.GetStats()
	assert.NoError(err, "GetFromAdminStats")
}

func (suite *LoggerTester) Test05Pagination() {
	t := suite.T()
	assert := assert.New(t)

	if MOCKING {
		//t.Skip("Skipping test, because mocking.")
	}

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	client.SetService("myservice", "1.2.3.4")

	err := client.Debug("d")
	assert.NoError(err)
	sleep()
	err = client.Info("i")
	assert.NoError(err)
	sleep()
	err = client.Warn("w")
	assert.NoError(err)
	sleep()
	err = client.Error("e")
	assert.NoError(err)
	sleep()
	err = client.Fatal("f")
	assert.NoError(err)
	sleep()

	format := piazza.JsonPagination{
		PerPage: 1,
		Page:    0,
		SortBy:  "createdOn",
		Order:   piazza.PaginationOrderDescending,
	}
	ms, count, err := client.GetMessages(&format, nil)
	assert.NoError(err)
	assert.Len(ms, 1)
	assert.EqualValues(SeverityDebug, ms[0].Severity)
	assert.Equal(5, count)

	format = piazza.JsonPagination{
		PerPage: 5,
		Page:    0,
		SortBy:  "createdOn",
		Order:   piazza.PaginationOrderAscending,
	}
	ms, count, err = client.GetMessages(&format, nil)
	assert.NoError(err)
	assert.Len(ms, 5)
	assert.EqualValues(SeverityFatal, ms[4].Severity)
	assert.Equal(5, count)

	format = piazza.JsonPagination{
		PerPage: 3,
		Page:    1,
		SortBy:  "createdOn",
		Order:   piazza.PaginationOrderDescending,
	}
	ms, count, err = client.GetMessages(&format, nil)
	assert.NoError(err)
	assert.Len(ms, 2)

	assert.EqualValues(SeverityError, ms[1].Severity)
	assert.EqualValues(SeverityFatal, ms[0].Severity)
	assert.Equal(5, count)
}

func (suite *LoggerTester) Test06OtherParams() {
	t := suite.T()
	assert := assert.New(t)

	if MOCKING {
		t.Skip("Skipping test, because mocking.")
	}

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	client.SetService("myservice", "1.2.3.4")

	sometime := time.Now()

	var testData = []Message{
		{
			Address:   "gnemud7smfv/10.254.0.66",
			Message:   "Received Message to Relay on topic Request-Job with key f3b63085-b482-4ae8-8297-3c7d1fcfff7d",
			Service:   "Dispatcher",
			Severity:  "Info",
			CreatedOn: sometime,
		}, {
			Address:   "gnfbnqsn5m9/10.254.0.14",
			Message:   "Processed Update Status for Job 6d0ea538-4382-4ea5-9669-56519b8c8f58 with Status Success",
			Service:   "JobManager",
			Severity:  "Info",
			CreatedOn: sometime,
		}, {
			Address:   "0.0.0.0",
			Message:   "generated 1: 09d4ec60-ea61-4066-8697-5568a47f84bf",
			Service:   "pz-uuidgen",
			Severity:  "Info",
			CreatedOn: sometime,
		}, {
			Address:   "gnfbnqsn5m9/10.254.0.14",
			Message:   "Handling Job with Topic Create-Job for Job ID 09d4ec60-ea61-4066-8697-5568a47f84bf",
			Service:   "JobManager",
			Severity:  "Info",
			CreatedOn: sometime,
		}, {
			Address:   "gnfbnqsn5m9/10.254.0.14",
			Message:   "Handling Job with Topic Update-Job for Job ID be4ce034-1187-4a4f-95a9-a9c31826248b",
			Service:   "JobManager",
			Severity:  "Info",
			CreatedOn: sometime,
		},
	}

	for _, e := range testData {
		err := client.PostMessage(&e)
		assert.NoError(err)
	}

	sleep()

	format := piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   piazza.PaginationOrderDescending,
		SortBy:  "createdOn",
	}

	params := piazza.HttpQueryParams{}
	params.AddString("service", "JobManager")
	params.AddString("contains", "Success")

	msgs, count, err := client.GetMessages(&format, &params)
	assert.NoError(err)
	assert.Len(msgs, 1)
	assert.Equal(5, count)
}

func (suite *LoggerTester) TestConstructDsl() {
	t := suite.T()
	assert := assert.New(t)

	format := &piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   piazza.PaginationOrderDescending,
		SortBy:  "createdOn",
	}

	startS := "2016-07-26T01:00:00.000Z"
	endS := "2016-07-26T02:00:00.000Z"

	startT, err := time.Parse(time.RFC3339, startS)
	assert.NoError(err)
	endT, err := time.Parse(time.RFC3339, endS)
	assert.NoError(err)

	params := &piazza.HttpQueryParams{}
	params.AddTime("before", startT)
	params.AddTime("after", endT)

	actual, err := createQueryDslAsString(format, params)
	assert.NoError(err)
	assert.NotEmpty(actual)

	expected := `
	{
		"from":0,
		"query": {
			"filtered": {
				"query": {
					"bool": { 
						"must": [
							{
								"range": {
									"createdOn": {
										"gte":"2016-07-26T02:00:00Z",
										"lte":"2016-07-26T01:00:00Z"
									}
								}
							}
						]
					}
				}
			}
		},
		"size":100, 
		"sort":{"createdOn":"desc"}
	}`
	assert.JSONEq(expected, actual)
}
