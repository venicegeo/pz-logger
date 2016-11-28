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

package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
)

const MESSAGE string = "message"
const AUDIT string = "audit"
const METRIC string = "metric"

type SyslogMessage struct {
	Facility    int         `json:"facility"`
	Severity    int         `json:"severity"`
	Version     int         `json:"version"`
	Timestamp   interface{} `json:"timestamp"`
	Hostname    interface{} `json:"hostname"`
	IPAddress   interface{} `json:"ipaddress"`
	Application interface{} `json:"application"`
	Process     interface{} `json:"process"`
	MessageID   interface{} `json:"messageId"`
	AuditData   interface{} `json:"audit_data"`
	MetricData  interface{} `json:"metric_data"`
	Message     interface{} `json:"message"`
}
type AuditElement struct {
	Actor  interface{} `json:"actor"`
	Action interface{} `json:"action"`
	Actee  interface{} `json:"actee"`
}
type MetricElement struct {
	Name   interface{} `json:"name"`
	Value  interface{} `json:"value"`
	Object interface{} `json:"object"`
}

func (m *SyslogMessage) validate() error {
	if m.Hostname == nil && m.IPAddress == nil {
		return errors.New("Neither hostname nor IP address were supplied")
	} else if m.Hostname != nil && m.IPAddress != nil {
		return errors.New("Both hostname and IP address were supplied")
	}
	if m.Facility != 1 {
		return errors.New("Bad facility")
	}
	if m.Severity != 2 && m.Severity != 3 && m.Severity != 4 && m.Severity != 5 && m.Severity != 6 && m.Severity != 7 {
		return errors.New("Bad severity")
	}
	if m.Version != 1 {
		return errors.New("Version is not 1")
	}
	if m.Timestamp == nil { //TODO
		m.Timestamp = "-"
	}
	if m.Application == nil {
		m.Application = "-"
	} else if _, ok := m.Application.(string); !ok {
		return errors.New("Application not type string")
	}
	//TODO valid application
	if m.Process == nil {
		m.Process = "-"
	}
	//TODO valid process
	if m.MetricData == nil {
		m.MetricData = "-"
	}
	if m.Message == nil {
		m.Message = ""
	} else {
		m.Message = " " + fmt.Sprint(m.Message)
	}
	return nil
}
func (me *MetricElement) Validate() error {
	//TODO
	//Valid Name
	//Valid object
	return nil
}

func (ae *AuditElement) Validate() error {
	//TODO
	//Valid uuid
	//Valid action
	return nil
}

//TODO
func (m *SyslogMessage) ValidateMessage() error {
	var err error

	if err = m.validate(); err != nil {
		return err
	}

	return nil
}
func (m *SyslogMessage) ValidateAudit() error {
	var err error

	if err = m.validate(); err != nil {
		return err
	}

	ae := &AuditElement{}
	if ae, err = m.getAuditData(); err != nil {
		return err
	}
	if err = ae.Validate(); err != nil {
		return err
	}

	return nil
}
func (m *SyslogMessage) ValidateMetric() error {
	var err error

	if err = m.validate(); err != nil {
		return err
	}

	me := &MetricElement{}
	if me, err = m.getMetricData(); err != nil {
		return err
	}
	if err = me.Validate(); err != nil {
		return err
	}

	return nil
}

func (m *SyslogMessage) getAuditData() (*AuditElement, error) {
	if m.AuditData == nil {
		return nil, errors.New("No audit data supplied")
	}
	str, err := piazza.StructInterfaceToString(m.AuditData)
	if err != nil {
		return nil, err
	}
	ae := AuditElement{}
	if err = json.Unmarshal([]byte(str), &ae); err != nil {
		return nil, err
	}
	if ae.Actee == nil || ae.Action == nil || ae.Actor == nil {
		return nil, errors.New("Invalid audit data")
	}
	return &ae, nil
}
func (m *SyslogMessage) getMetricData() (*MetricElement, error) {
	if m.MetricData == nil {
		return nil, errors.New("No metric data supplied")
	}
	str, err := piazza.StructInterfaceToString(m.MetricData)
	if err != nil {
		return nil, err
	}
	me := MetricElement{}
	if err = json.Unmarshal([]byte(str), &me); err != nil {
		return nil, err
	}
	if me.Name == nil || me.Object == nil || me.Value == nil {
		return nil, errors.New("Invalid metric data")
	}
	return &me, nil
}

func (m *SyslogMessage) toRFC(typ string) (string, error) {
	res := "<" + strconv.Itoa(m.Facility*8+m.Severity) + ">" + strconv.Itoa(m.Version) + " "
	res += fmt.Sprint(m.Timestamp) + " "
	if m.Hostname == nil {
		res += fmt.Sprint(m.IPAddress) + " "
	} else {
		res += fmt.Sprint(m.Hostname) + " "
	}
	res += fmt.Sprint(m.Application) + " "
	res += fmt.Sprint(m.Process) + " "
	res += fmt.Sprint(m.MessageID) + " "
	switch typ {
	case AUDIT:
		ae, err := m.getAuditData()
		if err != nil {
			return "", err
		}
		res += fmt.Sprintf(`[pzaudit@%v actor="%s" action="%s" actee="%s"]`, "TODO", ae.Actor, ae.Action, ae.Actee)
	case METRIC:
		me, err := m.getMetricData()
		if err != nil {
			return "", err
		}
		res += fmt.Sprintf(`[pzmetric@%v name="%s" value="%f" object="%s"]`, "TODO", me.Name, me.Value, me.Object)

	}
	res += fmt.Sprint(m.Message)
	return res, nil
}
