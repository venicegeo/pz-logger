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

	ms, err := logger.GetFromMessages()
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

func (suite *LoggerTester) Test04ConvenienceFunctions() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	logger := suite.logger

	logger.SetService("myservice", "1.2.3.4")

	expectedPrefix := "[myservice, 1.2.3.4, 201"

	err := logger.Debug("a DEBUG message")
	assert.NoError(err)
	//time.Sleep(3 * time.Second)
	assert.Contains(suite.getLastMessage(), expectedPrefix)
	assert.Contains(suite.getLastMessage(), ", Debug, a DEBUG message]")

	err = logger.Info("an INFO message")
	assert.NoError(err)
	//time.Sleep(3 * time.Second)
	assert.Contains(suite.getLastMessage(), expectedPrefix)
	assert.Contains(suite.getLastMessage(), ", Info, an INFO message]")

	err = logger.Warn("a WARN message")
	assert.NoError(err)
	//time.Sleep(3 * time.Second)
	assert.Contains(suite.getLastMessage(), expectedPrefix)
	assert.Contains(suite.getLastMessage(), ", Warning, a WARN message]")

	err = logger.Error("an ERROR message")
	assert.NoError(err)
	//time.Sleep(3 * time.Second)
	assert.Contains(suite.getLastMessage(), expectedPrefix)
	assert.Contains(suite.getLastMessage(), ", Error, an ERROR message]")

	err = logger.Fatal("a FATAL message")
	assert.NoError(err)
	//time.Sleep(3 * time.Second)
	assert.Contains(suite.getLastMessage(), expectedPrefix)
	assert.Contains(suite.getLastMessage(), ", Fatal, a FATAL message]")
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
