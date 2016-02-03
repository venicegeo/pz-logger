package client

import (
	"errors"
	piazza "github.com/venicegeo/pz-gocommon"
	"fmt"
	"encoding/json"
	"net/http"
	"bytes"
	"io/ioutil"
	"log"
	"time"
	"os"
)

type PzLoggerClient struct{
	Url string
	Name string
	Address string
}

func NewPzLoggerClient(address string) *PzLoggerClient {
	c := new(PzLoggerClient)
	c.Url = fmt.Sprintf("http://%s/v1", address)
	c.Name = "pz-logger"
	c.Address = address
	return c
}

func (c *PzLoggerClient) PostToMessages(mssg *LogMessage) error {

	mssgData, err := json.Marshal(mssg)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.Url + "/messages", piazza.ContentTypeJSON, bytes.NewBuffer(mssgData))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

func (c *PzLoggerClient) GetFromMessages() ([]LogMessage, error) {

	resp, err := http.Get(c.Url + "/messages")
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

func (c *PzLoggerClient) GetFromAdminStats() (*LoggerAdminStats, error) {

	resp, err := http.Get(c.Url + "/admin/stats")
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

func (c *PzLoggerClient) GetFromAdminSettings() (*LoggerAdminSettings, error) {

	resp, err := http.Get(c.Url + "/admin/settings")
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

func (c *PzLoggerClient) PostToAdminSettings(settings *LoggerAdminSettings) error {

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.Url + "/admin/settings", piazza.ContentTypeJSON, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}

	return nil
}

///////////////////

func (pz *PzLoggerClient) postLogMessage(mssg *LogMessage) error {

	data, err := json.Marshal(mssg)
	if err != nil {
		log.Printf("pz-logger failed to marshall request: %v", err)
		return err
	}

	resp, err := http.Post(pz.Url + "/messages", piazza.ContentTypeJSON, bytes.NewBuffer(data))
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
func (pz *PzLoggerClient) Log(severity string, message string) error {

	mssg := LogMessage{Service: pz.Name, Address: pz.Address, Severity: severity, Message: message, Time: time.Now().String()}

	return pz.postLogMessage(&mssg)
}

func (pz *PzLoggerClient) Fatal(err error) {
	log.Printf("Fatal: %v", err)

	mssg := LogMessage{Service: pz.Name, Address: pz.Address, Severity: SeverityFatal, Message: fmt.Sprintf("%v", err), Time: time.Now().String()}
	pz.postLogMessage(&mssg)

	os.Exit(1)
}

func (pz *PzLoggerClient) Error(text string, err error) error {
	log.Printf("Error: %v", err)

	s := fmt.Sprintf("%s: %v", text, err)

	mssg := LogMessage{Service: pz.Name, Address: pz.Address, Severity: SeverityError, Message: s, Time: time.Now().String()}
	return pz.postLogMessage(&mssg)
}
