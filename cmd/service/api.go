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
	"io/ioutil"
	"net/http"
	"os"
)

var api = []*serviceCommand{
	configurationCmd,
	statusCmd,
}

var (
	configurationCmd = &serviceCommand{
		Path: "/v1/configuration",
		GET:  getConfiguration,
		POST: postConfiguration,
	}
	statusCmd = &serviceCommand{
		Path: "/v1/status",
		GET:  getStatus,
		POST: postStatus,
	}
	validTokens map[string]bool
)

func getConfiguration(c *serviceCommand, writer http.ResponseWriter, request *http.Request) {
	config := make(map[string]interface{})
	if err := readConfiguration(configurationPaths, config); err == nil {
		sendHTTPResponse(writer, makeResponse(http.StatusOK, config))
	} else {
		resp := makeErrorResponse(http.StatusInternalServerError, "Failed to read configuration data", "internal-error")
		sendHTTPResponse(writer, resp)
	}
}

func postConfiguration(c *serviceCommand, writer http.ResponseWriter, request *http.Request) {
	path := getConfigOnPath(os.Getenv("SNAP_DATA"))
	config := make(map[string]interface{})
	if readConfiguration([]string{path}, config) != nil {
		resp := makeErrorResponse(http.StatusInternalServerError,
			"Failed to read existing configuration file", "internal-error")
		sendHTTPResponse(writer, resp)
		return
	}

	if validTokens == nil || len(validTokens) == 0 {
		errResponse := makeErrorResponse(http.StatusInternalServerError, "No default configuration file available", "internal-error")
		sendHTTPResponse(writer, errResponse)
		return
	}

	file, err := os.Create(path)
	if err != nil {
		resp := makeErrorResponse(http.StatusInternalServerError, "Can't write configuration file", "internal-error")
		sendHTTPResponse(writer, resp)
		return
	}
	defer file.Close()

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		resp := makeErrorResponse(http.StatusInternalServerError, "Error reading the request body", "internal-error")
		sendHTTPResponse(writer, resp)
		return
	}

	var items map[string]interface{}
	if err = json.Unmarshal(body, &items); err != nil {
		resp := makeErrorResponse(http.StatusInternalServerError, "Malformed request", "internal-error")
		sendHTTPResponse(writer, resp)
		return
	}

	// Add the items in the config, but only if all are in the whitelist
	for key, value := range items {
		if _, present := validTokens[key]; !present {
			errResponse := makeErrorResponse(http.StatusInternalServerError, `Invalid key "`+key+`"`, "internal-error")
			sendHTTPResponse(writer, errResponse)
			return
		}
		config[key] = value
	}

	for key, value := range config {
		key = convertKeyToStorageFormat(key)
		value = escapeTextForShell(value)
		file.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	if err := restartAccessPoint(c); err != nil {
		resp := makeErrorResponse(http.StatusInternalServerError, "Failed to restart AP process", "internal-error")
		sendHTTPResponse(writer, resp)
		return
	}

	sendHTTPResponse(writer, makeResponse(http.StatusOK, nil))
}

func restartAccessPoint(c *serviceCommand) error {
	if c.s.ap != nil {
		// Now that we have all configuration changes successfully applied
		// we can safely restart the service.
		if err := c.s.ap.Restart(); err != nil {
			return err
		}
	}
	return nil
}

func getStatus(c *serviceCommand, writer http.ResponseWriter, request *http.Request) {
	status := make(map[string]interface{})

	status["ap.active"] = false
	if c.s.ap != nil && c.s.ap.Running() {
		status["ap.active"] = true
	}

	sendHTTPResponse(writer, makeResponse(http.StatusOK, status))
}

func postStatus(c *serviceCommand, writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		resp := makeErrorResponse(http.StatusInternalServerError, "Error reading the request body", "internal-error")
		sendHTTPResponse(writer, resp)
		return
	}

	var items map[string]string
	if json.Unmarshal(body, &items) != nil {
		resp := makeErrorResponse(http.StatusInternalServerError, "Malformed request", "internal-error")
		sendHTTPResponse(writer, resp)
		return
	}

	action, ok := items["action"]
	if !ok {
		resp := makeErrorResponse(http.StatusInternalServerError, "Mailformed request", "internal-error")
		sendHTTPResponse(writer, resp)
		return
	}

	switch action {
	case "restart-ap":
		if err = restartAccessPoint(c); err != nil {
			resp := makeErrorResponse(http.StatusInternalServerError, "Failed to restart AP process", "internal-error")
			sendHTTPResponse(writer, resp)
			return
		}

		resp := makeResponse(http.StatusOK, nil)
		sendHTTPResponse(writer, resp)
	}

	resp := makeErrorResponse(http.StatusInternalServerError, "Invalid request", "internal-error")
	sendHTTPResponse(writer, resp)
}
