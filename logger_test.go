package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/venicegeo/pz-gocommon"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

/*type LogMessage struct {
	Service  string `json:"service"`
	Address  string `json:"address"`
	Time     string `json:"time"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}*/

// @TODO: need to automate call to setup() and/or kill thread after each test
func setup(port string, debug bool) {
	s := fmt.Sprintf("-discovery http://localhost:3000 -port %s", port)
	if debug {
		s += " -debug"
	}

	go main2(s)

	time.Sleep(250 * time.Millisecond)
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
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("bad post response: %s: %s", resp.Status, string(data))
	}

	if string(data) != string(expected) {
		t.Logf("Expected: %s\n", string(expected))
		t.Logf("Actual:   %s\n", string(data))
		t.Fatalf("returned log incorrect")
	}
}

func TestOkay(t *testing.T) {
	setup("12341", false)

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

	resp, err := http.Post("http://localhost:12341/log", "application/json", bytes.NewBuffer(jsonData1))
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	checkValidResponse(t, resp)

	resp, err = http.Get("http://localhost:12341/log")
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

	resp, err = http.Post("http://localhost:12341/log", "application/json", bytes.NewBuffer(jsonData2))
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	checkValidResponse(t, resp)

	resp, err = http.Get("http://localhost:12341/log")
	if err != nil {
		t.Fatalf("get failed: %s", err)
	}

	data22 := []piazza.LogMessage{data1, data2}
	jsonData22, err := json.Marshal(data22)
	if err != nil {
		t.Fatalf("marshall failed: %s", err)
	}
	checkValidResponse2(t, resp, jsonData22)

	resp, err = http.Get("http://localhost:12341/log/admin")
	if err != nil {
		t.Fatalf("admin get failed: %s", err)
	}
	checkValidAdminResponse(t, resp)
}
