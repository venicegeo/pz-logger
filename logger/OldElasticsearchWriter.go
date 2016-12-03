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

//---------------------------------------------------------------------
/****
const oldSchemaMapping = `{
	"dynamic": "strict",
	"properties": {
		"service": {
			"type": "string",
			"store": true,
			"index": "not_analyzed"
		},
	"address": {
		"type": "string",
		"store": true,
		"index": "not_analyzed"
	},
	"createdOn": {
		"type": "date",
		"store": true,
		"index": "not_analyzed"
	},
	"severity": {
		"type": "string",
		"store": true,
		"index": "not_analyzed"
	},
	"message": {
		"type": "string",
		"store": true,
		"index": "analyzed"
	}
}`

const oldLogSchema = "LogData7"
const oldAuditSchema = "AuditData7"
const oldLogSchemaMapping = "\"LogData7\": " + oldSchemaMapping + " }"
const oldAuditSchemaMapping = "\"AuditData7\": " + oldSchemaMapping + " }"

type OldElasticsearchWriter struct {
	sync.Mutex
	id      int
	esIndex elasticsearch.IIndex
}

func NewOldElasticsearchWriter(esIndex elasticsearch.IIndex) (*ElasticsearchWriter, error) {
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
func (w *OldElasticsearchWriter) Write(mssg *syslog.Message) error {
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

	_, err = w.esIndex.PostData(oldLogSchema, idStr, mssgOld)
	if err != nil {
		log.Printf("old message post: %s", err.Error())
		// don't return yet, the audit post might still work
	}

	if mssg.AuditData != nil {
		_, err = w.esIndex.PostData(oldAuditSchema, idStr, mssgOld)
		if err != nil {
			log.Printf("old message audit post: %s", err.Error())
		}
	}

	return nil
}
****/
