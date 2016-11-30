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
	"fmt"

	"gopkg.in/check.v1"
)

func (s *S) TestBackgroundProcessStartStop(c *check.C) {
	p, err := NewBackgroundProcess("/bin/sleep", "1000")
	c.Assert(err, check.IsNil)
	c.Assert(p.Running(), check.Equals, false)
	c.Assert(p.Start(), check.IsNil)
	c.Assert(p.Running(), check.Equals, true)
	c.Assert(p.Stop(), check.IsNil)
	c.Assert(p.Running(), check.Equals, false)
	c.Assert(p.Restart(), check.IsNil)
	c.Assert(p.Running(), check.Equals, true)
	c.Assert(p.Start(), check.DeepEquals, fmt.Errorf("Background process is already running"))
	c.Assert(p.Running(), check.Equals, true)
}
