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
	"errors"
	"fmt"
	"time"

	piazza "github.com/venicegeo/pz-gocommon"
)

// LogMessage represents the contents of a messge for the logger service.
// All fields are required.
type LogMessage struct {
	Service  piazza.ServiceName `json:"service"`
	Address  string             `json:"address"`
	Time     time.Time          `json:"time"`
	Severity Severity           `json:"severity"`
	Message  string             `json:"message"`
}

type ILoggerService interface {
	GetName() piazza.ServiceName
	GetAddress() string

	// low-level interfaces
	GetFromMessages() ([]LogMessage, error)
	GetFromAdminStats() (*LoggerAdminStats, error)
	GetFromAdminSettings() (*LoggerAdminSettings, error)
	PostToAdminSettings(*LoggerAdminSettings) error

	// high-level interfaces
	LogMessage(mssg *LogMessage) error
	Log(service piazza.ServiceName, address string, severity Severity, t time.Time, message string, v ...interface{}) error
}

//---------------------------------------------------------------------------

// CustomLogger is for convenience, allowing the logger user to avoid passing all the other params.
type CustomLogger struct {
	iLogger       *ILoggerService
	myServiceName piazza.ServiceName
	myAddress     string
}

func NewCustomLogger(logger *ILoggerService, serviceName piazza.ServiceName, address string) *CustomLogger {
	return &CustomLogger{iLogger: logger, myServiceName: serviceName, myAddress: address}
}

func (logger *CustomLogger) post(severity Severity, message string, v ...interface{}) error {
	str := fmt.Sprintf(message, v...)
	return (*logger.iLogger).Log(logger.myServiceName, logger.myAddress, severity, time.Now(), str)
}

// Debug sends a Debug-level message to the logger.
func (logger *CustomLogger) Debug(message string, v ...interface{}) error {
	return logger.post(SeverityDebug, message, v...)
}

// Info sends an Info-level message to the logger.
func (logger *CustomLogger) Info(message string, v ...interface{}) error {
	return logger.post(SeverityInfo, message, v...)
}

// Warn sends a Waring-level message to the logger.
func (logger *CustomLogger) Warn(message string, v ...interface{}) error {
	return logger.post(SeverityWarning, message, v...)
}

// Error sends a Error-level message to the logger.
func (logger *CustomLogger) Error(message string, v ...interface{}) error {
	return logger.post(SeverityError, message, v...)
}

// Fatal sends a Fatal-level message to the logger.
func (logger *CustomLogger) Fatal(message string, v ...interface{}) error {
	return logger.post(SeverityFatal, message, v...)
}

//---------------------------------------------------------------------------

type LoggerAdminStats struct {
	StartTime   time.Time `json:"starttime"`
	NumMessages int       `json:"num_messages"`
}

type LoggerAdminSettings struct {
	Debug bool `json:"debug"`
}

// ToString returns a LogMessage as a formatted string.
func (mssg *LogMessage) String() string {
	s := fmt.Sprintf("[%s, %s, %s, %s, %s]",
		mssg.Service, mssg.Address, mssg.Time, mssg.Severity, mssg.Message)
	return s
}

type Severity string

const (
	// SeverityDebug is for log messages that are only used in development.
	SeverityDebug Severity = "Debug"

	// SeverityInfo is for log messages that are only informative, no action needed.
	SeverityInfo Severity = "Info"

	// SeverityWarning is for log messages that indicate possible problems. Execution continues normally.
	SeverityWarning Severity = "Warning"

	// SeverityError is for log messages that indicate something went wrong. The problem is usually handled and execution continues.
	SeverityError Severity = "Error"

	// SeverityFatal is for log messages that indicate an internal error and the system is likely now unstable. These should never happen.
	SeverityFatal Severity = "Fatal"
)

// Validate checks to make sure a LogMessage is properly filled out. If not, a non-nil error is returned.
func (mssg *LogMessage) Validate() error {
	if mssg.Service == "" {
		return errors.New("required field 'service' not set")
	}
	if mssg.Address == "" {
		return errors.New("required field 'address' not set")
	}
	if mssg.Time.IsZero() {
		return errors.New("required field 'time' not set")
	}
	if mssg.Severity == "" {
		return errors.New("required field 'severity' not set")
	}
	if mssg.Message == "" {
		return errors.New("required field 'message' not set")
	}

	return nil
}
