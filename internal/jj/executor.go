package jj

import "os/exec"

type Executor interface {
	Execute(args []string) ([]byte, error)
}

type DefaultExecutor struct{}

func (e *DefaultExecutor) Execute(args []string) ([]byte, error) {
	return exec.Command("jj", args...).Output()
}
