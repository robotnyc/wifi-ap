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
	"net/http"
	"os"
	"path/filepath"
)

const (
	serviceAddress = "127.0.0.1"
	servicePort    = 5005
)

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
	router *mux.Router
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

func (s *service) Run() error {
	s.addRoutes()

	var err error
	if validTokens, err = loadValidTokens(filepath.Join(os.Getenv("SNAP"), "conf", "default-config")); err != nil {
		log.Println("Failed to read default configuration:", err)
	}

	addr := fmt.Sprintf("%s:%d", serviceAddress, servicePort)
	err = http.ListenAndServe(addr, s.router)
	if err != nil {
		return err
	}

	return nil
}
