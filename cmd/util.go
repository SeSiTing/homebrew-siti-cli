package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// lookupEnv returns the value of the named environment variable, or "" if unset.
func lookupEnv(key string) string {
	v, _ := os.LookupEnv(key)
	return v
}

// firstNonEmpty returns the first non-empty string from the provided values.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// runCmd runs a command inheriting stdout/stderr so the user sees live output.
func runCmd(name string, args ...string) error {
	return runCmdIn("", name, args...)
}

// runCmdTee is like runCmd but also tees stdout into buf for post-run parsing.
func runCmdTee(buf *bytes.Buffer, name string, args ...string) error {
	return runCmdInTee("", buf, name, args...)
}

// runCmdIn runs a command in the specified directory (empty = inherit cwd).
func runCmdIn(dir, name string, args ...string) error {
	return runCmdInTee(dir, nil, name, args...)
}

func runCmdInTee(dir string, buf *bytes.Buffer, name string, args ...string) error {
	fmt.Fprintf(os.Stderr, "  $ %s %s\n", name, strings.Join(args, " "))
	c := exec.Command(name, args...)
	c.Stdin = os.Stdin
	c.Dir = dir
	if buf != nil {
		c.Stdout = io.MultiWriter(os.Stdout, buf)
	} else {
		c.Stdout = os.Stdout
	}
	c.Stderr = os.Stderr
	return c.Run()
}

// runCmdOutput runs a command and captures stdout into buf without printing.
func runCmdOutput(buf *bytes.Buffer, name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdin = os.Stdin
	c.Stdout = buf
	c.Stderr = nil
	return c.Run()
}
