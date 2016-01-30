package main

import (
	"bytes"
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

func checkValidResponse2(t *testing.T, resp *http.Response, expected []byte) {
	var expectedMssgs []piazza.LogMessage
	err := json.Unmarshal(expected, &expectedMssgs)
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("bad post response: %s: %s", resp.Status, string(data))
	}

	var actualMssgs []piazza.LogMessage
	err = json.Unmarshal(data, &actualMssgs)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(actualMssgs) != len(expectedMssgs) {
		t.Fatalf("expected %d mssgs, got %d", len(expectedMssgs), len(actualMssgs))
	}
	for i := 0; i < len(actualMssgs); i++ {
		if actualMssgs[i] != expectedMssgs[i] {
			t.Logf("Expected: %s\n", string(expected))
			t.Logf("Actual:   %s\n", string(data))
			t.Fatalf("returned log incorrect")
		}
	}
}

func (suite *LoggerTester) TestOkay() {
	t := suite.T()

	//var resp *http.Response
	var err error

	data1 := piazza.LogMessage{
		Service:  "log-tester",
		Address:  "128.1.2.3",
		Time:     "2007-04-05T14:30Z",
		Severity: "Info",
		Message:  "The quick brown fox",
	}
	jsonData1, err := json.Marshal(data1)
	if err != nil {
		t.Fatalf("marshall failed: %s", err)
	}

	resp, err := http.Post("http://localhost:12341/v1/messages", "application/json", bytes.NewBuffer(jsonData1))
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	t.Log(string(jsonData1))
	checkValidResponse(t, resp)

	resp, err = http.Get("http://localhost:12341/v1/messages")
	if err != nil {
		t.Fatalf("get failed: %s", err)
	}

	data11 := []piazza.LogMessage{data1}
	jsonData11, err := json.Marshal(data11)
	if err != nil {
		t.Fatalf("marshall failed: %s", err)
	}
	checkValidResponse2(t, resp, jsonData11)

	///////////////////

	data2 := piazza.LogMessage{
		Service:  "log-tester",
		Address:  "128.0.0.0",
		Time:     "2006-04-05T14:30Z",
		Severity: "Fatal",
		Message:  "The quick brown fox",
	}
	jsonData2, err := json.Marshal(data2)
	if err != nil {
		t.Fatalf("marshall failed: %s", err)
	}

	resp, err = http.Post("http://localhost:12341/v1/messages", "application/json", bytes.NewBuffer(jsonData2))
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	checkValidResponse(t, resp)

	resp, err = http.Get("http://localhost:12341/v1/messages")
	if err != nil {
		t.Fatalf("get failed: %s", err)
	}

	data22 := []piazza.LogMessage{data1, data2}
	jsonData22, err := json.Marshal(data22)
	if err != nil {
		t.Fatalf("marshall failed: %s", err)
	}
	checkValidResponse2(t, resp, jsonData22)

	resp, err = http.Get("http://localhost:12341/v1/admin/stats")
	if err != nil {
		t.Fatalf("admin get failed: %s", err)
	}
	checkValidAdminResponse(t, resp)

	err = pzService.Log(piazza.SeverityInfo, "message from pz-logger unit test via piazza.Log()")
	if err != nil {
		t.Fatalf("piazza.Log() failed: %s", err)
	}

	////

	resp, err = http.Get("http://localhost:12341/v1/admin/settings")
	if err != nil {
		t.Fatalf("admin settings get failed: %s", err)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	sm := map[string]string{}
	err = json.Unmarshal(data, &sm)
	if err != nil {
		t.Fatalf("admin settings get failed: %s", err)
	}
	if sm["debug"] != "false" {
		t.Error("settings get had invalid response")
	}

	m := map[string]string{"debug": "true"}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("admin settings %s", err)
	}
	resp, err = http.Post("http://localhost:12341/v1/admin/settings", "application/json", bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("admin settings post failed: %s", err)
	}

	resp, err = http.Get("http://localhost:12341/v1/admin/settings")
	if err != nil {
		t.Fatalf("admin settings get failed: %s", err)
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}
	sm = map[string]string{}
	err = json.Unmarshal(data, &sm)
	if err != nil {
		t.Fatalf("admin settings get failed: %s", err)
	}
	if sm["debug"] != "true" {
		t.Error("settings get had invalid response")
	}
}
