package main

import (
	"encoding/json"
	"fmt"
	"github.com/venicegeo/pz-gocommon"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

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

func checkValidResponse2(t *testing.T, resp *http.Response, expected string) {
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("bad post response: %s: %s", resp.Status, string(data))
	}

	if string(data) != expected {
		t.Logf("Expected: %s\n", expected)
		t.Logf("Actual:   %s\n", string(data))
		t.Fatalf("returned log incorrect")
	}
}

func TestOkay(t *testing.T) {
	setup("12341", false)

	//var resp *http.Response
	var err error

	data := strings.NewReader(
		`{
	"service": "log-tester",
	"address": "128.1.2.3",
	"time": "2007-04-05T14:30Z",
	"severity": "Info",
	"message": "The quick brown fox"
}`)

	expected := "[log-tester, 128.1.2.3, 2007-04-05T14:30Z, Info, The quick brown fox]\n"

	resp, err := http.Post("http://localhost:12341/log", "application/json", data)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	checkValidResponse(t, resp)

	resp, err = http.Get("http://localhost:12341/log")
	if err != nil {
		t.Fatalf("get failed: %s", err)
	}
	checkValidResponse2(t, resp, expected)

	///////////////////

	data = strings.NewReader(
		`{
	"service": "log-tester",
	"address": "128.0.0.0",
	"time": "2006-04-05T14:30Z",
	"severity": "Fatal",
	"message": "The qiuck brown fox"
}`)

	expected += "[log-tester, 128.0.0.0, 2006-04-05T14:30Z, Fatal, The qiuck brown fox]\n"

	resp, err = http.Post("http://localhost:12341/log", "application/json", data)
	if err != nil {
		t.Fatalf("post failed: %s", err)
	}
	checkValidResponse(t, resp)

	resp, err = http.Get("http://localhost:12341/log")
	if err != nil {
		t.Fatalf("get failed: %s", err)
	}
	checkValidResponse2(t, resp, expected)

	resp, err = http.Get("http://localhost:12341/log/admin")
	if err != nil {
		t.Fatalf("admin get failed: %s", err)
	}
	checkValidAdminResponse(t, resp)
}
