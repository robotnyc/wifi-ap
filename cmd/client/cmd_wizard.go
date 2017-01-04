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
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type wizardStep func(map[string]interface{}, *bufio.Reader, bool) error

// Go stores both IPv4 and IPv6 as [16]byte
// with IPv4 addresses stored in the end of the buffer
// in bytes 13..16
const ipv4Offset = net.IPv6len - net.IPv4len

var defaultIp = net.IPv4(10, 0, 60, 1)

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
	const sysNetPath = "/sys/class/net"
	ifaces, err := net.Interfaces()
	wifis = []string{}
	if err == nil {
		for _, iface := range ifaces {
			if iface.Flags&net.FlagLoopback == 0 {
				// The "wireless" subdirectory exists only for wireless interfaces
				if _, err := os.Stat(filepath.Join(sysNetPath, iface.Name, "wireless")); os.IsNotExist(err) != wifi {
					wifis = append(wifis, iface.Name)
				}
			}
		}
	}

	return
}

func findFreeSubnet(startIp net.IP) (net.IP, error) {
	curIp := startIp

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	// Start from startIp and increment the third octect every time
	// until we found an IP which isn't assigned to any interface
	for found := false; !found; {
		// Cycle through all the assigned addresses
		for _, addr := range addrs {
			found = true
			_, subnet, _ := net.ParseCIDR(addr.String())
			// If busy, increment the third octect and retry
			if subnet.Contains(curIp) {
				found = false
				if curIp[ipv4Offset+2] == 255 {
					return nil, fmt.Errorf("No free netmask found")
				}
				curIp[ipv4Offset+2]++
				break
			}
		}
	}

	return curIp, nil
}

var allSteps = [...]wizardStep{
	// determine the WiFi interface
	func(configuration map[string]interface{}, reader *bufio.Reader, nonInteractive bool) error {
		ifaces := findExistingInterfaces(true)
		if len(ifaces) == 0 {
			return fmt.Errorf("There are no valid wireless network interfaces available")
		} else if len(ifaces) == 1 {
			fmt.Println("Automatically selected only available wireless network interface " + ifaces[0])
			return nil
		}

		if nonInteractive {
			fmt.Println("Selecting interface", ifaces[0])
			configuration["wifi.interface"] = ifaces[0]
			return nil
		}

		fmt.Print("Which wireless interface you want to use? ")
		ifacesVerb := "are"
		if len(ifaces) == 1 {
			ifacesVerb = "is"
		}
		fmt.Printf("Available %s %s: ", ifacesVerb, strings.Join(ifaces, ", "))
		iface := readUserInput(reader)
		if re := regexp.MustCompile("^[[:alnum:]]+$"); !re.MatchString(iface) {
			return fmt.Errorf("Invalid interface name '%s' given", iface)
		}
		configuration["wifi.interface"] = iface

		return nil
	},

	// Ask for WiFi ESSID
	func(configuration map[string]interface{}, reader *bufio.Reader, nonInteractive bool) error {
		if nonInteractive {
			configuration["wifi.ssid"] = "Ubuntu"
			return nil
		}

		fmt.Print("Which SSID you want to use for the access point: ")
		iface := readUserInput(reader)
		if len(iface) == 0 || len(iface) > 31 {
			return fmt.Errorf("ESSID length must be between 1 and 31 characters")
		}
		configuration["wifi.ssid"] = iface

		return nil
	},

	// Select WiFi encryption type
	func(configuration map[string]interface{}, reader *bufio.Reader, nonInteractive bool) error {
		if nonInteractive {
			configuration["wifi.security"] = "open"
			return nil
		}

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
	func(configuration map[string]interface{}, reader *bufio.Reader, nonInteractive bool) error {
		if configuration["wifi.security"] == "open" {
			return nil
		}
		fmt.Print("Please enter the WPA2 passphrase: ")
		key := readUserInput(reader)
		if len(key) < 8 || len(key) > 63 {
			return fmt.Errorf("WPA2 passphrase must be between 8 and 63 characters")
		}
		configuration["wifi.security-passphrase"] = key

		return nil
	},

	// Configure WiFi AP IP address
	func(configuration map[string]interface{}, reader *bufio.Reader, nonInteractive bool) error {
		if nonInteractive {
			wifiIp, err := findFreeSubnet(defaultIp)
			if err != nil {
				return err
			}

			fmt.Println("AccessPoint IP set to", wifiIp.String())

			configuration["wifi.address"] = wifiIp.String()
			configuration["wifi.netmask"] = "255.255.255.0"
			return nil
		}

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
		configuration["wifi.netmask"] = ipv4.DefaultMask().String()

		return nil
	},

	// Configure the DHCP pool
	func(configuration map[string]interface{}, reader *bufio.Reader, nonInteractive bool) error {
		if nonInteractive {
			wifiIp := net.ParseIP(configuration["wifi.address"].(string))

			// Set the DCHP in the range 2..199 with 198 total hosts
			// leave about 50 hosts outside the pool for static addresses
			// wifiIp[ipv4Offset + 3] is the last octect of the IP address
			wifiIp[ipv4Offset+3] = 2
			configuration["dhcp.range-start"] = wifiIp.String()
			wifiIp[ipv4Offset+3] = 199
			configuration["dhcp.range-stop"] = wifiIp.String()

			return nil
		}

		var maxpoolsize byte
		ipv4 := net.ParseIP(configuration["wifi.address"].(string))
		if ipv4[ipv4Offset+3] <= 128 {
			maxpoolsize = 254 - ipv4[ipv4Offset+3]
		} else {
			maxpoolsize = ipv4[ipv4Offset+3] - 1
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
		if octect3 := ipv4[ipv4Offset+3]; octect3 <= 128 {
			ipv4[ipv4Offset+3] = octect3 + 1
			configuration["dhcp.range-start"] = ipv4.String()
			ipv4[ipv4Offset+3] = octect3 + nhosts
			configuration["dhcp.range-stop"] = ipv4.String()
		} else {
			ipv4[ipv4Offset+3] = octect3 - nhosts
			configuration["dhcp.range-start"] = ipv4.String()
			ipv4[ipv4Offset+3] = octect3 - 1
			configuration["dhcp.range-stop"] = ipv4.String()
		}

		return nil
	},

	// Enable or disable connection sharing
	func(configuration map[string]interface{}, reader *bufio.Reader, nonInteractive bool) error {
		if nonInteractive {
			configuration["share.disabled"] = false
			return nil
		}

		fmt.Print("Do you want to enable connection sharing? (y/n) ")
		switch resp := strings.ToLower(readUserInput(reader)); resp {
		case "y":
			configuration["share.disabled"] = false
		case "n":
			configuration["share.disabled"] = true
		default:
			return fmt.Errorf("Invalid answer: %s", resp)
		}

		return nil
	},

	// Select the wired interface to share
	func(configuration map[string]interface{}, reader *bufio.Reader, nonInteractive bool) error {
		if nonInteractive {
			configuration["share.disabled"] = true

			procNetRoute, err := os.Open("/proc/net/route")
			if err != nil {
				return err
			}
			defer procNetRoute.Close()

			var iface string
			// Using something like math.MaxUint32 causes an overflow on
			// some architectures so lets use just a high enough value
			// for typical systems here.
			minMetric := 100000

			scanner := bufio.NewScanner(procNetRoute)
			// Skip the first line with table header
			scanner.Text()
			for scanner.Scan() {
				route := strings.Fields(scanner.Text())

				if len(route) < 8 {
					continue
				}

				// If we picked the interface already for the AP to operate on
				// ignore it.
				if route[0] == configuration["wifi.interface"] {
					break
				}

				// A /proc/net/route line is in the form:
				// iface destination gateway ...
				// eth1 00000000 0155A8C0 ...
				// look for a 0 destination (0.0.0.0) which is our default route
				metric, err := strconv.Atoi(route[7])
				if err != nil {
					metric = 0
				}

				if route[1] == "00000000" && metric < minMetric {
					iface = route[0]
					minMetric = metric
					break
				}
			}

			if len(iface) == 0 {
				configuration["share.disabled"] = true
			} else {
				fmt.Println("Selecting", iface, "for connection sharing")
				configuration["share.disabled"] = false
				configuration["share.network-interface"] = iface
			}

			return nil
		}

		if configuration["share.disabled"] == true {
			return nil
		}

		ifaces := findExistingInterfaces(false)
		if len(ifaces) == 0 {
			fmt.Println("No network interface available which's connection can be shared. Disabling connection sharing.")
			configuration["share.disabled"] = true
			return nil
		}
		ifacesVerb := "are"
		if len(ifaces) == 1 {
			ifacesVerb = "is"
		}
		fmt.Println("Which network interface you want to use for connection sharing?")
		fmt.Printf("Available %s %s: ", ifacesVerb, strings.Join(ifaces, ", "))
		iface := readUserInput(reader)
		if re := regexp.MustCompile("^[[:alnum:]]+$"); !re.MatchString(iface) {
			return fmt.Errorf("Invalid interface name '%s' given", iface)
		}
		configuration["share.network-interface"] = iface

		return nil
	},

	func(configuration map[string]interface{}, reader *bufio.Reader, nonInteractive bool) error {
		if nonInteractive {
			configuration["disabled"] = false
			return nil
		}

		fmt.Print("Do you want to enable the AP now? (y/n) ")
		switch resp := strings.ToLower(readUserInput(reader)); resp {
		case "y":
			configuration["disabled"] = false

			fmt.Println("In order to get the AP correctly enabled you have to restart the backend service:")
			fmt.Println(" $ systemctl restart snap.wifi-ap.backend")
		case "n":
			configuration["disabled"] = true
		default:
			return fmt.Errorf("Invalid answer: %s", resp)
		}

		return nil
	},
}

// Use the REST API to set the configuration
func applyConfiguration(configuration map[string]interface{}) error {
	for key, value := range configuration {
		log.Printf("%v=%v\n", key, value)
	}
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

type wizardCommand struct {
	Auto bool `long:"auto" description:"Automatically configure the AP"`
}

func (cmd *wizardCommand) Execute(args []string) error {
	// Setup some sane default values, we don't ask the user for everything
	configuration := make(map[string]interface{})

	reader := bufio.NewReader(os.Stdin)

	for _, step := range allSteps {
		for {
			if err := step(configuration, reader, cmd.Auto); err != nil {
				if cmd.Auto {
					return err
				}
				fmt.Println("Error: ", err)
				fmt.Print("Do you want to try again? (y/n) ")
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
