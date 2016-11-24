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

	"gopkg.in/tomb.v2"
)

type backgroundProcessImpl struct {
	path    string
	args    []string
	command *exec.Cmd
	tomb    *tomb.Tomb
}

// BackgroundProcess provides control over a process running in the
// background.
type BackgroundProcess interface {
	Start() error
	Stop() error
	Restart() error
	Running() bool
}

func NewBackgroundProcess(path string, args ...string) (BackgroundProcess, error) {
	p := &backgroundProcessImpl{
		path:    path,
		args:    args,
		command: nil,
	}
	if p == nil {
		return nil, fmt.Errorf("Failed to create background process")
	}

	return p, nil
}

func (p *backgroundProcessImpl) Start() error {
	if p.Running() {
		return fmt.Errorf("Background process is already running")
	}

	p.command = exec.Command(p.path, p.args...)
	if p.command == nil {
		return fmt.Errorf("Failed to create background process")
	}

	// Forward output to regular stdout/stderr
	p.command.Stdout = os.Stdout
	p.command.Stderr = os.Stderr

	// Create a new process group
	p.command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// We need to recreate the tomb here everytime as otherwise
	// it will not cleanup its state from the last time.
	p.tomb = &tomb.Tomb{}

	c := make(chan int)
	p.tomb.Go(func() error {
		err := p.command.Start()
		if err != nil {
			fmt.Printf("Failed to execute process for binary '%s'", p.path)
			return err
		}
		c <- 1
		p.command.Wait()
		p.command = nil
		return nil
	})

	// Wait until the process is really started
	_ = <-c

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
	if !p.Running() {
		return nil
	}
	timer := time.AfterFunc(10*time.Second, func() {
		p.command.Process.Kill()
	})
	p.command.Process.Signal(syscall.SIGTERM)
	p.tomb.Kill(nil)
	p.tomb.Wait()
	timer.Stop()
	return nil
}

func (p *backgroundProcessImpl) Running() bool {
	return p.command != nil
}
