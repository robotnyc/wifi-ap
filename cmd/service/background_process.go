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
	"os/exec"
)

type backgroundProcess struct {
	Command *exec.Cmd
}

func NewBackgroundProcess(path string) (*backgroundProcess, error) {
	p := &backgroundProcess{
		Command: nil,
	}
	if p == nil {
		return nil, fmt.Errorf("Failed to create background process")
	}
	p.Command = exec.Command(path)
	if p.Command == nil {
		return nil, fmt.Errorf("failed to create background process")
	}

	// Forward output to regular stdout/stderr
	p.Command.Stdout = os.Stdout
	p.Command.Stderr = os.Stderr

	return p, nil
}

func (p *backgroundProcess) Start() error {
	return p.Command.Start()
}

func (p *backgroundProcess) Restart() error {
	if err := p.Stop(); err != nil {
		return err
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func (p *backgroundProcess) Stop() error {
	return p.Command.Process.Kill()
}

func (p *backgroundProcess) Running() bool {
	return p.Command.Process != nil
}
