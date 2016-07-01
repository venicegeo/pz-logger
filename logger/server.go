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
	_ "fmt"

	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

func handleGetRoot(c *gin.Context) {
	resp, err := GetRoot(c)
	if err != nil {
		c.JSON(resp.StatusCode, resp)
	} else {
		c.JSON(resp.StatusCode, resp)
	}
}

func handlePostMessages(c *gin.Context) {
	resp, err := PostMessages(c)
	if err != nil {
		c.JSON(err.StatusCode, err)
	} else {
		c.JSON(resp.StatusCode, resp)
	}
}

func handleGetAdminStats(c *gin.Context) {
	resp, err := GetAdminStats(c)
	if err != nil {
		c.JSON(err.StatusCode, err)
	} else {
		c.JSON(resp.StatusCode, resp)
	}
}

func handleGetMessages(c *gin.Context) {
	resp, err := GetMessages(c)
	if err != nil {
		c.JSON(err.StatusCode, err)
	} else {
		c.JSON(resp.StatusCode, resp)
	}
}

var Routes = []piazza.RouteData{
	{"GET", "/", handleGetRoot},
	{"GET", "/message", handleGetMessages},
	{"GET", "/admin/stats", handleGetAdminStats},
	{"POST", "message", handlePostMessages},
}
