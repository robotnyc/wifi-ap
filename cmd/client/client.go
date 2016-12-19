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
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

const (
	socketPathSuffix   = "sockets/control"
	configurationV1Uri = "/v1/configuration"
	statusV1Uri        = "/v1/status"
)

type serviceResponse struct {
	Result     map[string]interface{} `json:"result"`
	Status     string                 `json:"status"`
	StatusCode int                    `json:"status-code"`
	Type       string                 `json:"type"`
}

func getServiceConfigurationURI() string {
	return fmt.Sprintf("http://unix%s", configurationV1Uri)
}

func getServiceStatusURI() string {
	return fmt.Sprintf("http://unix%s", statusV1Uri)
}

type doer interface {
	Do(*http.Request) (*http.Response, error)
}

var customDoer doer

func unixDialer(_, _ string) (net.Conn, error) {
	path := filepath.Join(os.Getenv("SNAP_DATA"), socketPathSuffix)
	return net.Dial("unix", path)
}

func sendHTTPRequest(uri string, method string, body io.Reader) (*serviceResponse, error) {
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	var resp *http.Response

	if customDoer == nil {
		client := &http.Client{
			Transport: &http.Transport{
				Dial: unixDialer,
			},
		}
		resp, err = client.Do(req)
	} else {
		resp, err = customDoer.Do(req)
	}

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
