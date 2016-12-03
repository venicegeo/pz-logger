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
	"time"

	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	syslog "github.com/venicegeo/pz-gocommon/syslog"
)

type IClient interface {
	// admin interfaces
	GetVersion() (*piazza.Version, error)
	GetStats() (*Stats, error)

	// read support
	GetAsOld(format *piazza.JsonPagination, params *piazza.HttpQueryParams) ([]Message, int, error)
	GetAsString(format *piazza.JsonPagination, params *piazza.HttpQueryParams) ([]string, int, error)
	GetAsJson(format *piazza.JsonPagination, params *piazza.HttpQueryParams) ([]syslog.Message, int, error)

	// config support
	SetService(name piazza.ServiceName, address string)
}

//---------------------------------------------------------------------------

type Stats struct {
	CreatedOn time.Time `json:"createdOn"`

	// this is the number of messages since the service was started,
	// not the total number of messages in the system
	NumMessages float64 `json:"numMessages"`
}

//---------------------------------------------------------------------------

func init() {
	piazza.JsonResponseDataTypes["logger.Message"] = "logmessage"
	piazza.JsonResponseDataTypes["*logger.Message"] = "logmessage"
	piazza.JsonResponseDataTypes["[]logger.Message"] = "logmessage-list"
	piazza.JsonResponseDataTypes["logger.Stats"] = "logstats"
	piazza.JsonResponseDataTypes["*logger.Stats"] = "logstats"
	piazza.JsonResponseDataTypes["syslog.Message"] = "syslogmessage"
	piazza.JsonResponseDataTypes["*syslog.Message"] = "syslogmessage"
	piazza.JsonResponseDataTypes["[]syslog.Message"] = "syslogmessage-list"
}
