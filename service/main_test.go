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
	"gopkg.in/check.v1"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// gopkg.in/check.v1 stuff
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

// Test the config file path append routine
func (s *S) TestPath(c *check.C) {
	c.Assert(getConfigOnPath("/test"), check.Equals, "/test/config")
}

// List of tokens to be translated
var cfgKeys = [...][2]string{
	{"DISABLED", "disabled"},
	{"WIFI_SSID", "wifi.ssid"},
	{"WIFI_INTERFACE", "wifi.interface"},
	{"WIFI_INTERFACE_MODE", "wifi.interface-mode"},
	{"DHCP_RANGE_START", "dhcp.range-start"},
	{"MYTOKEN", "mytoken"},
	{"CFG_TOKEN", "cfg.token"},
	{"MY_TOKEN$", "my.token$"},
}

// Test token conversion from internal format
func (s *S) TestConvertKeyToRepresentationFormat(c *check.C) {
	for _, st := range cfgKeys {
		c.Assert(convertKeyToRepresentationFormat(st[0]), check.Equals, st[1])
	}
}

// Test token conversion to internal format
func (s *S) TestConvertKeyToStorageFormat(c *check.C) {
	for _, st := range cfgKeys {
		c.Assert(convertKeyToStorageFormat(st[1]), check.Equals, st[0])
	}
}

// List of malicious tokens which needs to be escaped
func (s *S) TestEscapeShell(c *check.C) {
	cmds := [...][2]string{
		{"my_ap", "my_ap"},
		{`my ap`, `"my ap"`},
		{`my "ap"`, `"my \"ap\""`},
		{`$(ps ax)`, `"\$(ps ax)"`},
		{"`ls /`", "\"\\`ls /\\`\""},
		{`c:\dir`, `"c:\\dir"`},
	}
	for _, st := range cmds {
		c.Assert(escapeTextForShell(st[0]), check.Equals, st[1])
	}
}

func (s *S) TestGetConfiguration(c *check.C) {
	// Check it we get a valid JSON as configuration
	req, err := http.NewRequest(http.MethodGet, CONFIGURATION_API_URI, nil)
	c.Assert(err, check.IsNil)

	rec := httptest.NewRecorder()

	getConfiguration(rec, req)

	body, err := ioutil.ReadAll(rec.Result().Body)
	c.Assert(err, check.IsNil)

	// Parse the returned JSON
	var resp Response
	err = json.Unmarshal(body, &resp)
	c.Assert(err, check.IsNil)

	// Check for 200 status code
	c.Assert(resp.Status, check.Equals, http.StatusText(http.StatusOK))
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Type, check.Equals, "sync")
}

func (s *S) TestChangeConfiguration(c *check.C) {
	// Test a non writable path:
	os.Setenv("SNAP_DATA", "/nodir")

	req, err := http.NewRequest(http.MethodPost, CONFIGURATION_API_URI, nil)
	c.Assert(err, check.IsNil)

	rec := httptest.NewRecorder()

	changeConfiguration(rec, req)

	c.Assert(rec.Code, check.Equals, http.StatusInternalServerError)

	body, err := ioutil.ReadAll(rec.Result().Body)
	c.Assert(err, check.IsNil)

	// Parse the returned JSON
	var resp Response
	err = json.Unmarshal(body, &resp)
	c.Assert(err, check.IsNil)

	// Check for 500 status code and other error fields
	c.Assert(resp.Status, check.Equals, http.StatusText(http.StatusInternalServerError))
	c.Assert(resp.StatusCode, check.Equals, http.StatusInternalServerError)
	c.Assert(resp.Type, check.Equals, "error")
	c.Assert(resp.Result["kind"], check.Equals, "internal-error")
	c.Assert(resp.Result["message"], check.Equals, "Can't write configuration file")

	// Test an invalid JSON
	os.Setenv("SNAP_DATA", "/tmp")
	req, err = http.NewRequest(http.MethodPost, CONFIGURATION_API_URI, strings.NewReader("not a JSON content"))
	c.Assert(err, check.IsNil)

	rec = httptest.NewRecorder()

	changeConfiguration(rec, req)

	c.Assert(rec.Code, check.Equals, http.StatusInternalServerError)

	body, err = ioutil.ReadAll(rec.Result().Body)
	c.Assert(err, check.IsNil)

	// Parse the returned JSON
	resp = Response{}
	err = json.Unmarshal(body, &resp)
	c.Assert(err, check.IsNil)

	// Check for 500 status code and other error fields
	c.Assert(resp.Status, check.Equals, http.StatusText(http.StatusInternalServerError))
	c.Assert(resp.StatusCode, check.Equals, http.StatusInternalServerError)
	c.Assert(resp.Type, check.Equals, "error")
	c.Assert(resp.Result["kind"], check.Equals, "internal-error")
	c.Assert(resp.Result["message"], check.Equals, "Malformed request")

	// Test a succesful configuration set

	// Values to be used in the config
	values := map[string]string{
		"wifi.security":            "wpa2",
		"wifi.ssid":                "UbuntuAP",
		"wifi.security-passphrase": "12345678",
	}

	// Convert the map into JSON
	args, err := json.Marshal(values)
	c.Assert(err, check.IsNil)

	req, err = http.NewRequest(http.MethodPost, CONFIGURATION_API_URI, bytes.NewReader(args))
	c.Assert(err, check.IsNil)

	rec = httptest.NewRecorder()

	// Do the request
	changeConfiguration(rec, req)

	c.Assert(rec.Code, check.Equals, http.StatusOK)

	// Read the result JSON
	body, err = ioutil.ReadAll(rec.Result().Body)
	c.Assert(err, check.IsNil)

	// Parse the returned JSON
	resp = Response{}
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

	// don't leave garbage in /tmp
	os.Remove(getConfigOnPath(os.Getenv("SNAP_DATA")))
}
