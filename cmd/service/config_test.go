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
	"strings"
	"testing"

	"gopkg.in/check.v1"
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
