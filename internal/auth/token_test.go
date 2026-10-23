package auth

import (
	"errors"
	"testing"
	"time"
)

func TestJWTManagerRoundTrip(t *testing.T) {
	manager := NewJWTManager("test-secret", time.Hour)
	manager.now = func() time.Time { return time.Unix(1_700_000_000, 0) }
	token, err := manager.Issue(42)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	userID, err := manager.Verify(token)
	if err != nil || userID != 42 {
		t.Fatalf("Verify() = %d, %v", userID, err)
	}
}

func TestJWTManagerRejectsExpiredToken(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	manager := NewJWTManager("test-secret", time.Minute)
	manager.now = func() time.Time { return now }
	token, err := manager.Issue(42)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	manager.now = func() time.Time { return now.Add(2 * time.Minute) }
	_, err = manager.Verify(token)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Verify() error = %v, want ErrInvalidToken", err)
	}
}
