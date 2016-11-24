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
	"sort"
)

func printMapSorted(m map[string]string) {
	sortedKeys := make([]string, 0, len(m))
	for key, _ := range m {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	for n := range sortedKeys {
		fmt.Fprintf(os.Stdout, "%s: %s\n", sortedKeys[n], m[sortedKeys[n]])
	}
}
