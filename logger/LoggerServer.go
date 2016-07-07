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

type LoggerServer struct {
	logger *LoggerService
	Routes []piazza.RouteData
}

func (server *LoggerServer) handleGetRoot(c *gin.Context) {
	resp := server.logger.GetRoot()
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *LoggerServer) handlePostMessage(c *gin.Context) {
	var mssg Message
	err := c.BindJSON(&mssg)
	if err != nil {
		resp := &piazza.JsonResponse{StatusCode: http.StatusBadRequest, Message: err.Error()}
		c.IndentedJSON(resp.StatusCode, resp)
	}
	resp := server.logger.PostMessage(&mssg)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *LoggerServer) handleGetStats(c *gin.Context) {
	resp := server.logger.GetStats()
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *LoggerServer) handleGetMessage(c *gin.Context) {
	resp := server.logger.GetMessage(c.Query, c.GetQuery)
	c.IndentedJSON(resp.StatusCode, resp)
}

func (server *LoggerServer) Init(logger *LoggerService) {
	server.logger = logger

	server.Routes = []piazza.RouteData{
		{"GET", "/", server.handleGetRoot},
		{"GET", "/message", server.handleGetMessage},
		{"POST", "/message", server.handlePostMessage},
		{"GET", "/admin/stats", server.handleGetStats},
	}
}
