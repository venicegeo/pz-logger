package client

import (
	"fmt"
	"log"
	"time"
	"github.com/venicegeo/pz-gocommon"
)

// implements Logger
type MockLoggerService struct{
	name    string
	address string
}

func NewMockLoggerService(sys *piazza.System) (ILoggerService, error) {
	var _ piazza.IService = new(MockLoggerService)
	var _ ILoggerService = new(MockLoggerService)

	service := &MockLoggerService{name: piazza.PzLogger, address: "0.0.0.0"}

	sys.Services[piazza.PzLogger] = service

	return service, nil
}

func (m *MockLoggerService) GetName() string {
	return m.name
}

func (m *MockLoggerService) GetAddress() string {
	return m.address
}

func (*MockLoggerService) PostToMessages(m *LogMessage) error {
	log.Printf("MOCKLOG: %v", m)
	return nil
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

func (mock *MockLoggerService) Log(severity string, message string) error {
	mssg := LogMessage{Service: "MockLogger", Address: "0.0.0.0", Severity: severity, Message: message, Time: time.Now().String()}
	return mock.PostToMessages(&mssg)
}

func (mock *MockLoggerService) Fatal(err error) error {
	return mock.Log(SeverityFatal, fmt.Sprintf("%v", err))
}

func (mock *MockLoggerService) Error(text string, err error) error {
	return mock.Log(SeverityError, fmt.Sprintf("%v", err))
}
