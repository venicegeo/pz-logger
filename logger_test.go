package main

import (
	"encoding/json"
	//assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	piazza "github.com/venicegeo/pz-gocommon"
	"io/ioutil"
	"net/http"
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

func checkValidAdminResponse(t *testing.T, resp *http.Response) {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("bad admin response: %s", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}

	var m piazza.AdminResponse
	err = json.Unmarshal(data, &m)
	if err != nil {
		t.Fatalf("unmarshall of admin response: %v", err)
	}

	if time.Since(m.StartTime).Seconds() > 5 {
		t.Fatalf("service start time too long ago")
	}

	if m.Logger == nil {
		t.Fatal("admin response didn't have logger data set")
	}
	if m.Logger.NumMessages != 2 {
		t.Fatalf("wrong number of logs")
	}
}

func checkValidResponse(t *testing.T, resp *http.Response) {
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("bad post response: %s: %s", resp.Status, string(data))
	}
}

func checkValidResponse2(t *testing.T, actualMssgs []piazza.LogMessage, expectedMssgs []piazza.LogMessage) {

	if len(actualMssgs) != len(expectedMssgs) {
		t.Fatalf("expected %d mssgs, got %d", len(expectedMssgs), len(actualMssgs))
	}
	for i := 0; i < len(actualMssgs); i++ {
		if actualMssgs[i] != expectedMssgs[i] {
			t.Logf("Expected[%d]: %v\n", i, expectedMssgs[i])
			t.Logf("Actual[%d]:   %v\n", i, actualMssgs[i])
			t.Fatalf("returned log incorrect")
		}
	}
}

func (suite *LoggerTester) TestOkay() {
	t := suite.T()

	var err error
	var actualMssgs []piazza.LogMessage
	var expectedMssgs []piazza.LogMessage

	client := NewPzLoggerClient("localhost:12341")

	data1 := piazza.LogMessage{
		Service:  "log-tester",
		Address:  "128.1.2.3",
		Time:     "2007-04-05T14:30Z",
		Severity: "Info",
		Message:  "The quick brown fox",
	}
	err = client.PostToMessages(&data1)
	if err != nil {
		t.Fatalf("%s", err)
	}

	actualMssgs, err = client.GetFromMessages()
	if err != nil {
		t.Fatalf("%s", err)
	}

	expectedMssgs = []piazza.LogMessage{data1}
	checkValidResponse2(t, actualMssgs, expectedMssgs)

	///////////////////

	data2 := piazza.LogMessage{
		Service:  "log-tester",
		Address:  "128.0.0.0",
		Time:     "2006-04-05T14:30Z",
		Severity: "Fatal",
		Message:  "The quick brown fox",
	}

	err = client.PostToMessages(&data2)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}

	actualMssgs, err = client.GetFromMessages()
	if err != nil {
		t.Fatalf("get failed: %s", err)
	}

	expectedMssgs = []piazza.LogMessage{data1, data2}
	checkValidResponse2(t, actualMssgs, expectedMssgs)

	stats, err := client.GetFromAdminStats()
	if err != nil {
		t.Fatalf("admin get failed: %s", err)
	}
	if stats.NumMessages != 2 {
		t.Fatalf("stats wrong, expected 3, got %d", stats.NumMessages)
	}

	err = pzService.Log(piazza.SeverityInfo, "message from pz-logger unit test via piazza.Log()")
	if err != nil {
		t.Fatalf("piazza.Log() failed: %s", err)
	}

	////

	settings, err := client.GetFromAdminSettings()
	if err != nil {
		t.Fatalf("admin settings get failed: %s", err)
	}
	if settings.Debug {
		t.Error("settings get had invalid response")
	}

	settings.Debug = true
	err = client.PostToAdminSettings(settings)
	if err != nil {
		t.Fatalf("admin settings post failed: %s", err)
	}

	settings, err = client.GetFromAdminSettings()
	if err != nil {
		t.Fatalf("admin settings get failed: %s", err)
	}
	if !settings.Debug {
		t.Error("settings get had invalid response")
	}
}
