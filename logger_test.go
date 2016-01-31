package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	piazza "github.com/venicegeo/pz-gocommon"
	"testing"
	"time"
)

type LoggerTester struct {
	suite.Suite
}

func (suite *LoggerTester) SetupSuite() {
	t := suite.T()

	done := make(chan bool, 1)
	go Main(done, true)
	<-done

	err := pzService.WaitForService(pzService.Name, 1000)
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

func checkMessageArrays(t *testing.T, actualMssgs []piazza.LogMessage, expectedMssgs []piazza.LogMessage) {
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
	var actualMssgs []piazza.LogMessage
	var expectedMssgs []piazza.LogMessage

	client := NewPzLoggerClient("localhost:12341")

	assert := assert.New(t)

	////

	data1 := piazza.LogMessage{
		Service:  "log-tester",
		Address:  "128.1.2.3",
		Time:     "2007-04-05T14:30Z",
		Severity: "Info",
		Message:  "The quick brown fox",
	}
	err = client.PostToMessages(&data1)
	assert.NoError(err, "PostToMessages")

	actualMssgs, err = client.GetFromMessages()
	assert.NoError(err, "GetFromMessages")

	expectedMssgs = []piazza.LogMessage{data1}
	checkMessageArrays(t, actualMssgs, expectedMssgs)

	////

	data2 := piazza.LogMessage{
		Service:  "log-tester",
		Address:  "128.0.0.0",
		Time:     "2006-04-05T14:30Z",
		Severity: "Fatal",
		Message:  "The quick brown fox",
	}

	err = client.PostToMessages(&data2)
	assert.NoError(err, "PostToMessages")

	actualMssgs, err = client.GetFromMessages()
	assert.NoError(err, "GetFromMessages")

	expectedMssgs = []piazza.LogMessage{data1, data2}
	checkMessageArrays(t, actualMssgs, expectedMssgs)

	stats, err := client.GetFromAdminStats()
	assert.NoError(err, "GetFromAdminStats")
	assert.Equal(2, stats.NumMessages, "stats check")
	assert.WithinDuration(time.Now(), stats.StartTime, 5*time.Second, "service start time too long ago")

	////

	err = pzService.Log(piazza.SeverityInfo, "message from pz-logger unit test via piazza.Log()")
	assert.NoError(err, "pzService.Log()")

	////

	settings, err := client.GetFromAdminSettings()
	assert.NoError(err, "GetFromAdminSettings")
	assert.False(settings.Debug, "settings.Debug")

	settings.Debug = true
	err = client.PostToAdminSettings(settings)
	assert.NoError(err, "PostToAdminSettings")

	settings, err = client.GetFromAdminSettings()
	assert.NoError(err, "GetFromAdminSettings")
	assert.True(settings.Debug, "settings.Debug")
}
