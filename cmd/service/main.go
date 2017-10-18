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
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	s := &service{}

	// Wait until the configure hook, which is called on snap
	// installation, marked us as successfully setup. If we
	// continue before that happen we will miss any initial
	// configuration set via a gadget snap.
	for {
		_, err := os.Stat(os.Getenv("SNAP_COMMON") + "/.setup_done")
		if err == nil {
			break
		}
		time.Sleep(time.Second/2)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func(s *service) {
		_ = <-c
		s.Shutdown()
	}(s)

	if err := s.Run(); err != nil {
		log.Fatalf("Failed to start service: %s", err)
	}
}
