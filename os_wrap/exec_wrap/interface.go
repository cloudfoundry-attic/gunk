package exec_wrap

import (
	"io"
)

//go:generate counterfeiter -o execfakes/fake_cmd.go . Cmd

type Cmd interface {
	Start() error
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Wait() error
}

//go:generate counterfeiter -o execfakes/fake_exec.go . Exec

type Exec interface {
	Command(name string, arg ...string) Cmd
}