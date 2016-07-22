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

package logger_demo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/venicegeo/pz-gocommon/gocommon"
	"github.com/venicegeo/pz-logger/logger"
)

func sleep() {
	time.Sleep(1 * time.Second)
}

type LoggerTester struct {
	suite.Suite
	client *logger.Client
}

func (suite *LoggerTester) setupFixture() {
	t := suite.T()
	assert := assert.New(t)

	client, err := logger.NewClient2("https://pz-logger.int.geointservices.io")
	assert.NoError(err)
	suite.client = client
}

func (suite *LoggerTester) teardownFixture() {
}

func TestRunSuite(t *testing.T) {
	s := &LoggerTester{}
	suite.Run(t, s)
}

func (suite *LoggerTester) getMessages(key string) []logger.Message {
	t := suite.T()
	assert := assert.New(t)

	client := suite.client

	format := piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   piazza.PaginationOrderAscending,
		SortBy:  "createdOn",
	}
	ms, _, err := client.GetMessages(format, map[string]string{})
	assert.NoError(err)
	assert.True(len(ms) > 0)

	ret := make([]logger.Message, 0)
	for _, m := range ms {
		if m.Message == key {
			ret = append(ret, m)
		}
	}

	return ret
}

func (suite *LoggerTester) verifyMessageExists(expected *logger.Message) bool {
	t := suite.T()
	assert := assert.New(t)

	client := suite.client

	format := piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   piazza.PaginationOrderAscending,
		SortBy:  "createdOn",
	}
	ms, _, err := client.GetMessages(format, map[string]string{})
	assert.NoError(err)
	assert.Len(ms, format.PerPage)

	for _, m := range ms {
		if m.String() == expected.String() {
			return true
		}
	}
	return false
}

func (suite *LoggerTester) TestPost() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	var err error

	key := time.Now().String()

	data := &logger.Message{
		Service:   "log-tester",
		Address:   "128.1.2.3",
		CreatedOn: time.Now(),
		Severity:  "Info",
		Message:   key,
	}

	err = client.PostMessage(data)
	assert.NoError(err, "PostToMessages")

	sleep()

	assert.True(suite.verifyMessageExists(data))
}

func (suite *LoggerTester) xTestAdmin() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	stats, err := client.GetStats()
	assert.NoError(err, "GetFromAdminStats")
	assert.True(stats.NumMessages > 0)
}

func (suite *LoggerTester) xTestPagination() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	format := piazza.JsonPagination{
		PerPage: 1,
		Page:    0,
		SortBy:  "createdOn",
		Order:   piazza.PaginationOrderAscending,
	}
	params := map[string]string{}

	ms, _, err := client.GetMessages(format, params)
	assert.NoError(err)
	assert.Len(ms, format.PerPage)
	//	log.Printf("%#v ===", ms)
}

func (suite *LoggerTester) xTest06OtherParams() {
	t := suite.T()
	assert := assert.New(t)

	suite.setupFixture()
	defer suite.teardownFixture()

	client := suite.client

	client.SetService("myservice", "1.2.3.4")

	sometime := time.Now()

	var testData = []logger.Message{
		{
			Address:   "gnemud7smfv/10.254.0.66",
			Message:   "Received Message to Relay on topic Request-Job with key f3b63085-b482-4ae8-8297-3c7d1fcfff7d",
			Service:   "Dispatcher",
			Severity:  "Info",
			CreatedOn: sometime,
		}, {
			Address:   "gnfbnqsn5m9/10.254.0.14",
			Message:   "Processed Update Status for Job 6d0ea538-4382-4ea5-9669-56519b8c8f58 with Status Success",
			Service:   "JobManager",
			Severity:  "Info",
			CreatedOn: sometime,
		}, {
			Address:   "0.0.0.0",
			Message:   "generated 1: 09d4ec60-ea61-4066-8697-5568a47f84bf",
			Service:   "pz-uuidgen",
			Severity:  "Info",
			CreatedOn: sometime,
		}, {
			Address:   "gnfbnqsn5m9/10.254.0.14",
			Message:   "Handling Job with Topic Create-Job for Job ID 09d4ec60-ea61-4066-8697-5568a47f84bf",
			Service:   "JobManager",
			Severity:  "Info",
			CreatedOn: sometime,
		}, {
			Address:   "gnfbnqsn5m9/10.254.0.14",
			Message:   "Handling Job with Topic Update-Job for Job ID be4ce034-1187-4a4f-95a9-a9c31826248b",
			Service:   "JobManager",
			Severity:  "Info",
			CreatedOn: sometime,
		},
	}

	for _, e := range testData {
		// log.Printf("%d, %v\n", i, e)
		err := client.PostMessage(&e)
		assert.NoError(err)
	}

	sleep()

	format := piazza.JsonPagination{
		PerPage: 100,
		Page:    0,
		Order:   piazza.PaginationOrderDescending,
		SortBy:  "createdOn",
	}

	msgs, _, err := client.GetMessages(format,
		map[string]string{
			"service":  "JobManager",
			"contains": "Success",
		})
	assert.NoError(err)
	assert.Len(msgs, 1)

	//for _, msg := range msgs {
	//log.Printf("%v\n", msg)
	//}

	msgs, _, err = client.GetMessages(format,
		map[string]string{
			"before": "1461181461",
			"after":  "1461181362",
		})

	assert.NoError(err)
	assert.Len(msgs, 4)

	//for _, msg := range msgs {
	//log.Printf("%v\n", msg)
	//}

}
