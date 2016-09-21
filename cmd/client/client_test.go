//
// Copyright (C) 2016 Canonical Ltd
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License version 3 as
// published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"gopkg.in/check.v1"
)

// gopkg.in/check.v1 stuff
func Test(t *testing.T) { check.TestingT(t) }

type ClientSuite struct {
	req     *http.Request
	rsp     string
	err     error
	doCalls int
	header  http.Header
	status  int
}

var _ = check.Suite(&ClientSuite{})

func (s *ClientSuite) SetUpTest(c *check.C) {
	s.req = nil
	s.rsp = ""
	s.err = nil
	s.doCalls = 0
	s.header = nil
	s.status = http.StatusOK
	// Inject ourself as doer into the client so we get called
	// for the actual http requests and they are not send out
	// over the network to a not existing service.
	customDoer = s
}

func (s *ClientSuite) Do(req *http.Request) (*http.Response, error) {
	s.req = req
	rsp := &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader(s.rsp)),
		Header:     s.header,
		StatusCode: s.status,
	}
	s.doCalls++
	return rsp, s.err
}

func (s *ClientSuite) TestServiceConfigurationUriIsCorrect(c *check.C) {
	c.Assert(getServiceConfigurationURI(), check.Equals, "http://localhost:5005/v1/configuration")
}

func (s *ClientSuite) TestSendHTTPRequestWithSuccessfullResponse(c *check.C) {
	s.rsp = `{"result":{"test1":"abc"},"status":"OK","status-code":200,"type":"sync"}`
	rsp, err := sendHTTPRequest(getServiceConfigurationURI(), "GET", nil)
	c.Assert(s.doCalls, check.Equals, 1)
	c.Assert(rsp, check.NotNil)
	c.Assert(err, check.IsNil)
	c.Assert(s.req.Method, check.Equals, "GET")
	c.Assert(rsp.Status, check.Equals, "OK")
	c.Assert(rsp.StatusCode, check.Equals, 200)
	c.Assert(rsp.Type, check.Equals, "sync")
	c.Assert(rsp.Result["test1"], check.Equals, "abc")
}

func (s *ClientSuite) TestSendHTTPRequestFails(c *check.C) {
	s.err = fmt.Errorf("Failed")
	rsp, err := sendHTTPRequest(getServiceConfigurationURI(), "GET", nil)
	c.Assert(rsp, check.IsNil)
	c.Assert(err, check.Equals, s.err)
	c.Assert(s.req.Method, check.Equals, "GET")
}

func (s *ClientSuite) TestSendHTTPRequestInvalidResponseJson(c *check.C) {
	s.rsp = `{invalid}`
	rsp, err := sendHTTPRequest(getServiceConfigurationURI(), "GET", nil)
	c.Assert(rsp, check.IsNil)
	c.Assert(err, check.NotNil)
	c.Assert(s.req.Method, check.Equals, "GET")
}

func (s *ClientSuite) TestSendHTTPRequestErrorFromService(c *check.C) {
	s.rsp = `{"result":{},"status":"Failed","status-code":500,"type":"sync"}`
	rsp, err := sendHTTPRequest(getServiceConfigurationURI(), "GET", nil)
	c.Assert(rsp, check.IsNil)
	c.Assert(err, check.NotNil)
	c.Assert(s.req.Method, check.Equals, "GET")
}

func (s *ClientSuite) TestSendHTTPRequestSendsCorrectContent(c *check.C) {
	s.rsp = `{"result":{},"status":"OK","status-code":200,"type":"sync"}`
	request := make(map[string]string)
	request["test1"] = "abc"
	b, err := json.Marshal(request)
	c.Assert(err, check.IsNil)
	c.Assert(b, check.NotNil)
	rsp, err := sendHTTPRequest(getServiceConfigurationURI(), "POST", bytes.NewReader(b))
	c.Assert(rsp, check.NotNil)
	c.Assert(err, check.IsNil)
	c.Assert(s.req.Body, check.NotNil)
}
