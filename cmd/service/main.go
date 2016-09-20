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
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

/* JSON message format, as described here:
{
	"result": {
		"key" : "val"
	},
	"status": "OK",
	"status-code": 200,
	"type": "sync"
}
*/

type serviceResponse struct {
	Result     map[string]string `json:"result"`
	Status     string            `json:"status"`
	StatusCode int               `json:"status-code"`
	Type       string            `json:"type"`
}

func makeErrorResponse(code int, message, kind string) *serviceResponse {
	return &serviceResponse{
		Type:       "error",
		Status:     http.StatusText(code),
		StatusCode: code,
		Result: map[string]string{
			"message": message,
			"kind":    kind,
		},
	}
}

func makeResponse(status int, result map[string]string) *serviceResponse {
	resp := &serviceResponse{
		Type:       "sync",
		Status:     http.StatusText(status),
		StatusCode: status,
		Result:     result,
	}

	if resp.Result == nil {
		resp.Result = make(map[string]string)
	}

	return resp
}

func sendHTTPResponse(writer http.ResponseWriter, response *serviceResponse) {
	writer.WriteHeader(response.StatusCode)
	data, _ := json.Marshal(response)
	fmt.Fprintln(writer, string(data))
}

func getConfigOnPath(path string) string {
	return filepath.Join(path, "config")
}

// Array of paths where the config file can be found.
// The first one is readonly, the others are writable
// they are readed in order and the configuration is merged
var configurationPaths = []string{
	filepath.Join(os.Getenv("SNAP"), "conf", "default-config"),
	getConfigOnPath(os.Getenv("SNAP_DATA")),
	getConfigOnPath(os.Getenv("SNAP_USER_DATA"))}

const (
	servicePort        = 5005
	configurationV1Uri = "/v1/configuration"
)

func main() {
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc(configurationV1Uri, getConfiguration).Methods(http.MethodGet)
	r.HandleFunc(configurationV1Uri, changeConfiguration).Methods(http.MethodPost)

	log.Fatal(http.ListenAndServe("127.0.0.1:"+strconv.Itoa(servicePort), r))
}

// Convert eg. WIFI_OPERATION_MODE to wifi.operation-mode
func convertKeyToRepresentationFormat(key string) string {
	newKey := strings.ToLower(key)
	newKey = strings.Replace(newKey, "_", ".", 1)
	return strings.Replace(newKey, "_", "-", -1)
}

func convertKeyToStorageFormat(key string) string {
	// Convert eg. wifi.operation-mode to WIFI_OPERATION_MODE
	newKey := strings.ToUpper(key)
	newKey = strings.Replace(newKey, ".", "_", -1)
	return strings.Replace(newKey, "-", "_", -1)
}

func readConfigurationFile(filePath string, config map[string]string) (err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	defer file.Close()

	for scanner := bufio.NewScanner(file); scanner.Scan(); {
		// Ignore all empty or commented lines
		if line := scanner.Text(); len(line) != 0 && line[0] != '#' {
			// Line must be in the KEY=VALUE format
			if parts := strings.Split(line, "="); len(parts) == 2 {
				value := unescapeTextByShell(parts[1])
				config[convertKeyToRepresentationFormat(parts[0])] = value
			}
		}
	}

	return nil
}

func readConfiguration(paths []string, config map[string]string) (err error) {
	for _, location := range paths {
		if readConfigurationFile(location, config) != nil {
			return fmt.Errorf("Failed to read configuration file '%s'", location)
		}
	}

	return nil
}

func getConfiguration(writer http.ResponseWriter, request *http.Request) {
	config := make(map[string]string)
	if readConfiguration(configurationPaths, config) == nil {
		sendHTTPResponse(writer, makeResponse(http.StatusOK, config))
	} else {
		errResponse := makeErrorResponse(http.StatusInternalServerError, "Failed to read configuration data", "internal-error")
		sendHTTPResponse(writer, errResponse)
	}
}

// Escape shell special characters, avoid injection
// eg. SSID set to "My AP$(nc -lp 2323 -e /bin/sh)"
// to get a root shell
func escapeTextForShell(input string) string {
	if strings.ContainsAny(input, "\\\"'`$\n\t #") {
		input = strings.Replace(input, `\`, `\\`, -1)
		input = strings.Replace(input, `"`, `\"`, -1)
		input = strings.Replace(input, "`", "\\`", -1)
		input = strings.Replace(input, `$`, `\$`, -1)

		input = `"` + input + `"`
	}
	return input
}

// Do the reverse of escapeTextForShell() here
// strip any \ followed by \$`"
func unescapeTextByShell(input string) string {
	input = strings.Trim(input, `"'`)
	if strings.ContainsAny(input, "\\") {
		re := regexp.MustCompile("\\\\([\\\\$\\`\\\"])")
		input = re.ReplaceAllString(input, "$1")
	}
	return input
}

func changeConfiguration(writer http.ResponseWriter, request *http.Request) {
	// Write in SNAP_DATA
	confWrite := getConfigOnPath(os.Getenv("SNAP_DATA"))

	file, err := os.Create(confWrite)
	if err != nil {
		errResponse := makeErrorResponse(http.StatusInternalServerError, "Can't write configuration file", "internal-error")
		sendHTTPResponse(writer, errResponse)
		return
	}
	defer file.Close()

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		errResponse := makeErrorResponse(http.StatusInternalServerError, "Error reading the request body", "internal-error")
		sendHTTPResponse(writer, errResponse)
		return
	}

	var items map[string]string
	if json.Unmarshal(body, &items) != nil {
		errResponse := makeErrorResponse(http.StatusInternalServerError, "Malformed request", "internal-error")
		sendHTTPResponse(writer, errResponse)
		return
	}

	for key, value := range items {
		key = convertKeyToStorageFormat(key)
		value = escapeTextForShell(value)
		file.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	sendHTTPResponse(writer, makeResponse(http.StatusOK, nil))
}
