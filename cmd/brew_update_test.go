package cmd

import (
	"errors"
	"testing"
	"time"
)

func TestRunBrewUpdateRetriesTransientFailureOnce(t *testing.T) {
	attempts := 0
	sleeps := 0
	err := runBrewUpdateWith(func() (string, error) {
		attempts++
		if attempts == 1 {
			return "curl: (35) LibreSSL SSL_connect: SSL_ERROR_SYSCALL\nHTTP status: 000", errors.New("exit status 1")
		}
		return "", nil
	}, func(delay time.Duration) {
		sleeps++
		if delay != brewUpdateRetryDelay {
			t.Fatalf("delay = %s", delay)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if attempts != 2 || sleeps != 1 {
		t.Fatalf("attempts = %d, sleeps = %d", attempts, sleeps)
	}
}

func TestRunBrewUpdateDoesNotRetryPermanentFailure(t *testing.T) {
	attempts := 0
	wantErr := errors.New("exit status 1")
	err := runBrewUpdateWith(func() (string, error) {
		attempts++
		return "Refusing to load formula from untrusted tap", wantErr
	}, func(time.Duration) {
		t.Fatal("unexpected sleep")
	})
	if !errors.Is(err, wantErr) || attempts != 1 {
		t.Fatalf("err = %v, attempts = %d", err, attempts)
	}
}

func TestRunBrewUpdateStopsAfterSecondFailure(t *testing.T) {
	attempts := 0
	wantErr := errors.New("second failure")
	err := runBrewUpdateWith(func() (string, error) {
		attempts++
		if attempts == 1 {
			return "curl: (28) Operation timed out", errors.New("first failure")
		}
		return "curl: (28) Operation timed out", wantErr
	}, func(time.Duration) {})
	if !errors.Is(err, wantErr) || attempts != 2 {
		t.Fatalf("err = %v, attempts = %d", err, attempts)
	}
}

func TestTransientBrewUpdateErrorClassification(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{name: "TLS syscall", output: "curl: (35) LibreSSL SSL_connect: SSL_ERROR_SYSCALL", want: true},
		{name: "DNS", output: "curl: (6) Could not resolve host", want: true},
		{name: "HTTP 429", output: "HTTP status: 429", want: true},
		{name: "HTTP 503", output: "HTTP status: 503", want: true},
		{name: "certificate", output: "curl: (60) SSL certificate problem", want: false},
		{name: "HTTP 404", output: "HTTP status: 404", want: false},
		{name: "tap trust", output: "Refusing to load formula from untrusted tap", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTransientBrewUpdateError(tt.output); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
