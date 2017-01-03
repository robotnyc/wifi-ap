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
	"fmt"
	"os"
	"os/user"

	"github.com/jessevdk/go-flags"
)

type commonOptions struct {
	Verbose []bool `short:"v" long:"verbose" description:"Verbose output"`
}

var parser = flags.NewParser(&commonOptions{}, flags.Default)

func addCommand(name string, shortHelp string, longHelp string, data interface{}) (*flags.Command, error) {
	cmd, err := parser.AddCommand(name, shortHelp, longHelp, data)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func main() {
	user, err := user.Current()
	if err == nil {
		if user.Uid != "0" {
			fmt.Println("ERROR: You need to execute this command as root to be allowed to")
			fmt.Println("talk to the service. Run")
			fmt.Println(" $ sudo wifi-ap.config get")
			fmt.Println("for example.")
			os.Exit(1)
		}
	}

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}
