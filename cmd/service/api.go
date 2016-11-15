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
	"log"
	"net/http"
	"os"
)

var api = []*serviceCommand{
	configurationCmd,
}

var (
	configurationCmd = &serviceCommand{
		Path: "/v1/configuration",
		GET:  getConfiguration,
		POST: changeConfiguration,
	}
	validTokens map[string]bool
)

func getConfiguration(c *serviceCommand, writer http.ResponseWriter, request *http.Request) {
	config := make(map[string]string)
	if err := readConfiguration(configurationPaths, config); err == nil {
		sendHTTPResponse(writer, makeResponse(http.StatusOK, config))
	} else {
		log.Println("Read configuration failed:", err)
		errResponse := makeErrorResponse(http.StatusInternalServerError, "Failed to read configuration data", "internal-error")
		sendHTTPResponse(writer, errResponse)
	}
}

func changeConfiguration(c *serviceCommand, writer http.ResponseWriter, request *http.Request) {
	path := getConfigOnPath(os.Getenv("SNAP_DATA"))
	config := map[string]string{}
	if readConfiguration([]string{path}, config) != nil {
		errResponse := makeErrorResponse(http.StatusInternalServerError,
			"Failed to read existing configuration file", "internal-error")
		sendHTTPResponse(writer, errResponse)
		return
	}

	if validTokens == nil || len(validTokens) == 0 {
		errResponse := makeErrorResponse(http.StatusInternalServerError, "No default configuration file available", "internal-error")
		sendHTTPResponse(writer, errResponse)
		return
	}

	file, err := os.Create(path)
	if err != nil {
		log.Printf("Write to %q failed: %v\n", path, err)
		errResponse := makeErrorResponse(http.StatusInternalServerError, "Can't write configuration file", "internal-error")
		sendHTTPResponse(writer, errResponse)
		return
	}
	defer file.Close()

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Println("Failed to process incoming configuration change request:", err)
		errResponse := makeErrorResponse(http.StatusInternalServerError, "Error reading the request body", "internal-error")
		sendHTTPResponse(writer, errResponse)
		return
	}

	var items map[string]string
	if err = json.Unmarshal(body, &items); err != nil {
		log.Println("Invalid input data", err)
		errResponse := makeErrorResponse(http.StatusInternalServerError, "Malformed request", "internal-error")
		sendHTTPResponse(writer, errResponse)
		return
	}

	// Add the items in the config, but only if all are in the whitelist
	for key, value := range items {
		if _, present := validTokens[key]; !present {
			log.Println(`Invalid key "` + key + `": ignoring request`)
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

	/*
		fmt.Println("ap", ap)
		if ap != nil {
			// Now that we have all configuration changes successfully applied
			// we can safely restart the service.
			if err := ap.Restart(); err != nil {
				log.Println("error: ", err)
				response := makeErrorResponse(http.StatusInternalServerError, "Failed to restart AP process", "internal-error")
				sendHTTPResponse(writer, response)
				return
			}
		}
	*/

	sendHTTPResponse(writer, makeResponse(http.StatusOK, nil))
}
