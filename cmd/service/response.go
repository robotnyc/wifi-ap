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
	"encoding/json"
	"fmt"
	"net/http"
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
