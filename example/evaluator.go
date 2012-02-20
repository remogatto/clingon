package main

import (
	"bytes"
	"github.com/remogatto/clingon"
	"os"
	"os/exec"
)

type ShellEvaluator struct{}

func (eval *ShellEvaluator) Run(console *clingon.Console, command string) error {
	var buf bytes.Buffer

	cmd := exec.Command(os.Getenv("SHELL"), "-c", command)
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()

	console.Print(buf.String())

	return err
}
