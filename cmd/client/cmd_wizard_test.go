//
// Copyright (C) 2017 Canonical Ltd
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
	"bufio"
	"math/rand"
	"strconv"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

// gopkg.in/check.v1 stuff
func TestWizard(t *testing.T) { TestingT(t) }

type WizardSuite struct{}

var _ = Suite(&WizardSuite{})

func (s *WizardSuite) SetUpTest(c *C) {
	rand.Seed(time.Now().UnixNano())
}

func mockUserInput(reply string) func (reader *bufio.Reader) string {
	return func (_ *bufio.Reader) string {
		return reply
	}
}

var passRegExp = "[a-zA-Z0-9+/]{"+strconv.Itoa(DefaultPassworthLength)+"}"

func (s *WizardSuite) TestPasswordGeneration(c *C) {
	password := generatePassword(DefaultPassworthLength)
	c.Assert(password, HasLen, DefaultPassworthLength)
	c.Assert(password, Matches, passRegExp)

	readUserInput = mockUserInput("short")
	password, err := askForPassword(nil)
	c.Assert(err, NotNil)

	readUserInput = mockUserInput("4/Valid+Passw0rd")
	password, err = askForPassword(nil)
	c.Assert(err, IsNil)

	c.Assert(password, HasLen, DefaultPassworthLength)
	c.Assert(password, Matches, passRegExp)
}
