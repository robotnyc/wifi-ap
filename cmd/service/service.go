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
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"gopkg.in/tomb.v2"
)

const socketPathSuffix = "sockets/control"

type responceFunc func(*serviceCommand, http.ResponseWriter, *http.Request)

type serviceCommand struct {
	Path   string
	GET    responceFunc
	PUT    responceFunc
	POST   responceFunc
	DELETE responceFunc
	s      *service
}

type service struct {
	tomb     tomb.Tomb
	server   *http.Server
	listener net.Listener
	router   *mux.Router
	ap       BackgroundProcess
}

func (c *serviceCommand) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var rspf responceFunc

	switch r.Method {
	case "GET":
		rspf = c.GET
	case "PUT":
		rspf = c.PUT
	case "POST":
		rspf = c.POST
	case "DELETE":
		rspf = c.DELETE
	}

	if rspf == nil {
		rsp := makeErrorResponse(http.StatusInternalServerError, "Invalid method called", "internal-error")
		sendHTTPResponse(w, rsp)
		return
	}

	rspf(c, w, r)
}

func (s *service) addRoutes() {
	s.router = mux.NewRouter()

	for _, c := range api {
		c.s = s
		log.Println("Adding route for ", c.Path)
		s.router.Handle(c.Path, c).Name(c.Path)
	}
}

func (s *service) setupAccesPoint() error {
	path := path.Join(os.Getenv("SNAP"), "bin", "ap.sh")
	ap, err := NewBackgroundProcess(path)
	if err != nil {
		return err
	}

	s.ap = ap
	err = s.ap.Start()
	if err != nil {
		return err
	}

	return nil
}

func (s *service) Shutdown() {
	log.Println("Shutting down ...")
	s.listener.Close()
	s.tomb.Kill(nil)
	s.tomb.Wait()
}

func (s *service) Run() error {
	s.addRoutes()
	if err := s.setupAccesPoint(); err != nil {
		return err
	}

	var err error
	if validTokens, err = loadValidTokens(filepath.Join(os.Getenv("SNAP"), "conf", "default-config")); err != nil {
		log.Println("Failed to read default configuration:", err)
	}

	// Create the socket directory and remove any stale socket
	path := filepath.Join(os.Getenv("SNAP_DATA"), socketPathSuffix)
	if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
		os.Mkdir(filepath.Dir(path), 0755)
	}
	os.Remove(path)

	s.server = &http.Server{Handler: s.router}
	s.listener, err = net.Listen("unix", path)
	if err != nil {
		return err
	}

	s.tomb.Go(func() error {
		err := s.server.Serve(s.listener)
		if err != nil {
			return fmt.Errorf("Failed to server HTTP: %s", err)
		}
		return nil
	})

	s.tomb.Wait()

	if s.ap.Running() {
		s.ap.Stop()
	}

	return nil
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
