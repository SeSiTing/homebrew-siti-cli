package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const brewUpdateRetryDelay = time.Second

var transientHTTPStatusPattern = regexp.MustCompile(`(?i)HTTP status:\s*(429|5[0-9]{2})\b`)

func runBrewUpdate() error {
	if err := preflightUpgradeProxy(); err != nil {
		return err
	}
	hint, err := brewProxyHint()
	if err != nil {
		return fmt.Errorf("读取代理配置: %w", err)
	}
	if hint != "" {
		fmt.Fprintln(os.Stderr, hint)
	}
	return runBrewUpdateWith(runBrewUpdateAttempt, time.Sleep)
}

func runBrewUpdateWith(attempt func() (string, error), sleep func(time.Duration)) error {
	stderr, err := attempt()
	if err == nil || !isTransientBrewUpdateError(stderr) {
		return err
	}

	fmt.Fprintln(os.Stderr, "! brew update 遇到临时网络错误，1 秒后重试（1/1）")
	sleep(brewUpdateRetryDelay)
	_, err = attempt()
	return err
}

func runBrewUpdateAttempt() (string, error) {
	fmt.Fprintln(os.Stderr, "  $ brew update")
	var stderr bytes.Buffer
	cmd := exec.Command("brew", "update")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	err := cmd.Run()
	return stderr.String(), err
}

func isTransientBrewUpdateError(output string) bool {
	lower := strings.ToLower(output)
	for _, marker := range []string{
		"curl: (6)",
		"curl: (7)",
		"curl: (18)",
		"curl: (28)",
		"curl: (35)",
		"curl: (52)",
		"curl: (55)",
		"curl: (56)",
		"ssl_error_syscall",
		"could not resolve host",
		"failed to connect",
		"connection reset",
		"operation timed out",
		"http status: 000",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return transientHTTPStatusPattern.MatchString(output)
}
