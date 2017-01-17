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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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

func readConfigurationFile(filePath string, config map[string]interface{}) (err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	defer file.Close()

	for scanner := bufio.NewScanner(file); scanner.Scan(); {
		// Ignore all empty or commented lines
		if line := scanner.Text(); len(line) != 0 && line[0] != '#' {
			// Line must be in the KEY=VALUE format
			if i := strings.IndexRune(line, '='); i > 0 {
				var value interface{} = unescapeTextByShell(line[i+1:])
				switch (value) {
				case "true":
					value = true
				case "false":
					value = false
				}
				config[convertKeyToRepresentationFormat(line[:i])] = value
			}
		}
	}

	return nil
}

func readConfiguration(paths []string, config map[string]interface{}) (err error) {
	for _, location := range paths {
		if readConfigurationFile(location, config) != nil {
			return fmt.Errorf("Failed to read configuration file '%s'", location)
		}
	}

	return nil
}

// Escape shell special characters, avoid injection
// eg. SSID set to "My AP$(nc -lp 2323 -e /bin/sh)"
// to get a root shell
func escapeTextForShell(input interface{}) string {
	data := fmt.Sprint(input)
	if strings.ContainsAny(data, "\\\"'`$\n\t #") {
		data = strings.Replace(data, `\`, `\\`, -1)
		data = strings.Replace(data, `"`, `\"`, -1)
		data = strings.Replace(data, "`", "\\`", -1)
		data = strings.Replace(data, `$`, `\$`, -1)
		data = `"` + data + `"`
	}
	return data
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
