package jj

import "os/exec"

type Executor interface {
	Execute(args []string) ([]byte, error)
	ExecuteInDir(args []string, dir string) ([]byte, error)
}

type DefaultExecutor struct{}

func (e *DefaultExecutor) Execute(args []string) ([]byte, error) {
	return exec.Command("jj", args...).Output()
}

func (e *DefaultExecutor) ExecuteInDir(args []string, dir string) ([]byte, error) {
	cmd := exec.Command("jj", args...)
	cmd.Dir = dir
	return cmd.Output()
}
