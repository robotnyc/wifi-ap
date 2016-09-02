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

type Response struct {
	Result     map[string]string `json:"result"`
	Status     string            `json:"status"`
	StatusCode int               `json:"status-code"`
	Type       string            `json:"type"`
}

func makeErrorResponse(code int, message, kind string) Response {
	return Response{
		Type:       "error",
		Status:     http.StatusText(code),
		StatusCode: code,
		Result: map[string]string{
			"message": message,
			"kind":    kind,
		},
	}
}

func makeResponse(status int, result map[string]string) Response {
	return Response{
		Type:       "sync",
		Status:     http.StatusText(status),
		StatusCode: status,
		Result:     result,
	}
}

func sendHttpResponse(writer http.ResponseWriter, response Response) {
	writer.WriteHeader(response.StatusCode)
	data, _ := json.Marshal(response)
	fmt.Fprintln(writer, string(data))
}

func getConfigOnPath(path string) string {
	return path + "/config"
}

// Array of paths where the config file can be found.
// The first one is readonly, the others are writable
// they are readed in order and the configuration is merged
var cfgpaths []string = []string{getConfigOnPath(os.Getenv("SNAP")),
	getConfigOnPath(os.Getenv("SNAP_DATA")), getConfigOnPath(os.Getenv("SNAP_USER_DATA"))}

const (
	PORT                  = 5005
	CONFIGURATION_API_URI = "/v1/configuration"
)

func main() {
	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc(CONFIGURATION_API_URI, getConfiguration).Methods(http.MethodGet)
	r.HandleFunc(CONFIGURATION_API_URI, changeConfiguration).Methods(http.MethodPost)

	log.Fatal(http.ListenAndServe("127.0.0.1:"+strconv.Itoa(PORT), r))
}

// Convert eg. WIFI_OPERATION_MODE to wifi.operation-mode
func convertKeyToRepresentationFormat(key string) string {
	new_key := strings.ToLower(key)
	new_key = strings.Replace(new_key, "_", ".", 1)
	return strings.Replace(new_key, "_", "-", -1)
}

func convertKeyToStorageFormat(key string) string {
	// Convert eg. wifi.operation-mode to WIFI_OPERATION_MODE
	key = strings.ToUpper(key)
	key = strings.Replace(key, ".", "_", -1)
	return strings.Replace(key, "-", "_", -1)
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
	if readConfiguration(cfgpaths, config) == nil {
		sendHttpResponse(writer, makeResponse(http.StatusOK, config))
	} else {
		errResponse := makeErrorResponse(http.StatusInternalServerError, "Failed to read configuration data", "internal-error")
		sendHttpResponse(writer, errResponse)
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
		sendHttpResponse(writer, errResponse)
		return
	}
	defer file.Close()

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		errResponse := makeErrorResponse(http.StatusInternalServerError, "Error reading the request body", "internal-error")
		sendHttpResponse(writer, errResponse)
		return
	}

	var items map[string]string
	if json.Unmarshal(body, &items) != nil {
		errResponse := makeErrorResponse(http.StatusInternalServerError, "Malformed request", "internal-error")
		sendHttpResponse(writer, errResponse)
		return
	}

	for key, value := range items {
		key = convertKeyToStorageFormat(key)
		value = escapeTextForShell(value)
		file.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	sendHttpResponse(writer, makeResponse(http.StatusOK, nil))
}
