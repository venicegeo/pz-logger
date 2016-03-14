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

package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	piazza "github.com/venicegeo/pz-gocommon"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type PzLoggerService struct {
	url     string
	name    piazza.ServiceName
	address string
}

func NewPzLoggerService(sys *piazza.System, address string) (*PzLoggerService, error) {
	var _ piazza.IService = new(PzLoggerService)
	var _ ILoggerService = new(PzLoggerService)

	var err error

	service := &PzLoggerService{
		url:     fmt.Sprintf("http://%s/v1", address),
		name:    piazza.PzLogger,
		address: address,
	}

	err = sys.WaitForService(service)
	if err != nil {
		return nil, err
	}

	sys.Services[piazza.PzLogger] = service

	return service, nil
}

func (c PzLoggerService) GetName() piazza.ServiceName {
	return c.name
}

func (c PzLoggerService) GetAddress() string {
	return c.address
}

func (c *PzLoggerService) PostToMessages(mssg *LogMessage) error {

	mssgData, err := json.Marshal(mssg)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.url+"/messages", piazza.ContentTypeJSON, bytes.NewBuffer(mssgData))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

func (c *PzLoggerService) GetFromMessages() ([]LogMessage, error) {

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

	var mssgs []LogMessage
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

func (pz *PzLoggerService) LogMessage(mssg *LogMessage) error {

	data, err := json.Marshal(mssg)
	if err != nil {
		log.Printf("pz-logger failed to marshall request: %v", err)
		return err
	}

	resp, err := http.Post(pz.url+"/messages", piazza.ContentTypeJSON, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("pz-logger failed to post request: %v", err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("pz-logger failed to post request: %v", err)
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
	mssg := LogMessage{Service: service, Address: address, Severity: severity, Time: t, Message: str}

	return pz.LogMessage(&mssg)
}
