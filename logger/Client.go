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

package logger

import (
	"errors"
	"fmt"
	"time"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

//---------------------------------------------------------------------

type Client struct {
	url            string
	serviceName    piazza.ServiceName
	serviceAddress string
}

//---------------------------------------------------------------------

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

//---------------------------------------------------------------------

func (c *Client) GetMessages(format elasticsearch.QueryFormat, params map[string]string) ([]Message, error) {

	url := fmt.Sprintf("%s/message?size=%d&from=%d&key=%s&order=%t", c.url, format.Size, format.From, format.Key, format.Order)

	var names = []string{"before", "after", "service", "contains"}

	for _, name := range names {
		if value, ok := params[name]; ok {
			//do something here
			url = fmt.Sprintf("%s&%s=%s", url, name, value)
		}
	}

	//log.Printf("%s\n", url)

	jresp := piazza.HttpGetJson(url)
	if jresp.IsError() {
		return nil, jresp.ToError()
	}

	var mssgs []Message
	err := jresp.ExtractData(&mssgs)
	if err != nil {
		return nil, err
	}

	return mssgs, nil
}

func (c *Client) GetStats() (*LoggerAdminStats, error) {

	jresp := piazza.HttpGetJson(c.url + "/admin/stats")
	if jresp.IsError() {
		return nil, jresp.ToError()
	}

	stats := &LoggerAdminStats{}
	err := jresp.ExtractData(stats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

//---------------------------------------------------------------------

// LogMessage puts a new message into Elasticsearch.
func (pz *Client) PostMessage(mssg *Message) error {

	err := mssg.Validate()
	if err != nil {
		return errors.New("message did not validate")
	}

	jresp := piazza.HttpPostJson(pz.url+"/message", mssg)
	if jresp.IsError() {
		return jresp.ToError()
	}

	return nil
}

// Log sends the components of a LogMessage to the logger.
func (pz *Client) PostLog(
	service piazza.ServiceName,
	address string,
	severity Severity,
	t time.Time,
	message string, v ...interface{}) error {

	str := fmt.Sprintf(message, v...)
	mssg := Message{Service: service, Address: address, Severity: severity, CreatedOn: t, Message: str}

	return pz.PostMessage(&mssg)
}

func (logger *Client) SetService(name piazza.ServiceName, address string) {
	logger.serviceName = name
	logger.serviceAddress = address
}

func (logger *Client) post(severity Severity, message string, v ...interface{}) error {
	str := fmt.Sprintf(message, v...)
	return logger.PostLog(logger.serviceName, logger.serviceAddress, severity, time.Now(), str)
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
