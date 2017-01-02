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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/check.v1"
)

// gopkg.in/check.v1 stuff
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

type mockBackgroundProcess struct {
	running bool
}

func (p *mockBackgroundProcess) Start() error {
	p.running = true
	return nil
}

func (p *mockBackgroundProcess) Stop() error {
	p.running = false
	return nil
}

func (p *mockBackgroundProcess) Restart() error {
	p.running = true
	return nil
}

func (p *mockBackgroundProcess) Running() bool {
	return p.running
}

func newMockServiceCommand() *serviceCommand {
	return &serviceCommand{
		s: &service{
			ap: &mockBackgroundProcess{},
		},
	}
}

func (s *S) TestGetConfiguration(c *check.C) {
	// Check it we get a valid JSON as configuration
	req, err := http.NewRequest(http.MethodGet, "/v1/configuration", nil)
	c.Assert(err, check.IsNil)

	rec := httptest.NewRecorder()

	cmd := newMockServiceCommand()
	getConfiguration(cmd, rec, req)

	body, err := ioutil.ReadAll(rec.Body)
	c.Assert(err, check.IsNil)

	// Parse the returned JSON
	var resp serviceResponse
	err = json.Unmarshal(body, &resp)
	c.Assert(err, check.IsNil)

	// Check for 200 status code
	c.Assert(resp.Status, check.Equals, http.StatusText(http.StatusOK))
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Type, check.Equals, "sync")
}

func (s *S) TestNoDefaultConfiguration(c *check.C) {
	oldsnap := os.Getenv("SNAP")
	os.Setenv("SNAP", "/nodir")
	os.Setenv("SNAP_DATA", "/tmp")

	req, err := http.NewRequest(http.MethodPost, "/v1/configuration", nil)
	c.Assert(err, check.IsNil)

	rec := httptest.NewRecorder()
	cmd := newMockServiceCommand()

	validTokens = nil

	postConfiguration(cmd, rec, req)

	c.Assert(rec.Code, check.Equals, http.StatusInternalServerError)

	body, err := ioutil.ReadAll(rec.Body)
	c.Assert(err, check.IsNil)

	// Parse the returned JSON
	var resp serviceResponse
	err = json.Unmarshal(body, &resp)
	c.Assert(err, check.IsNil)

	// Check for 500 status code and other error fields
	c.Assert(resp.Status, check.Equals, http.StatusText(http.StatusInternalServerError))
	c.Assert(resp.StatusCode, check.Equals, http.StatusInternalServerError)
	c.Assert(resp.Type, check.Equals, "error")
	c.Assert(resp.Result["kind"], check.Equals, "internal-error")
	c.Assert(resp.Result["message"], check.Equals, "No default configuration file available")

	os.Setenv("SNAP", oldsnap)
}

func (s *S) TestWriteError(c *check.C) {
	// Test a non writable path:
	os.Setenv("SNAP_DATA", "/nodir")

	req, err := http.NewRequest(http.MethodPost, "/v1/configuration", nil)
	c.Assert(err, check.IsNil)

	rec := httptest.NewRecorder()
	cmd := newMockServiceCommand()

	validTokens, err = loadValidTokens(filepath.Join(os.Getenv("SNAP"), "/conf/default-config"))
	c.Assert(validTokens, check.NotNil)
	c.Assert(err, check.IsNil)

	postConfiguration(cmd, rec, req)

	c.Assert(rec.Code, check.Equals, http.StatusInternalServerError)

	body, err := ioutil.ReadAll(rec.Body)
	c.Assert(err, check.IsNil)

	// Parse the returned JSON
	var resp serviceResponse
	err = json.Unmarshal(body, &resp)
	c.Assert(err, check.IsNil)

	// Check for 500 status code and other error fields
	c.Assert(resp.Status, check.Equals, http.StatusText(http.StatusInternalServerError))
	c.Assert(resp.StatusCode, check.Equals, http.StatusInternalServerError)
	c.Assert(resp.Type, check.Equals, "error")
	c.Assert(resp.Result["kind"], check.Equals, "internal-error")
	c.Assert(resp.Result["message"], check.Equals, "Can't write configuration file")
}

func (s *S) TestInvalidJSON(c *check.C) {
	// Test an invalid JSON
	os.Setenv("SNAP_DATA", "/tmp")
	req, err := http.NewRequest(http.MethodPost, "/v1/configuration", strings.NewReader("not a JSON content"))
	c.Assert(err, check.IsNil)

	rec := httptest.NewRecorder()
	cmd := newMockServiceCommand()

	postConfiguration(cmd, rec, req)

	c.Assert(rec.Code, check.Equals, http.StatusInternalServerError)

	body, err := ioutil.ReadAll(rec.Body)
	c.Assert(err, check.IsNil)

	// Parse the returned JSON
	resp := serviceResponse{}
	err = json.Unmarshal(body, &resp)
	c.Assert(err, check.IsNil)

	// Check for 500 status code and other error fields
	c.Assert(resp.Status, check.Equals, http.StatusText(http.StatusInternalServerError))
	c.Assert(resp.StatusCode, check.Equals, http.StatusInternalServerError)
	c.Assert(resp.Type, check.Equals, "error")
	c.Assert(resp.Result["kind"], check.Equals, "internal-error")
	c.Assert(resp.Result["message"], check.Equals, "Malformed request")
}

func (s *S) TestInvalidToken(c *check.C) {
	// Test a succesful configuration set
	// Values to be used in the config
	values := map[string]string{
		"wifi.security":            "wpa2",
		"wifi.ssid":                "UbuntuAP",
		"wifi.security-passphrase": "12345678",
		"bad.token":                "xyz",
	}

	// Convert the map into JSON
	args, err := json.Marshal(values)
	c.Assert(err, check.IsNil)

	req, err := http.NewRequest(http.MethodPost, "/v1/configuration", bytes.NewReader(args))
	c.Assert(err, check.IsNil)

	rec := httptest.NewRecorder()
	cmd := newMockServiceCommand()

	validTokens, err = loadValidTokens(filepath.Join(os.Getenv("SNAP"), "/conf/default-config"))
	c.Assert(validTokens, check.NotNil)
	c.Assert(err, check.IsNil)

	// Do the request
	postConfiguration(cmd, rec, req)

	c.Assert(rec.Code, check.Equals, http.StatusInternalServerError)

	// Read the result JSON
	body, err := ioutil.ReadAll(rec.Body)
	c.Assert(err, check.IsNil)

	// Parse the returned JSON
	resp := serviceResponse{}
	err = json.Unmarshal(body, &resp)
	c.Assert(err, check.IsNil)

	// Check for 500 status code and other succesful fields
	c.Assert(resp.Status, check.Equals, http.StatusText(http.StatusInternalServerError))
	c.Assert(resp.StatusCode, check.Equals, http.StatusInternalServerError)
	c.Assert(resp.Type, check.Equals, "error")
}

func (s *S) TestChangeConfiguration(c *check.C) {
	os.Setenv("SNAP", "../..")

	// Values to be used in the config
	values := map[string]string{
		"disabled":                 "0",
		"wifi.security":            "wpa2",
		"wifi.ssid":                "UbuntuAP",
		"wifi.security-passphrase": "12345678",
	}

	// Convert the map into JSON
	args, err := json.Marshal(values)
	c.Assert(err, check.IsNil)

	req, err := http.NewRequest(http.MethodPost, "/v1/configuration", bytes.NewReader(args))
	c.Assert(err, check.IsNil)

	rec := httptest.NewRecorder()
	cmd := newMockServiceCommand()

	validTokens, err = loadValidTokens(filepath.Join(os.Getenv("SNAP"), "/conf/default-config"))
	c.Assert(validTokens, check.NotNil)
	c.Assert(err, check.IsNil)

	// Do the request
	postConfiguration(cmd, rec, req)

	c.Assert(rec.Code, check.Equals, http.StatusOK)

	// Read the result JSON
	body, err := ioutil.ReadAll(rec.Body)
	c.Assert(err, check.IsNil)

	// Parse the returned JSON
	resp := serviceResponse{}
	err = json.Unmarshal(body, &resp)
	c.Assert(err, check.IsNil)

	// Check for 200 status code and other succesful fields
	c.Assert(resp.Status, check.Equals, http.StatusText(http.StatusOK))
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Type, check.Equals, "sync")

	// Read the generated config and check that values were set
	config, err := ioutil.ReadFile(getConfigOnPath(os.Getenv("SNAP_DATA")))
	c.Assert(err, check.IsNil)

	for key, value := range values {
		c.Assert(strings.Contains(string(config),
			convertKeyToStorageFormat(key)+"="+value+"\n"),
			check.Equals, true)
	}

	// As we've set 'disabled' to '0' above the AP should be active
	// now as the configuration post request will trigger an automatic
	// restart of the relevant background processes.
	c.Assert(cmd.s.ap.Running(), check.Equals, true)

	// Don't leave garbage in /tmp
	os.Remove(getConfigOnPath(os.Getenv("SNAP_DATA")))
}

func (s *S) TestGetStatusDefaultOk(c *check.C) {
	req, err := http.NewRequest(http.MethodGet, "/v1/status", nil)
	c.Assert(err, check.IsNil)

	rec := httptest.NewRecorder()

	cmd := newMockServiceCommand()

	getStatus(cmd, rec, req)

	body, err := ioutil.ReadAll(rec.Body)
	c.Assert(err, check.IsNil)

	var resp serviceResponse
	err = json.Unmarshal(body, &resp)
	c.Assert(err, check.IsNil)

	c.Assert(resp.Status, check.Equals, http.StatusText(http.StatusOK))
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Type, check.Equals, "sync")

	c.Assert(resp.Result["ap.active"], check.Equals, false)
}

func (s *S) TestGetStatusReturnsCorrectApStatus(c *check.C) {
	req, err := http.NewRequest(http.MethodGet, "/v1/status", nil)
	c.Assert(err, check.IsNil)

	rec := httptest.NewRecorder()

	cmd := newMockServiceCommand()
	cmd.s.ap.Start()

	getStatus(cmd, rec, req)

	body, err := ioutil.ReadAll(rec.Body)
	c.Assert(err, check.IsNil)

	var resp serviceResponse
	err = json.Unmarshal(body, &resp)
	c.Assert(err, check.IsNil)

	c.Assert(resp.Status, check.Equals, http.StatusText(http.StatusOK))
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Type, check.Equals, "sync")

	c.Assert(resp.Result["ap.active"], check.Equals, true)
}
