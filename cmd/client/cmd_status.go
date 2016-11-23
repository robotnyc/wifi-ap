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
)

type restartCommand struct{}

func (cmd *restartCommand) Execute(args []string) error {
	req := make(map[string]string)
	req["action"] = "restart-ap"

	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = sendHTTPRequest(getServiceStatusURI(), "POST", bytes.NewReader(b))
	if err != nil {
		return err
	}

	return nil
}

type statusCommand struct{}

func (cmd *statusCommand) Execute(args []string) error {
	response, err := sendHTTPRequest(getServiceStatusURI(), "GET", nil)
	if err != nil {
		return err
	}
    printMapSorted(response.Result)
	return nil
}

func init() {
	cmd, _ := addCommand("status", "Show various status information about the access point", "", &statusCommand{})
	cmd.SubcommandsOptional = true

	cmd.AddCommand("restart-ap", "Restart access point", "", &restartCommand{})
}
