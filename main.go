package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const (
	ExitFatal = 111
)

func main() {
	env, err := makeEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ExitFatal)
	}

	if len(os.Args) < 2 {
		usage()
	}

	path := os.Args[1]
	args := os.Args[2:]
	cmd := commandWithEnv(path, args, env)

	var exitStatus int
	if err := cmd.Run(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if s, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				// Exit with non-zero
				exitStatus = s.ExitStatus()
			}
		} else {
			// Exit unsuccessfully
			fmt.Fprintf(os.Stderr, "envjson: fatal: failed to run command: %s\n", err)
			os.Exit(ExitFatal)
		}
	} else {
		// Exit with zero
		exitStatus = 0
	}

	os.Exit(exitStatus)
}

func makeEnv() ([]string, error) {
	decoder := json.NewDecoder(os.Stdin)

	var json = make(map[string]string)

	if err := decoder.Decode(&json); err != nil {
		return nil, errors.New(fmt.Sprintf("envjson: fatal: unable to load JSON: %s", err))
	}

	origEnv := os.Environ()
	env := make([]string, len(origEnv)+len(json))

	copy(env, origEnv)

	for key, value := range json {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env, nil
}

func commandWithEnv(path string, args []string, env []string) *exec.Cmd {
	cmd := exec.Command(path, args...)

	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

func usage() {
	fmt.Fprintln(os.Stderr, "envjson: usage: echo JSON | envjson child")
	os.Exit(ExitFatal)
}
