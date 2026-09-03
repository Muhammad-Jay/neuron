package process

import "os"

type Command struct {
	Path string
	Args []string
	Dir string
	Env []string
}

type Process struct {
	Cmd Command
}

func NewProcess(cmd Command) *Process {
	if cmd.Dir == "" {
		cmd.Dir = getWorkingDirectory()
	}

	return &Process{ Cmd: cmd }
}

func getWorkingDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}