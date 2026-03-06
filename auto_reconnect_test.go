package main

import (
	"testing"
	"time"
)

func TestCalculateAutoReconnectDelay(t *testing.T) {
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{attempt: 0, expected: 5 * time.Second},
		{attempt: 1, expected: 5 * time.Second},
		{attempt: 2, expected: 10 * time.Second},
		{attempt: 3, expected: 20 * time.Second},
		{attempt: 7, expected: 320 * time.Second},
		{attempt: 8, expected: 5 * time.Minute},
		{attempt: 20, expected: 5 * time.Minute},
	}

	for _, tt := range tests {
		if got := calculateAutoReconnectDelay(tt.attempt); got != tt.expected {
			t.Fatalf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, got)
		}
	}
}

func TestMyClientBeginAutoReconnectGuard(t *testing.T) {
	mycli := &MyClient{}

	if !mycli.beginAutoReconnect() {
		t.Fatalf("expected first reconnect scheduling to succeed")
	}
	if mycli.beginAutoReconnect() {
		t.Fatalf("expected duplicate reconnect scheduling to be blocked")
	}

	mycli.disableAutoReconnect()
	if mycli.beginAutoReconnect() {
		t.Fatalf("expected reconnect scheduling to stop once disabled")
	}

	mycli.resetAutoReconnect()
	if !mycli.beginAutoReconnect() {
		t.Fatalf("expected reconnect scheduling after reset")
	}
}
