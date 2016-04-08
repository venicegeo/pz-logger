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
	"fmt"
	"log"
	"time"

	piazza "github.com/venicegeo/pz-gocommon"
)

// implements Logger
type MockClient struct {
	serviceName    piazza.ServiceName
	serviceAddress string
	lastMessage    Message
	stats          LoggerAdminStats
	settings       LoggerAdminSettings
}

func NewMockClient(sys *piazza.SystemConfig) (IClient, error) {
	var _ IClient = new(MockClient)

	service := &MockClient{}

	service.stats.StartTime = time.Now()

	return service, nil
}

func (logger *MockClient) GetFromMessages() ([]Message, error) {
	ms := []Message{logger.lastMessage}
	return ms, nil
}

func (logger *MockClient) GetFromAdminStats() (*LoggerAdminStats, error) {
	return &logger.stats, nil
}

func (logger *MockClient) GetFromAdminSettings() (*LoggerAdminSettings, error) {
	return &logger.settings, nil
}

func (logger *MockClient) PostToAdminSettings(settings *LoggerAdminSettings) error {
	logger.settings = *settings
	return nil
}

func (logger *MockClient) LogMessage(mssg *Message) error {
	//tim := mssg.Time.Format("Jan _2 15:04:05")
	log.Printf("%s", *mssg)
	logger.lastMessage = *mssg
	logger.stats.NumMessages++
	return nil
}

func (mock *MockClient) Log(service piazza.ServiceName, address string, severity Severity, t time.Time, message string, v ...interface{}) error {
	mssg := Message{Service: service, Address: address, Severity: severity, Message: message, Time: t}
	return mock.LogMessage(&mssg)
}

func (logger *MockClient) SetService(name piazza.ServiceName, address string) {
	logger.serviceName = name
	logger.serviceAddress = address
}

func (logger *MockClient) post(severity Severity, message string, v ...interface{}) error {
	str := fmt.Sprintf(message, v...)
	return logger.Log(logger.serviceName, logger.serviceAddress, severity, time.Now(), str)
}

// Debug sends a Debug-level message to the logger.
func (logger *MockClient) Debug(message string, v ...interface{}) error {
	return logger.post(SeverityDebug, message, v...)
}

// Info sends an Info-level message to the logger.
func (logger *MockClient) Info(message string, v ...interface{}) error {
	return logger.post(SeverityInfo, message, v...)
}

// Warn sends a Waring-level message to the logger.
func (logger *MockClient) Warn(message string, v ...interface{}) error {
	return logger.post(SeverityWarning, message, v...)
}

// Error sends a Error-level message to the logger.
func (logger *MockClient) Error(message string, v ...interface{}) error {
	return logger.post(SeverityError, message, v...)
}

// Fatal sends a Fatal-level message to the logger.
func (logger *MockClient) Fatal(message string, v ...interface{}) error {
	return logger.post(SeverityFatal, message, v...)
}