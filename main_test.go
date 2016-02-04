package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-logger/client"
	"github.com/venicegeo/pz-logger/server"
	piazza "github.com/venicegeo/pz-gocommon"
	"testing"
	"time"
	"log"
)

type LoggerTester struct {
	suite.Suite
}

func (suite *LoggerTester) SetupSuite() {
	t := suite.T()

	config, err := piazza.GetConfig("pz-logger", true)
	if err != nil {
		log.Fatal(err)
	}

	discover, err := piazza.NewDiscoverClient(config)
	if err != nil {
		log.Fatal(err)
	}

	err = discover.RegisterServiceWithDiscover(config.ServiceName, config.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err = server.RunLoggerServer(config)
		if err != nil {
			log.Fatal(err)
		}
	}()

	err = discover.WaitForService(config.ServiceName, 1000)
	if err != nil {
		t.Fatal(err)
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

	var err error
	var actualMssgs []client.LogMessage
	var expectedMssgs []client.LogMessage

	logger := client.NewPzLoggerClient("localhost:12341")

	assert := assert.New(t)

	////

	data1 := client.LogMessage{
		Service:  "log-tester",
		Address:  "128.1.2.3",
		Time:     "2007-04-05T14:30Z",
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
		Time:     "2006-04-05T14:30Z",
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
	assert.WithinDuration(time.Now(), stats.StartTime, 5*time.Second, "service start time too long ago")

	////

	err = logger.Log(client.SeverityInfo, "message from pz-logger unit test via piazza.Log()")
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
