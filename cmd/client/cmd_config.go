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
	"os"
	"sort"
)

type setCommand struct{}

func (cmd *setCommand) Execute(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: %s set <key> <value>\n", os.Args[0])
	}

	request := make(map[string]string)
	request[args[0]] = args[1]
	b, err := json.Marshal(request)

	_, err = sendHTTPRequest(getServiceConfigurationURI(), "POST", bytes.NewReader(b))
	if err != nil {
		return err
	}

	return nil
}

type getCommand struct{}

func (cmd *getCommand) Execute(args []string) error {
	response, err := sendHTTPRequest(getServiceConfigurationURI(), "GET", nil)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		wantedKey := args[0]
		if val, ok := response.Result[wantedKey]; ok {
			fmt.Fprintf(os.Stdout, "%s\n", val)
		} else {
			return fmt.Errorf("Config item '%s' does not exist", wantedKey)
		}
	} else {
		sortedKeys := make([]string, 0, len(response.Result))
		for key, _ := range response.Result {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sortedKeys)
		for n := range sortedKeys {
			fmt.Fprintf(os.Stdout, "%s: %s\n", sortedKeys[n], response.Result[sortedKeys[n]])
		}
	}

	return nil
}

type configCommand struct{}

func (cmd *configCommand) Execute(args []string) error {
	return nil
}

func init() {
	cmdConfig, _ := addCommand("config", "Adjust the service configuration", "", &configCommand{})
	cmdConfig.AddCommand("set", "", "", &setCommand{})
	cmdConfig.AddCommand("get", "", "", &getCommand{})
}
