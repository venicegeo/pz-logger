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
	"os"
	"time"
)

type PzLoggerService struct {
	url     string
	name    string
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

func (c PzLoggerService) GetName() string {
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

func (pz *PzLoggerService) postLogMessage(mssg *LogMessage) error {

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

// Log sends a LogMessage to the logger.
// TODO: support fmt
func (pz *PzLoggerService) Log(severity string, message string) error {

	mssg := LogMessage{Service: pz.name, Address: pz.address, Severity: severity, Message: message, Time: time.Now().String()}

	return pz.postLogMessage(&mssg)
}

func (pz *PzLoggerService) Fatal(err error) error {
	log.Printf("Fatal: %v", err)

	mssg := LogMessage{Service: pz.name, Address: pz.address, Severity: SeverityFatal, Message: fmt.Sprintf("%v", err), Time: time.Now().String()}
	pz.postLogMessage(&mssg)

	os.Exit(1)

	/*notreached*/
	return nil
}

func (pz *PzLoggerService) Error(text string, err error) error {
	log.Printf("Error: %v", err)

	s := fmt.Sprintf("%s: %v", text, err)

	mssg := LogMessage{Service: pz.name, Address: pz.address, Severity: SeverityError, Message: s, Time: time.Now().String()}
	return pz.postLogMessage(&mssg)
}
