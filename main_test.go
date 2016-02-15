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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	piazza "github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-logger/client"
	"github.com/venicegeo/pz-logger/server"
	"log"
	"runtime"
	"testing"
	"time"
)

type LoggerTester struct {
	suite.Suite

	logger client.ILoggerService
}

func X() {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic(1)
	}
	f := runtime.FuncForPC(pc)
	log.Printf(">>>>>>>>>>>>>>>>>>>>>>>> %s", f.Name())
}

func (suite *LoggerTester) SetupSuite() {
	//t := suite.T()
	X()
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
		if actualMssgs[i] != expectedMssgs[i] {
			assert.Equal(t, expectedMssgs[i], actualMssgs[i], "message %d not equal", i)
		}
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
	err = logger.PostToMessages(&data1)
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

	err = logger.PostToMessages(&data2)
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

	err = logger.Log("mocktest", "0.0.0.0", client.SeverityInfo, "message from logger unit test via piazza.Log()", time.Now())
	assert.NoError(err, "pzService.Log()")

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
