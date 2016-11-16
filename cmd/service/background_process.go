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
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type backgroundProcess struct {
	Path string
	Command *exec.Cmd
}

func NewBackgroundProcess(path string) (*backgroundProcess, error) {
	p := &backgroundProcess{
		Path: path,
		Command: nil,
	}
	if p == nil {
		return nil, fmt.Errorf("Failed to create background process")
	}

	return p, nil
}

func (p *backgroundProcess) Start() error {
	log.Println("Starting background process")

	p.Command = exec.Command(p.Path)
	if p.Command == nil {
		return fmt.Errorf("failed to create background process")
	}

	// Forward output to regular stdout/stderr
	p.Command.Stdout = os.Stdout
	p.Command.Stderr = os.Stderr

	// Create a new process group
	p.Command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return p.Command.Start()
}

func (p *backgroundProcess) Restart() error {
	log.Println("Restarting background process")
	if err := p.Stop(); err != nil {
		return err
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func (p *backgroundProcess) Stop() error {
	log.Println("Stopping background process")
	timer := time.AfterFunc(3 * time.Second, func() {
		p.Command.Process.Kill()
	})
	p.Command.Process.Signal(syscall.SIGTERM)
	p.Command.Wait()
	timer.Stop()
	log.Println("process stopped: ", p.Command.ProcessState)
	p.Command = nil
	return nil
}

func (p *backgroundProcess) Running() bool {
	return p.Command.Process != nil && !p.Command.ProcessState.Exited()
}
