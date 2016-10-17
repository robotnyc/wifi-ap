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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Utility function to read input and strip the trailing \n
func readUserInput(reader *bufio.Reader) string {
	ret, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimRight(ret, "\n")
}

// Helper function to list network interfaces
func findExistingInterfaces(wifi bool) (wifis []string) {
	// Use sysfs to get interfaces list
	const sysNetPath = "/sys/class/net/"
	ifaces, err := ioutil.ReadDir(sysNetPath)
	wifis = []string{}
	if err == nil {
		for _, iface := range ifaces {
			if iface.Name() != "lo" {
				// The "wireless" subdirectory exists only for wireless interfaces
				if _, err := os.Stat(sysNetPath + iface.Name() + "/wireless"); os.IsNotExist(err) != wifi {
					wifis = append(wifis, iface.Name())
				}
			}
		}
	}

	return
}

type wizardStep func(map[string]string, *bufio.Reader) error

var allSteps = [...]wizardStep{
	// determine the WiFi interface
	func(configuration map[string]string, reader *bufio.Reader) error {
		ifaces := findExistingInterfaces(true)
		if len(ifaces) == 0 {
			return fmt.Errorf("There are no valid wireless network interfaces available")
		} else if len(ifaces) == 1 {
			fmt.Println("Automatically selected only available wireless network interface " + ifaces[0])
			return nil
		}
		fmt.Print("Which wireless interface you want to use? ")
		ifacesVerb := "are"
		if len(ifaces) == 1 {
			ifacesVerb = "is"
		}
		fmt.Print("Available " + ifacesVerb + " " + strings.Join(ifaces, ", ") + ": ")
		iface := readUserInput(reader)
		if re := regexp.MustCompile("^[[:alnum:]]+$"); !re.MatchString(iface) {
			return fmt.Errorf("Invalid interface name '%s' given", iface)
		}
		configuration["wifi.interface"] = iface

		return nil
	},

	// Ask for WiFi ESSID
	func(configuration map[string]string, reader *bufio.Reader) error {
		fmt.Print("Insert the ESSID of your access point: ")
		iface := readUserInput(reader)
		if len(iface) == 0 || len(iface) > 31 {
			return fmt.Errorf("ESSID length must be between 1 and 31 characters")
		}
		configuration["wifi.essid"] = iface

		return nil
	},

	// Select WiFi encryption type
	func(configuration map[string]string, reader *bufio.Reader) error {
		fmt.Print("Do you want to protect your network with a WPA2 password instead of staying open for everyone? (y/n) ")
		switch resp := strings.ToLower(readUserInput(reader)); resp {
		case "y":
			configuration["wifi.security"] = "wpa2"
		case "n":
			configuration["wifi.security"] = "open"
		default:
			return fmt.Errorf("Invalid answer: %s", resp)
		}

		return nil
	},

	// If WPA2 is set, ask for valid password
	func(configuration map[string]string, reader *bufio.Reader) error {
		if configuration["wifi.security"] == "open" {
			return nil
		}
		fmt.Print("Insert your WPA2 passphrase: ")
		key := readUserInput(reader)
		if len(key) < 8 || len(key) > 63 {
			return fmt.Errorf("WPA2 passphrase must be between 8 and 63 characters")
		}
		configuration["wifi.passphrase"] = key

		return nil
	},

	// Configure WiFi AP IP address
	func(configuration map[string]string, reader *bufio.Reader) error {
		fmt.Print("Insert the Access Point IP address: ")
		inputIp := readUserInput(reader)
		ipv4 := net.ParseIP(inputIp)
		if ipv4 == nil {
			return fmt.Errorf("Invalid IP address: %s", inputIp)
		}
		if !ipv4.IsGlobalUnicast() {
			return fmt.Errorf("%s is a reserved IPv4 address", inputIp)
		}
		if ipv4.To4() == nil {
			return fmt.Errorf("%s is not an IPv4 address", inputIp)
		}

		configuration["wifi.address"] = inputIp

		nmask := ipv4.DefaultMask()
		configuration["wifi.netmask"] = fmt.Sprintf("%d.%d.%d.%d", nmask[0], nmask[1], nmask[2], nmask[3])

		return nil
	},

	// Configure the DHCP pool
	func(configuration map[string]string, reader *bufio.Reader) error {
		var maxpoolsize byte
		ipv4 := net.ParseIP(configuration["wifi.address"])
		if ipv4[15] <= 128 {
			maxpoolsize = 254 - ipv4[15]
		} else {
			maxpoolsize = ipv4[15] - 1
		}

		fmt.Printf("How many host do you want your DHCP pool to hold to? (1-%d) ", maxpoolsize)
		input := readUserInput(reader)
		inputhost, err := strconv.ParseUint(input, 10, 8)
		if err != nil {
			return fmt.Errorf("Invalid answer: %s", input)
		}
		if byte(inputhost) > maxpoolsize {
			return fmt.Errorf("%d is bigger than the maximum pool size %d", inputhost, maxpoolsize)
		}

		nhosts := byte(inputhost)
		// Allocate the pool in the bigger half, trying to avoid overlap with access point IP
		if ipv4[15] <= 128 {
			configuration["dhcp.range-start"] = fmt.Sprintf("%d.%d.%d.%d", ipv4[12], ipv4[13], ipv4[14], ipv4[15]+1)
			configuration["dhcp.range-stop"] = fmt.Sprintf("%d.%d.%d.%d", ipv4[12], ipv4[13], ipv4[14], ipv4[15]+nhosts)
		} else {
			configuration["dhcp.range-start"] = fmt.Sprintf("%d.%d.%d.%d", ipv4[12], ipv4[13], ipv4[14], ipv4[15]-nhosts)
			configuration["dhcp.range-stop"] = fmt.Sprintf("%d.%d.%d.%d", ipv4[12], ipv4[13], ipv4[14], ipv4[15]-1)
		}

		return nil
	},

	// Select the wired interface to share
	func(configuration map[string]string, reader *bufio.Reader) error {
		ifaces := findExistingInterfaces(false)
		if len(ifaces) == 0 {
			fmt.Println("No network interface available which's connection can be shared. Disabling connection sharing.")
			configuration["share.disabled"] = "1"
			return nil
		}
		ifacesVerb := "are"
		if len(ifaces) == 1 {
			ifacesVerb = "is"
		}
		fmt.Println("Which network interface you want to use for connection sharing?")
		fmt.Print("Available " + ifacesVerb + " " + strings.Join(ifaces, ", ") + ": ")
		iface := readUserInput(reader)
		if re := regexp.MustCompile("^[[:alnum:]]+$"); !re.MatchString(iface) {
			return fmt.Errorf("Invalid interface name '%s' given", iface)
		}
		configuration["share.network-interface"] = iface

		return nil
	},
}

// Use the REST API to set the configuration
func applyConfiguration(configuration map[string]string) error {
	json, err := json.Marshal(configuration)
	if err != nil {
		return err
	}

	response, err := sendHTTPRequest(getServiceConfigurationURI(), "POST", bytes.NewReader(json))
	if err != nil {
		return err
	}

	if response.StatusCode == http.StatusOK && response.Status == http.StatusText(http.StatusOK) {
		fmt.Println("Configuration applied succesfully")
		return nil
	} else {
		return fmt.Errorf("Failed to set configuration, service returned: %d (%s)\n", response.StatusCode, response.Status)
	}
}

type wizardCommand struct{}

func (cmd *wizardCommand) Execute(args []string) error {
	// Setup some sane default values, we don't ask the user for everything
	configuration := map[string]string{
		"disabled":            "0",
		"wifi.channel":        "6",
		"wifi.operation-mode": "g",
		"dhcp.lease-time":     "12h",
	}

	reader := bufio.NewReader(os.Stdin)

	for _, step := range allSteps {
		for {
			if err := step(configuration, reader); err != nil {
				fmt.Println("Error: ", err)
				fmt.Print("You want to try again? (y/n) ")
				answer := readUserInput(reader)
				if strings.ToLower(answer) != "y" {
					return err
				}
			} else {
				// Good answer
				break
			}
		}
	}

	return applyConfiguration(configuration)
}

func init() {
	addCommand("wizard", "Start the interactive wizard configuration", "", &wizardCommand{})
}
