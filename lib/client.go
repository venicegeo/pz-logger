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

package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	piazza "github.com/venicegeo/pz-gocommon"
	"github.com/venicegeo/pz-gocommon/elasticsearch"
)

type Client struct {
	url            string
	serviceName    piazza.ServiceName
	serviceAddress string
}

func NewClient(sys *piazza.SystemConfig) (*Client, error) {
	var _ IClient = new(Client)

	var err error

	url, err := sys.GetURL(piazza.PzLogger)
	if err != nil {
		return nil, err
	}

	service := &Client{
		url:            url,
		serviceName:    "notset",
		serviceAddress: "0.0.0.0",
	}

	err = sys.WaitForService(piazza.PzLogger)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (c *Client) GetFromMessages(format elasticsearch.QueryFormat, params map[string]string) ([]Message, error) {

	url := fmt.Sprintf("%s/messages?size=%d&from=%d&key=%s&order=%t", c.url, format.Size, format.From, format.Key, format.Order)

	var names = []string{"before", "after", "service", "contains"}

	for _, name := range names {
		if value, ok := params[name]; ok {
			//do something here
			url = fmt.Sprintf("%s&%s=%s", url, name, value)
		}
	}

	log.Printf("%s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var mssgs []Message
	err = json.Unmarshal(data, &mssgs)
	if err != nil {
		return nil, err
	}

	return mssgs, nil
}

func (c *Client) GetFromAdminStats() (*LoggerAdminStats, error) {

	resp, err := http.Get(c.url + "/admin/stats")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	stats := new(LoggerAdminStats)
	err = json.Unmarshal(data, stats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

///////////////////

func (pz *Client) LogMessage(mssg *Message) error {

	err := mssg.Validate()
	if err != nil {
		return errors.New("message did not validate")
	}

	data, err := json.Marshal(mssg)
	if err != nil {
		return err
	}

	resp, err := http.Post(pz.url+"/messages", piazza.ContentTypeJSON, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

// Log sends the components of a LogMessage to the logger.
func (pz *Client) Log(
	service piazza.ServiceName,
	address string,
	severity Severity,
	t time.Time,
	message string, v ...interface{}) error {

	var secs int64 = t.Unix()

	str := fmt.Sprintf(message, v...)
	mssg := Message{Service: service, Address: address, Severity: severity, Stamp: secs, Message: str}

	return pz.LogMessage(&mssg)
}

func (logger *Client) SetService(name piazza.ServiceName, address string) {
	logger.serviceName = name
	logger.serviceAddress = address
}

func (logger *Client) post(severity Severity, message string, v ...interface{}) error {
	str := fmt.Sprintf(message, v...)
	return logger.Log(logger.serviceName, logger.serviceAddress, severity, time.Now(), str)
}

// Debug sends a Debug-level message to the logger.
func (logger *Client) Debug(message string, v ...interface{}) error {
	return logger.post(SeverityDebug, message, v...)
}

// Info sends an Info-level message to the logger.
func (logger *Client) Info(message string, v ...interface{}) error {
	return logger.post(SeverityInfo, message, v...)
}

// Warn sends a Waring-level message to the logger.
func (logger *Client) Warn(message string, v ...interface{}) error {
	return logger.post(SeverityWarning, message, v...)
}

// Error sends a Error-level message to the logger.
func (logger *Client) Error(message string, v ...interface{}) error {
	return logger.post(SeverityError, message, v...)
}

// Fatal sends a Fatal-level message to the logger.
func (logger *Client) Fatal(message string, v ...interface{}) error {
	return logger.post(SeverityFatal, message, v...)
}
