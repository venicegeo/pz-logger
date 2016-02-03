package client

import (
	"fmt"
	"log"
	"time"
)

// implements Logger
type MockLogger struct{}

func (*MockLogger) PostToMessages(m *LogMessage) error {
	log.Printf("MOCKLOG: %v", m)
	return nil
}

func (*MockLogger) GetFromMessages() ([]LogMessage, error) {
	return nil, nil
}

func (*MockLogger) GetFromAdminStats() (*LoggerAdminStats, error) {
	return &LoggerAdminStats{}, nil
}

func (*MockLogger) GetFromAdminSettings() (*LoggerAdminSettings, error) {
	return &LoggerAdminSettings{}, nil
}

func (*MockLogger) PostToAdminSettings(*LoggerAdminSettings) error {
	return nil
}

func (mock *MockLogger) Log(severity string, message string) error {
	mssg := LogMessage{Service: "MockLogger", Address: "0.0.0.0", Severity: severity, Message: message, Time: time.Now().String()}
	return mock.PostToMessages(&mssg)
}

func (mock *MockLogger) Fatal(err error) error {
	return mock.Log(SeverityFatal, fmt.Sprintf("%v", err))
}

func (mock *MockLogger) Error(text string, err error) error {
	return mock.Log(SeverityError, fmt.Sprintf("%v", err))
}
