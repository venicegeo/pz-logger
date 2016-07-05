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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/venicegeo/pz-gocommon/gocommon"
)

func handleGetRoot(c *gin.Context) {
	resp := GetRoot()
	c.JSON(resp.StatusCode, resp)
}

func handlePostMessages(c *gin.Context) {
	var mssg Message
	err := c.BindJSON(&mssg)
	if err != nil {
		resp := &piazza.JsonResponse{StatusCode: http.StatusBadRequest, Message: err.Error()}
		c.JSON(resp.StatusCode, resp)
	}
	resp := PostMessage(&mssg)
	c.JSON(resp.StatusCode, resp)
}

func handleGetAdminStats(c *gin.Context) {
	resp := GetAdminStats()
	c.JSON(resp.StatusCode, resp)
}

func handleGetMessages(c *gin.Context) {
	resp := GetMessages(c.Query, c.GetQuery)
	c.JSON(resp.StatusCode, resp)
}

var Routes = []piazza.RouteData{
	{"GET", "/", handleGetRoot},
	{"GET", "/message", handleGetMessages},
	{"GET", "/admin/stats", handleGetAdminStats},
	{"POST", "message", handlePostMessages},
}
