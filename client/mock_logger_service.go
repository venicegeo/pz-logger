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
	"log"
	"time"

	piazza "github.com/venicegeo/pz-gocommon"
)

// implements Logger
type MockLoggerService struct {
	name    piazza.ServiceName
	address string
}

func NewMockLoggerService(sys *piazza.SystemConfig) (ILoggerService, error) {
	var _ piazza.IService = new(MockLoggerService)
	var _ ILoggerService = new(MockLoggerService)

	service := &MockLoggerService{name: "piazza.PzLogger", address: "0.0.0.0"}

	return service, nil
}

func (m *MockLoggerService) GetName() piazza.ServiceName {
	return m.name
}

func (m *MockLoggerService) GetAddress() string {
	return m.address
}

func (*MockLoggerService) GetFromMessages() ([]LogMessage, error) {
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

func (*MockLoggerService) LogMessage(mssg *LogMessage) error {
	tim := mssg.Time.Format("Jan _2 15:04:05")
	log.Printf("[%s, %s, %s, %s]", tim, mssg.Service, mssg.Severity, mssg.Message)
	return nil
}

func (mock *MockLoggerService) Log(service piazza.ServiceName, address string, severity Severity, t time.Time, message string, v ...interface{}) error {
	mssg := LogMessage{Service: service, Address: address, Severity: severity, Message: message, Time: t}
	return mock.LogMessage(&mssg)
}
