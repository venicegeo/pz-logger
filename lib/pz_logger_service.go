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
	"net/http"
	"time"

	piazza "github.com/venicegeo/pz-gocommon"
)

type PzLoggerService struct {
	url            string
	serviceName    piazza.ServiceName
	serviceAddress string
}

func NewPzLoggerService(sys *piazza.SystemConfig) (*PzLoggerService, error) {
	var _ IClient = new(PzLoggerService)

	var err error

	url, err := sys.GetURL(piazza.PzLogger)
	if err != nil {
		return nil, err
	}

	service := &PzLoggerService{
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

func (c *PzLoggerService) GetFromMessages() ([]Message, error) {

	resp, err := http.Get(c.url + "/messages")
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

func (c *PzLoggerService) GetFromAdminStats() (*LoggerAdminStats, error) {

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

func (c *PzLoggerService) GetFromAdminSettings() (*LoggerAdminSettings, error) {

	resp, err := http.Get(c.url + "/admin/settings")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	settings := new(LoggerAdminSettings)
	err = json.Unmarshal(data, settings)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (c *PzLoggerService) PostToAdminSettings(settings *LoggerAdminSettings) error {

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url+"/admin/settings", piazza.ContentTypeJSON, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

///////////////////

func (pz *PzLoggerService) LogMessage(mssg *Message) error {

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
func (pz *PzLoggerService) Log(
	service piazza.ServiceName,
	address string,
	severity Severity,
	t time.Time,
	message string, v ...interface{}) error {

	str := fmt.Sprintf(message, v...)
	mssg := Message{Service: service, Address: address, Severity: severity, Time: t, Message: str}

	return pz.LogMessage(&mssg)
}

func (logger *PzLoggerService) post(severity Severity, message string, v ...interface{}) error {
	str := fmt.Sprintf(message, v...)
	return logger.Log(logger.serviceName, logger.serviceAddress, severity, time.Now(), str)
}

// Debug sends a Debug-level message to the logger.
func (logger *PzLoggerService) Debug(message string, v ...interface{}) error {
	return logger.post(SeverityDebug, message, v...)
}

// Info sends an Info-level message to the logger.
func (logger *PzLoggerService) Info(message string, v ...interface{}) error {
	return logger.post(SeverityInfo, message, v...)
}

// Warn sends a Waring-level message to the logger.
func (logger *PzLoggerService) Warn(message string, v ...interface{}) error {
	return logger.post(SeverityWarning, message, v...)
}

// Error sends a Error-level message to the logger.
func (logger *PzLoggerService) Error(message string, v ...interface{}) error {
	return logger.post(SeverityError, message, v...)
}

// Fatal sends a Fatal-level message to the logger.
func (logger *PzLoggerService) Fatal(message string, v ...interface{}) error {
	return logger.post(SeverityFatal, message, v...)
}
