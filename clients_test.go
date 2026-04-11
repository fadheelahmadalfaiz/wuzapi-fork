package main

import "testing"

func TestClientManagerBeginSessionStartGuard(t *testing.T) {
	cm := NewClientManager()

	if !cm.BeginSessionStart("user-1") {
		t.Fatalf("expected first session start to succeed")
	}

	if cm.BeginSessionStart("user-1") {
		t.Fatalf("expected duplicate session start to be blocked")
	}

	cm.FinishSessionStart("user-1")

	if !cm.BeginSessionStart("user-1") {
		t.Fatalf("expected session start after finish to succeed")
	}
}

func TestClientManagerHasSession(t *testing.T) {
	cm := NewClientManager()

	if cm.HasSession("user-1") {
		t.Fatalf("expected no session for new user")
	}

	cm.BeginSessionStart("user-1")
	if !cm.HasSession("user-1") {
		t.Fatalf("expected pending session start to count as active session")
	}

	cm.FinishSessionStart("user-1")
	cm.SetMyClient("user-1", &MyClient{})
	if !cm.HasSession("user-1") {
		t.Fatalf("expected in-memory client to count as active session")
	}

	cm.DeleteMyClient("user-1")
	if cm.HasSession("user-1") {
		t.Fatalf("expected session state to clear after deleting client")
	}
}
