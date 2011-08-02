package main

import (
	"bytes"
	"exec"
	"os"
	"clingon"
)

type ShellEvaluator struct{}

func (eval *ShellEvaluator) Run(console *clingon.Console, command string) os.Error {
	var buf bytes.Buffer

	cmd := exec.Command(os.Getenv("SHELL"), "-c", command)
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()

	console.Print(buf.String())

	return err
}
