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
	"syscall"
	"time"
)

type backgroundProcessImpl struct {
	path string
	command *exec.Cmd
}

// BackgroundProcess provides control over a process running in the
// background.
type BackgroundProcess interface {
	Start() error
	Stop() error
	Restart() error
	Running() bool
}

func NewBackgroundProcess(path string) (BackgroundProcess, error) {
	p := &backgroundProcessImpl{
		path: path,
		command: nil,
	}
	if p == nil {
		return nil, fmt.Errorf("Failed to create background process")
	}

	return p, nil
}

func (p *backgroundProcessImpl) Start() error {
	p.command = exec.Command(p.path)
	if p.command == nil {
		return fmt.Errorf("failed to create background process")
	}

	// Forward output to regular stdout/stderr
	p.command.Stdout = os.Stdout
	p.command.Stderr = os.Stderr

	// Create a new process group
	p.command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := p.command.Start()
	if err != nil {
		return err
	}

	return nil
}

func (p *backgroundProcessImpl) Restart() error {
	if err := p.Stop(); err != nil {
		return err
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func (p *backgroundProcessImpl) Stop() error {
	if p.command == nil {
		return nil
	}
	timer := time.AfterFunc(3 * time.Second, func() {
		p.command.Process.Kill()
	})
	p.command.Process.Signal(syscall.SIGTERM)
	p.command.Wait()
	timer.Stop()
	p.command = nil
	return nil
}

func (p *backgroundProcessImpl) Running() bool {
	return p.command != nil
}
