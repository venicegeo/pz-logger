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
	"log"
	"time"

	"github.com/venicegeo/pz-gocommon/gocommon"
)

//---------------------------------------------------------------------

type Client struct {
	url            string
	apiKey         string
	serviceName    piazza.ServiceName
	serviceAddress string
	h              piazza.Http
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
		serviceName:    sys.Name,
		serviceAddress: sys.Address,
		h: piazza.Http{
			BaseUrl: url,
			//ApiKey:  apiKey,
			//Preflight:  preflight,
			//Postflight: postflight,
		},
	}

	err = sys.WaitForService(piazza.PzLogger)
	if err != nil {
		return nil, err
	}

	return service, nil
}

func NewClient2(url string, apiKey string) (*Client, error) {
	var _ IClient = new(Client)

	service := &Client{
		url:            url,
		serviceName:    "notset",
		serviceAddress: "0.0.0.0",
		h: piazza.Http{
			BaseUrl: url,
			ApiKey:  apiKey,
			//Preflight:  preflight,
			//Postflight: postflight,
		},
	}

	return service, nil
}

//---------------------------------------------------------------------

func preflight(verb string, url string, obj string) {
	log.Printf("PREFLIGHT.verb: %s", verb)
	log.Printf("PREFLIGHT.url: %s", url)
	log.Printf("PREFLIGHT.obj: %s", obj)
}

func postflight(code int, obj string) {
	log.Printf("POSTFLIGHT.code: %d", code)
	log.Printf("POSTFLIGHT.obj: %s", obj)
}

func (c *Client) GetMessages(
	format *piazza.JsonPagination,
	params *piazza.HttpQueryParams) ([]Message, int, error) {

	formatString := format.ToParamString()
	paramString := params.ToParamString()

	var ext string
	if formatString != "" && paramString != "" {
		ext = "?" + formatString + "&" + paramString
	} else if formatString == "" && paramString != "" {
		ext = "?" + paramString
	} else if formatString != "" && paramString == "" {
		ext = "?" + formatString
	} else if formatString == "" && paramString == "" {
		ext = ""
	} else {
		return nil, 0, errors.New("Internal error: failed to parse query params")
	}

	endpoint := "/message" + ext

	jresp := c.h.PzGet(endpoint)
	if jresp.IsError() {
		return nil, 0, jresp.ToError()
	}

	var mssgs []Message
	err := jresp.ExtractData(&mssgs)
	if err != nil {
		return nil, 0, err
	}

	return mssgs, jresp.Pagination.Count, nil
}

func (c *Client) GetStats() (*LoggerAdminStats, error) {

	jresp := c.h.PzGet("/admin/stats")
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
func (c *Client) PostMessage(mssg *Message) error {

	err := mssg.Validate()
	if err != nil {
		return errors.New("message did not validate")
	}

	jresp := c.h.PzPost("/message", mssg)
	if jresp.IsError() {
		return jresp.ToError()
	}

	return nil
}

// Log sends the components of a LogMessage to the logger.
func (c *Client) PostLog(
	service piazza.ServiceName,
	address string,
	severity Severity,
	t time.Time,
	message string, v ...interface{}) error {

	str := fmt.Sprintf(message, v...)
	mssg := Message{Service: service, Address: address, Severity: severity, CreatedOn: t, Message: str}

	return c.PostMessage(&mssg)
}

func (c *Client) SetService(name piazza.ServiceName, address string) {
	c.serviceName = name
	c.serviceAddress = address
}

func (c *Client) post(severity Severity, message string, v ...interface{}) error {
	str := fmt.Sprintf(message, v...)
	return c.PostLog(c.serviceName, c.serviceAddress, severity, time.Now(), str)
}

// Debug sends a Debug-level message to the logger.
func (c *Client) Debug(message string, v ...interface{}) error {
	return c.post(SeverityDebug, message, v...)
}

// Info sends an Info-level message to the logger.
func (c *Client) Info(message string, v ...interface{}) error {
	return c.post(SeverityInfo, message, v...)
}

// Warn sends a Waring-level message to the logger.
func (c *Client) Warn(message string, v ...interface{}) error {
	return c.post(SeverityWarning, message, v...)
}

// Error sends a Error-level message to the logger.
func (c *Client) Error(message string, v ...interface{}) error {
	return c.post(SeverityError, message, v...)
}

// Fatal sends a Fatal-level message to the logger.
func (c *Client) Fatal(message string, v ...interface{}) error {
	return c.post(SeverityFatal, message, v...)
}
