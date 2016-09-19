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
	"io"
	"net/http"
	"os"

	"github.com/jessevdk/go-flags"
)

const (
	servicePort        = 5005
	configurationV1Uri = "/v1/configuration"
)

type serviceResponse struct {
	Result     map[string]string `json:"result"`
	Status     string            `json:"status"`
	StatusCode int               `json:"status-code"`
	Type       string            `json:"type"`
}

type commonOptions struct {
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
}

type setCommand struct{}

func getServiceConfigurationURI() string {
	return fmt.Sprintf("http://localhost:%d%s", servicePort, configurationV1Uri)
}

func sendHTTPRequest(uri string, method string, body io.Reader) (*serviceResponse, error) {
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	realResponse := &serviceResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&realResponse); err != nil {
		return nil, err
	}

	if realResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed: %s", realResponse.Result["message"])
	}

	return realResponse, nil
}

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
	if len(args) != 1 {
		return fmt.Errorf("usage: %s get <key>\n", os.Args[0])
	}

	response, err := sendHTTPRequest(getServiceConfigurationURI(), "GET", nil)
	if err != nil {
		return err
	}

	wantedKey := args[0]

	if val, ok := response.Result[wantedKey]; ok {
		fmt.Fprintf(os.Stdout, "%s\n", val)
	} else {
		return fmt.Errorf("Config item '%s' does not exist", wantedKey)
	}

	return nil
}

type configCommand struct{}

func (cmd *configCommand) Execute(args []string) error {
	return nil
}

func main() {
	var parser = flags.NewParser(&commonOptions{}, flags.Default)
	cmdConfig, _ := parser.AddCommand("config", "Adjust the service configuration", "", &configCommand{})
	cmdConfig.AddCommand("set", "", "", &setCommand{})
	cmdConfig.AddCommand("get", "", "", &getCommand{})

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}
