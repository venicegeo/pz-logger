package client

import (
	piazza "github.com/venicegeo/pz-gocommon"
	"errors"
	"fmt"
	"time"
)

// LogMessage represents the contents of a messge for the logger service.
// All fields are required.
type LogMessage struct {
	Service  piazza.ServiceName    `json:"service"`
	Address  string    `json:"address"`
	Time     time.Time `json:"time"`
	Severity Severity  `json:"severity"`
	Message  string    `json:"message"`
}

type ILoggerService interface {
	GetName() piazza.ServiceName
	GetAddress() string

	// low-level interfaces
	PostToMessages(*LogMessage) error
	GetFromMessages() ([]LogMessage, error)
	GetFromAdminStats() (*LoggerAdminStats, error)
	GetFromAdminSettings() (*LoggerAdminSettings, error)
	PostToAdminSettings(*LoggerAdminSettings) error

	// high-level interfaces
	Log(service piazza.ServiceName, address string, severity Severity, message string, t time.Time) error
}

type LoggerAdminStats struct {
	StartTime   time.Time `json:"starttime"`
	NumMessages int       `json:"num_messages"`
}

type LoggerAdminSettings struct {
	Debug bool `json:"debug"`
}

// ToString returns a LogMessage as a formatted string.
func (mssg *LogMessage) ToString() string {
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
