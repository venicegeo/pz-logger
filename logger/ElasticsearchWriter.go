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
	"log"
	"strconv"
	"sync"

	"github.com/venicegeo/pz-gocommon/elasticsearch"
	piazza "github.com/venicegeo/pz-gocommon/gocommon"
	"github.com/venicegeo/pz-gocommon/syslog"
)

//---------------------------------------------------------------------

const schemaMapping = `{
	"dynamic": "strict",
	"properties": {
		"facility": {
			"type": "int"
		},
		"severity": {
			"type": "int"
		},
		"version": {
			"type": "int"
		},
		"timeStamp": {
			"type": "date"
		},
		"hostName": {
			"type": "string"
		},
		"application": {
			"type": "string"
		},
		"process": {
			"type": "string"
		},
		"messageId": {
			"type": "string"
		},
		"message": {
			"type": "string"
		}
	}
}`

const logSchema = "LogData7x"
const auditSchema = "AuditData7x"
const logSchemaMapping = "{\"LogData7\": " + schemaMapping + " }"
const auditSchemaMapping = "{\"AuditData7\": " + schemaMapping + " }"

type ElasticsearchWriter struct {
	sync.Mutex
	id      int
	esIndex elasticsearch.IIndex
}

func NewElasticsearchWriter(esIndex elasticsearch.IIndex) (*ElasticsearchWriter, error) {
	w := &ElasticsearchWriter{
		esIndex: esIndex,
	}

	err := setupIndex(esIndex)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// Write writes the message to the supplied file.
func (w *ElasticsearchWriter) Write(mssg *syslog.Message) error {
	var err error

	err = mssg.Validate()
	if err != nil {
		return err
	}

	mssgOld, err := convertNewMessageToOld(mssg)
	if err != nil {
		return err
	}

	w.Lock()
	idStr := strconv.Itoa(w.id)
	w.id++
	w.Unlock()

	_, err = w.esIndex.PostData(logSchema, idStr, mssgOld)
	if err != nil {
		log.Printf("old message post: %s", err.Error())
		// don't return yet, the audit post might still work
	}

	if mssg.AuditData != nil {
		_, err = w.esIndex.PostData(auditSchema, idStr, mssgOld)
		if err != nil {
			log.Printf("old message audit post: %s", err.Error())
		}
	}

	return nil
}

//---------------------------------------------------------------------

func setupIndex(esIndex elasticsearch.IIndex) error {
	err := createIndex(esIndex)
	if err != nil {
		return err
	}
	err = createType(esIndex, logSchema, logSchemaMapping)
	if err != nil {
		return err
	}
	err = createType(esIndex, auditSchema, auditSchemaMapping)
	return err
}

func createIndex(esIndex elasticsearch.IIndex) error {
	ok, err := esIndex.IndexExists()
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	log.Printf("Creating index: %s", esIndex.IndexName())
	err = esIndex.Create("")
	return err
}

func createType(
	esIndex elasticsearch.IIndex,
	schema string,
	mapping string) error {

	ok, err := esIndex.TypeExists(schema)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	log.Printf("Creating type: %s", schema)
	err = esIndex.SetMapping(schema, piazza.JsonString(mapping))
	return err
}
