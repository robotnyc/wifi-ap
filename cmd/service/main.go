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
	"path"
	"path/filepath"
	"regexp"
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

func getConfigOnPath(confPath string) string {
	return filepath.Join(confPath, "config")
}

// Array of paths where the config file can be found.
// The first one is readonly, the others are writable
// they are readed in order and the configuration is merged
var configurationPaths = []string{
	filepath.Join(os.Getenv("SNAP"), "conf", "default-config"),
	getConfigOnPath(os.Getenv("SNAP_DATA")),
	getConfigOnPath(os.Getenv("SNAP_USER_DATA"))}

const (
	serviceAddress     = "127.0.0.1"
	servicePort        = 5005
	configurationV1Uri = "/v1/configuration"
)

var validTokens map[string]bool

func loadValidTokens(path string) (map[string]bool, error) {
	def, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer def.Close()

	tokens := map[string]bool{}

	scanner := bufio.NewScanner(def)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// Get the substring before the '='
		if eq := strings.IndexRune(line, '='); eq > 0 {
			// Add the token to the whitelist, converted in our format
			tokens[convertKeyToRepresentationFormat(line[:eq])] = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}

var apProcess *backgroundProcess = nil

func main() {
	path := path.Join(os.Getenv("SNAP"), "bin", "ap.sh")
	apProcess, err := NewBackgroundProcess(path)
	if err != nil {
		// If the creation of the process failed we don't shutdown
		// but will error out whenever the user performs a operation
		// which would require a valid process instance.
		log.Println("Failed to create background process for access point instance")
	}

	if apProcess != nil {
		apProcess.Start()
	}

	var err error
	if validTokens, err = loadValidTokens(filepath.Join(os.Getenv("SNAP"), "conf", "default-config")); err != nil {
		log.Println("Failed to read default configuration:", err)
	}

	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc(configurationV1Uri, getConfiguration).Methods(http.MethodGet)
	r.HandleFunc(configurationV1Uri, changeConfiguration).Methods(http.MethodPost)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", serviceAddress, servicePort), r))
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
				key := convertKeyToRepresentationFormat(parts[0])
				value := unescapeTextByShell(parts[1])
				config[key] = value
			}
		}
	}

	return nil
}

func readConfiguration(paths []string, config map[string]string) (err error) {
	for _, location := range paths {
		if readConfigurationFile(location, config) != nil {
			return fmt.Errorf(`Failed to read configuration file "%s"`, location)
		}
	}

	return nil
}

func getConfiguration(writer http.ResponseWriter, request *http.Request) {
	config := make(map[string]string)
	if err := readConfiguration(configurationPaths, config); err == nil {
		sendHTTPResponse(writer, makeResponse(http.StatusOK, config))
	} else {
		log.Println("Read configuration failed:", err)
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

	if apProcess != nil {
		// Now that we have all configuration changes successfully applied
		// we can safely restart the service.
		if err := apProcess.Restart(); err != nil {
			response := makeErrorResponse(http.StatusInternalServerError, "Failed to restart AP process", "internal-error")
			sendHTTPResponse(writer, response)
			return
		}
	}

	sendHTTPResponse(writer, makeResponse(http.StatusOK, nil))
}
