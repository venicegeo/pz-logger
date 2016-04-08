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
type MockLoggerService struct {
}

func NewMockLoggerService(sys *piazza.SystemConfig) (IClient, error) {
	var _ IClient = new(MockLoggerService)

	service := &MockLoggerService{}

	return service, nil
}

func (*MockLoggerService) GetFromMessages() ([]Message, error) {
	return nil, nil
}

func (*MockLoggerService) GetFromAdminStats() (*LoggerAdminStats, error) {
	return &LoggerAdminStats{}, nil
}

func (*MockLoggerService) GetFromAdminSettings() (*LoggerAdminSettings, error) {
	return &LoggerAdminSettings{}, nil
}

func (*MockLoggerService) PostToAdminSettings(*LoggerAdminSettings) error {
	return nil
}

func (*MockLoggerService) LogMessage(mssg *Message) error {
	tim := mssg.Time.Format("Jan _2 15:04:05")
	log.Printf("[[%s, %s, %s, %s]]", tim, mssg.Service, mssg.Severity, mssg.Message)
	return nil
}

func (mock *MockLoggerService) Log(service piazza.ServiceName, address string, severity Severity, t time.Time, message string, v ...interface{}) error {
	mssg := Message{Service: service, Address: address, Severity: severity, Message: message, Time: t}
	return mock.LogMessage(&mssg)
}

func (logger *MockLoggerService) post(severity Severity, message string, v ...interface{}) error {
	str := fmt.Sprintf(message, v...)
	return logger.Log("MockService", "0.0.0.0", severity, time.Now(), str)
}

// Debug sends a Debug-level message to the logger.
func (logger *MockLoggerService) Debug(message string, v ...interface{}) error {
	return logger.post(SeverityDebug, message, v...)
}

// Info sends an Info-level message to the logger.
func (logger *MockLoggerService) Info(message string, v ...interface{}) error {
	return logger.post(SeverityInfo, message, v...)
}

// Warn sends a Waring-level message to the logger.
func (logger *MockLoggerService) Warn(message string, v ...interface{}) error {
	return logger.post(SeverityWarning, message, v...)
}

// Error sends a Error-level message to the logger.
func (logger *MockLoggerService) Error(message string, v ...interface{}) error {
	return logger.post(SeverityError, message, v...)
}

// Fatal sends a Fatal-level message to the logger.
func (logger *MockLoggerService) Fatal(message string, v ...interface{}) error {
	return logger.post(SeverityFatal, message, v...)
}
