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
	"gopkg.in/check.v1"
	"net/http"
)

func (s *S) TestMakeErrorResponse(c *check.C) {
	resp := makeErrorResponse(http.StatusInternalServerError, "my error message", "internal-error")
	c.Assert(resp.Result, check.DeepEquals, map[string]interface{}{
		"message": "my error message",
		"kind":    "internal-error",
	})
	c.Assert(resp.Status, check.Equals, "Internal Server Error")
	c.Assert(resp.StatusCode, check.Equals, http.StatusInternalServerError)
	c.Assert(resp.Type, check.Equals, "error")
}

func (s *S) TestMakeResponse(c *check.C) {
	resp := makeResponse(http.StatusOK, nil)
	c.Assert(resp.Result, check.DeepEquals, map[string]interface{}{})
	c.Assert(resp.Status, check.Equals, "OK")
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Type, check.Equals, "sync")

	data := map[string]interface{}{"foo": "bar"}
	resp = makeResponse(http.StatusOK, data)
	c.Assert(resp.Result, check.DeepEquals, data)
	c.Assert(resp.Status, check.Equals, "OK")
	c.Assert(resp.StatusCode, check.Equals, http.StatusOK)
	c.Assert(resp.Type, check.Equals, "sync")
}
