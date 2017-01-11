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

package main

import (
	"log"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	"github.com/venicegeo/pz-gocommon/gocommon"
	syslogger "github.com/venicegeo/pz-gocommon/syslog"
	pzlogger "github.com/venicegeo/pz-logger/logger"
)

func main() {

	required := []piazza.ServiceName{piazza.PzElasticSearch}

	sys, err := piazza.NewSystemConfig(piazza.PzLogger, required)
	if err != nil {
		log.Fatal(err)
	}

	loggerIndex, loggerType, auditType, err := syslogger.GetRequiredEnvVars()
	if err != nil {
		log.Fatal(err)
	}
	pzlogger.SetLogSchema(loggerType)
	pzlogger.SetAuditSchema(auditType)

	idx, err := elasticsearch.NewIndex(sys, loggerIndex, "")
	if err != nil {
		log.Fatal(err)
	}

	logWriter, auditWriter, err := syslogger.GetRequiredESIWriters(idx, loggerType, auditType)
	if err != nil {
		log.Fatal(err)
	}

	stdOutWriter := syslogger.STDOUTWriter{}

	service := &pzlogger.Service{}
	err = service.Init(sys, logWriter, syslogger.NewMultiWriter([]syslogger.Writer{auditWriter, &stdOutWriter}), idx)
	if err != nil {
		log.Fatal(err)
	}

	server := &pzlogger.Server{}
	server.Init(service)

	genericServer := piazza.GenericServer{Sys: sys}

	err = genericServer.Configure(server.Routes)
	if err != nil {
		log.Fatal(err)
	}

	done, err := genericServer.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = <-done
	if err != nil {
		log.Fatal(err)
	}
}
