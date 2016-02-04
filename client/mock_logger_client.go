package client

import (
	"fmt"
	"log"
	"time"
	"github.com/venicegeo/pz-gocommon"
)

// implements Logger
type MockLoggerClient struct{}

func NewMockLoggerClient(sys *piazza.System) (LoggerClient, error) {
	m := MockLoggerClient{}
	return &m, nil
}

func (*MockLoggerClient) PostToMessages(m *LogMessage) error {
	log.Printf("MOCKLOG: %v", m)
	return nil
}

func (*MockLoggerClient) GetFromMessages() ([]LogMessage, error) {
	return nil, nil
}

func (*MockLoggerClient) GetFromAdminStats() (*LoggerAdminStats, error) {
	return &LoggerAdminStats{}, nil
}

func (*MockLoggerClient) GetFromAdminSettings() (*LoggerAdminSettings, error) {
	return &LoggerAdminSettings{}, nil
}

func (*MockLoggerClient) PostToAdminSettings(*LoggerAdminSettings) error {
	return nil
}

func (mock *MockLoggerClient) Log(severity string, message string) error {
	mssg := LogMessage{Service: "MockLogger", Address: "0.0.0.0", Severity: severity, Message: message, Time: time.Now().String()}
	return mock.PostToMessages(&mssg)
}

func (mock *MockLoggerClient) Fatal(err error) error {
	return mock.Log(SeverityFatal, fmt.Sprintf("%v", err))
}

func (mock *MockLoggerClient) Error(text string, err error) error {
	return mock.Log(SeverityError, fmt.Sprintf("%v", err))
}
