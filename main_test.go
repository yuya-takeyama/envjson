package main

import (
	"bytes"
	"io"
	"os/exec"
	"syscall"
	"testing"
)

func TestExec(t *testing.T) {
	stdin := bytes.NewBuffer([]byte(`{"FOO":"BAR","HOGE":"FUGA"}`))
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exitStatus, err := runCommand(stdin, stdout, stderr, "./envjson", "sh", "-c", `echo $FOO;echo $HOGE`)
	if err != nil {
		t.Fatalf("failed to run command: %s", err)
	}

	if bytes.Compare([]byte("BAR\nFUGA\n"), stdout.Bytes()) != 0 {
		t.Fatal("stdout not matched")
	}

	if bytes.Compare([]byte(""), stderr.Bytes()) != 0 {
		t.Fatal("stderr not matched")
	}

	if exitStatus != 0 {
		t.Fatal("exit status not matched")
	}
}

func TestExitWithNonZero(t *testing.T) {
	stdin := bytes.NewBuffer([]byte(`{"FOO":"BAR","HOGE":"FUGA"}`))
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exitStatus, err := runCommand(stdin, stdout, stderr, "./envjson", "sh", "-c", `echo $FOO;echo $HOGE; exit 1`)
	if err != nil {
		t.Fatalf("failed to run command: %s", err)
	}

	if bytes.Compare([]byte("BAR\nFUGA\n"), stdout.Bytes()) != 0 {
		t.Fatal("stdout not matched")
	}

	if bytes.Compare([]byte(""), stderr.Bytes()) != 0 {
		t.Fatal("stderr not matched")
	}

	if exitStatus != 1 {
		t.Fatal("exit status not matched")
	}
}

func TestWithBrokenJson(t *testing.T) {
	stdin := bytes.NewBuffer([]byte(`{"FOO":"BAR"`))
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exitStatus, err := runCommand(stdin, stdout, stderr, "./envjson", "sh", "-c", `echo $FOO`)
	if err != nil {
		t.Fatalf("failed to run command: %s", err)
	}

	if bytes.Compare([]byte(""), stdout.Bytes()) != 0 {
		t.Fatal("stdout not matched")
	}

	if bytes.Compare([]byte("envjson: fatal: unable to load JSON: unexpected EOF\n"), stderr.Bytes()) != 0 {
		t.Fatal("stderr not matched")
	}

	if exitStatus != 111 {
		t.Fatal("exit status not matched")
	}
}

func TestWithoutChild(t *testing.T) {
	stdin := bytes.NewBuffer([]byte(`{"FOO":"BAR","HOGE":"FUGA"}`))
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	exitStatus, err := runCommand(stdin, stdout, stderr, "./envjson")
	if err != nil {
		t.Fatalf("failed to run command: %s", err)
	}

	if bytes.Compare([]byte(""), stdout.Bytes()) != 0 {
		t.Fatal("stdout not matched")
	}

	if bytes.Compare([]byte("envjson: usage: echo JSON | envjson child\n"), stderr.Bytes()) != 0 {
		t.Fatal("stderr not matched")
	}

	if exitStatus != 111 {
		t.Fatal("exit status not matched")
	}
}

func runCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer, path string, args ...string) (int, error) {
	cmd := exec.Command(path, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	var exitStatus int
	if err := cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if s, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitStatus = s.ExitStatus()
			}
		} else {
			return 0, err
		}
	} else {
		exitStatus = 0
	}

	return exitStatus, nil
}
