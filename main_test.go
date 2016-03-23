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

package main

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	piazza "github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-logger/client"
	"github.com/venicegeo/pz-logger/server"
)

type LoggerTester struct {
	suite.Suite

	logger client.ILoggerService
}

func (suite *LoggerTester) SetupSuite() {
	config, err := piazza.NewConfig(piazza.PzLogger, piazza.ConfigModeTest)
	if err != nil {
		log.Fatal(err)
	}

	sys, err := piazza.NewSystem(config)
	if err != nil {
		log.Fatal(err)
	}

	_ = sys.StartServer(server.CreateHandlers(sys))

	suite.logger, err = client.NewPzLoggerService(sys, sys.Config.GetBindToAddress())
	if err != nil {
		log.Fatal(err)
	}
}

func (suite *LoggerTester) TearDownSuite() {
	//TODO: kill the go routine running the server
}

func TestRunSuite(t *testing.T) {
	s := new(LoggerTester)
	suite.Run(t, s)
}

func checkMessageArrays(t *testing.T, actualMssgs []client.LogMessage, expectedMssgs []client.LogMessage) {
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

func (suite *LoggerTester) TestOkay() {
	t := suite.T()

	logger := suite.logger

	var err error
	var actualMssgs []client.LogMessage
	var expectedMssgs []client.LogMessage

	assert := assert.New(t)

	////

	data1 := client.LogMessage{
		Service:  "log-tester",
		Address:  "128.1.2.3",
		Time:     time.Now(),
		Severity: "Info",
		Message:  "The quick brown fox",
	}
	err = logger.LogMessage(&data1)
	assert.NoError(err, "PostToMessages")

	actualMssgs, err = logger.GetFromMessages()
	assert.NoError(err, "GetFromMessages")

	expectedMssgs = []client.LogMessage{data1}
	checkMessageArrays(t, actualMssgs, expectedMssgs)

	////

	data2 := client.LogMessage{
		Service:  "log-tester",
		Address:  "128.0.0.0",
		Time:     time.Now(),
		Severity: "Fatal",
		Message:  "The quick brown fox",
	}

	err = logger.LogMessage(&data2)
	assert.NoError(err, "PostToMessages")

	actualMssgs, err = logger.GetFromMessages()
	assert.NoError(err, "GetFromMessages")

	expectedMssgs = []client.LogMessage{data1, data2}
	checkMessageArrays(t, actualMssgs, expectedMssgs)

	stats, err := logger.GetFromAdminStats()
	assert.NoError(err, "GetFromAdminStats")
	assert.Equal(2, stats.NumMessages, "stats check")
	assert.WithinDuration(time.Now(), stats.StartTime, 10*time.Second, "service start time too long ago")

	////

	err = logger.Log("mocktest", "0.0.0.0", client.SeverityInfo, time.Now(), "message from logger unit test via piazza.Log()")
	assert.NoError(err, "pzService.Log()")

	////

	clogger := client.NewCustomLogger(&logger, "TesingService", "123 Main St.")
	err = clogger.Debug("a %s message", "DEBUG")
	assert.NoError(err)
	err = clogger.Info("a %s message", "INFO")
	assert.NoError(err)
	err = clogger.Warn("a %s message", "WARN")
	assert.NoError(err)
	err = clogger.Error("a %s message", "ERROR")
	assert.NoError(err)
	err = clogger.Fatal("a %s message", "FATAL")
	assert.NoError(err)
	////

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
