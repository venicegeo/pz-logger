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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	piazza "github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
	pzlogger "github.com/venicegeo/pz-logger/lib"
)

const MOCKING = false

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

	logger, err := pzlogger.NewPzLoggerService(sys)
	assert.NoError(err)
	suite.logger = logger
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

func checkMessageArrays(t *testing.T, actualMssgs []pzlogger.Message, expectedMssgs []pzlogger.Message) {
	assert.Equal(t, len(expectedMssgs), len(actualMssgs), "wrong number of log messages")

	for i := 0; i < len(actualMssgs); i++ {
		assert.EqualValues(t, expectedMssgs[i].Address, actualMssgs[i].Address, "message.address %d not equal", i)
		assert.EqualValues(t, expectedMssgs[i].Message, actualMssgs[i].Message, "message.message %d not equal", i)
		assert.EqualValues(t, expectedMssgs[i].Service, actualMssgs[i].Service, "message.service %d not equal", i)
		assert.EqualValues(t, expectedMssgs[i].Severity, actualMssgs[i].Severity, "message.severity %d not equal", i)
		assert.EqualValues(t, expectedMssgs[i].Time.String(), actualMssgs[i].Time.String(), "message.time %d not equal", i)
		assert.EqualValues(t, expectedMssgs[i].String(), actualMssgs[i].String(), "message.string %d not equal", i)
	}
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
		Time:     time.Now(),
		Severity: "Info",
		Message:  "The quick brown fox",
	}

	data2 := pzlogger.Message{
		Service:  "log-tester",
		Address:  "128.0.0.0",
		Time:     time.Now(),
		Severity: "Fatal",
		Message:  "The quick brown fox",
	}

	{
		err = logger.LogMessage(&data1)
		assert.NoError(err, "PostToMessages")
	}

	//	time.Sleep(1 * time.Second)

	{
		actualMssgs, err := logger.GetFromMessages()
		assert.NoError(err, "GetFromMessages")
		assert.Len(actualMssgs, 1)
		expectedMssgs := []pzlogger.Message{data1}
		checkMessageArrays(t, actualMssgs, expectedMssgs)
	}

	{
		err = logger.LogMessage(&data2)
		assert.NoError(err, "PostToMessages")
	}

	time.Sleep(4 * time.Second)

	{
		actualMssgs, err := logger.GetFromMessages()
		assert.NoError(err, "GetFromMessages")

		expectedMssgs := []pzlogger.Message{data1, data2}
		checkMessageArrays(t, actualMssgs, expectedMssgs)
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

func (suite *LoggerTester) Test04Clogger() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	logger := suite.logger

	clogger := pzlogger.NewCustomLogger(&logger, "TestingService", "123 Main St.")
	err := clogger.Debug("a DEBUG message")
	assert.NoError(err)
	err = clogger.Info("a INFO message")
	assert.NoError(err)
	err = clogger.Warn("a WARN message")
	assert.NoError(err)
	err = clogger.Error("an ERROR message")
	assert.NoError(err)
	err = clogger.Fatal("a FATAL message")
	assert.NoError(err)
}

func (suite *LoggerTester) Test05Admin() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	logger := suite.logger

	settings, err := logger.GetFromAdminSettings()
	assert.NoError(err, "GetFromAdminSettings")
	assert.False(settings.Debug, "settings.Debug")

	settings.Debug = true
	err = logger.PostToAdminSettings(settings)
	assert.NoError(err, "PostToAdminSettings")

	settings, err = logger.GetFromAdminSettings()
	assert.NoError(err, "GetFromAdminSettings")
	assert.True(settings.Debug, "settings.Debug")
}
