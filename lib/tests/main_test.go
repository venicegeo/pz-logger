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

package tests

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	piazza "github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	pzlogger "github.com/venicegeo/pz-logger/lib"
)

const MOCKING = true

type LoggerTester struct {
	suite.Suite

	esi    elasticsearch.IIndex
	sys    *piazza.SystemConfig
	logger pzlogger.IClient
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

	esi, err := elasticsearch.NewIndexInterface(sys, "loggertest$", MOCKING)
	assert.NoError(err)
	suite.esi = esi

	_ = sys.StartServer(pzlogger.CreateHandlers(sys, esi))

	if MOCKING {
		logger, err := pzlogger.NewMockClient(sys)
		assert.NoError(err)
		suite.logger = logger
	} else {
		logger, err := pzlogger.NewClient(sys)
		assert.NoError(err)
		suite.logger = logger
	}
}

func (suite *LoggerTester) teardownFixture() {
	//TODO: kill the go routine running the server

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

	logger := suite.logger

	format := elasticsearch.QueryFormat{Size: 100, From: 0, Order: elasticsearch.SortDescending, Key: "stamp"}
	ms, err := logger.GetFromMessages(format, map[string]string{} )
	assert.NoError(err)
	assert.True(len(ms) > 0)

	return ms[len(ms)-1].String()
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

	logger := suite.logger

	var err error

	data1 := pzlogger.Message{
		Service:  "log-tester",
		Address:  "128.1.2.3",
		Stamp:    time.Now().Unix(),
		Severity: "Info",
		Message:  "The quick brown fox",
	}

	data2 := pzlogger.Message{
		Service:  "log-tester",
		Address:  "128.0.0.0",
		Stamp:    time.Now().Unix(),
		Severity: "Fatal",
		Message:  "The quick brown fox",
	}

	{
		err = logger.LogMessage(&data1)
		assert.NoError(err, "PostToMessages")
	}

	//	time.Sleep(1 * time.Second)

	{
		actualMssg := suite.getLastMessage()
		expectedMssg := data1.String()
		assert.EqualValues(actualMssg, expectedMssg)
	}

	{
		err = logger.LogMessage(&data2)
		assert.NoError(err, "PostToMessages")
	}

	time.Sleep(4 * time.Second)

	{
		actualMssg := suite.getLastMessage()
		expectedMssg := data2.String()
		assert.EqualValues(actualMssg, expectedMssg)
	}

	{
		stats, err := logger.GetFromAdminStats()
		assert.NoError(err, "GetFromAdminStats")
		assert.Equal(2, stats.NumMessages, "stats check")
		assert.WithinDuration(time.Now(), stats.StartTime, 30*time.Second, "service start time too long ago")
	}
}


func (suite *LoggerTester) Test03Help() {
	t := suite.T()
	assert := assert.New(t)

	err := suite.logger.Log("mocktest", "0.0.0.0", pzlogger.SeverityInfo, time.Now(), "message from logger unit test via piazza.Log()")
	assert.NoError(err, "pzService.Log()")
}


func (suite *LoggerTester) Test04Admin() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	logger := suite.logger

	_, err := logger.GetFromAdminStats()
	assert.NoError(err, "GetFromAdminStats")
}

func (suite *LoggerTester) Test05Pagination() {
	t := suite.T()
	assert := assert.New(t)

	if MOCKING {
		t.Skip("Skipping test, because mocking.")
	}

	suite.setupFixture()
	defer suite.teardownFixture()

	logger := suite.logger

	logger.SetService("myservice", "1.2.3.4")

	err := logger.Debug("d")
	assert.NoError(err)
	time.Sleep(1 * time.Second)
	err = logger.Info("i")
	assert.NoError(err)
	time.Sleep(1 * time.Second)
	err = logger.Warn("w")
	assert.NoError(err)
	time.Sleep(1 * time.Second)
	err = logger.Error("e")
	assert.NoError(err)
	time.Sleep(1 * time.Second)
	err = logger.Fatal("f")
	assert.NoError(err)
	time.Sleep(1 * time.Second)

	format := elasticsearch.QueryFormat{Size: 1, From: 0, Key: "stamp", Order: elasticsearch.SortDescending}
	ms, err := logger.GetFromMessages(format, map[string]string{})
	assert.NoError(err)
	assert.Len(ms, 1)
	assert.EqualValues(pzlogger.SeverityFatal, ms[0].Severity)

	format = elasticsearch.QueryFormat{Size: 3, From: 2, Key: "stamp", Order: elasticsearch.SortDescending}
	ms, err = logger.GetFromMessages(format, map[string]string{})
	assert.NoError(err)
	assert.Len(ms, 3)
	assert.EqualValues(pzlogger.SeverityWarning, ms[0].Severity)
	assert.EqualValues(pzlogger.SeverityInfo, ms[1].Severity)
	assert.EqualValues(pzlogger.SeverityDebug, ms[2].Severity)
}

func (suite *LoggerTester) Test06OtherParams() {
	t := suite.T()
	assert := assert.New(t)

	if MOCKING {
		t.Skip("Skipping test, because mocking.")
	}

	suite.setupFixture()
	defer suite.teardownFixture()

	logger := suite.logger

	logger.SetService("myservice", "1.2.3.4")

	var testData = []pzlogger.Message {
		 {
			Address: "gnemud7smfv/10.254.0.66",
			Message: "Received Message to Relay on topic Request-Job with key f3b63085-b482-4ae8-8297-3c7d1fcfff7d",
			Service: "Dispatcher",
			Severity: "Info",
			Stamp: 1461181461,
		}, {
			Address: "gnfbnqsn5m9/10.254.0.14",
			Message: "Processed Update Status for Job 6d0ea538-4382-4ea5-9669-56519b8c8f58 with Status Success",
			Service: "JobManager",
			Severity: "Info",
			Stamp: 1461181378,
		},  {
			Address: "0.0.0.0",
			Message: "generated 1: 09d4ec60-ea61-4066-8697-5568a47f84bf",
			Service: "pz-uuidgen",
			Severity: "Info",
			Stamp: 1461181362,
		}, {
			Address: "gnfbnqsn5m9/10.254.0.14",
			Message: "Handling Job with Topic Create-Job for Job ID 09d4ec60-ea61-4066-8697-5568a47f84bf",
			Service: "JobManager",
			Severity: "Info",
			Stamp: 1461181362,
		}, {
			Address: "gnfbnqsn5m9/10.254.0.14",
			Message: "Handling Job with Topic Update-Job for Job ID be4ce034-1187-4a4f-95a9-a9c31826248b",
			Service: "JobManager",
			Severity: "Info",
			Stamp: 1461181283,
		},
	}

	for _, e := range testData {
		// log.Printf("%d, %v\n", i, e)
		err := logger.LogMessage(&e)
		assert.NoError(err)
	}
	
	time.Sleep(1 * time.Second)
	
	
	format := elasticsearch.QueryFormat{
		Size: 100, From: 0, 
		Order: elasticsearch.SortDescending, 
		Key: "stamp"}
		
			
	msgs, err := logger.GetFromMessages(format, 
		map[string]string{
			"service": "JobManager",
			"contains": "Success",
		} )	
	assert.NoError(err)
	assert.Len(msgs, 1)
	
	for _, msg := range msgs {
		log.Printf("%v\n", msg)
	}
	
	msgs, err = logger.GetFromMessages(format, 
		map[string]string {
			"before": "1461181461",
			"after": "1461181362",
		} )
	
	assert.NoError(err)
	assert.Len(msgs, 4)
	
	for _, msg := range msgs {
		log.Printf("%v\n", msg)
	}
		
}
