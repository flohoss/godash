package auth

import (
	"testing"
	"time"
)

func TestSignAndVerifySessionRoundTrip(t *testing.T) {
	value := signSession("secret-key", "user-subject", time.Now().Add(time.Hour))

	subject, expiresAt, err := verifySession("secret-key", value, time.Now())
	if err != nil {
		t.Fatalf("verifySession() error = %v", err)
	}
	if subject != "user-subject" {
		t.Errorf("subject = %q, want %q", subject, "user-subject")
	}
	if time.Until(expiresAt) <= 0 {
		t.Error("expected a future expiry")
	}
}

func TestVerifySessionRejectsExpired(t *testing.T) {
	value := signSession("secret-key", "user-subject", time.Now().Add(-time.Minute))

	if _, _, err := verifySession("secret-key", value, time.Now()); err == nil {
		t.Fatal("expected an expired session to be rejected")
	}
}

func TestVerifySessionRejectsWrongKey(t *testing.T) {
	value := signSession("secret-key", "user-subject", time.Now().Add(time.Hour))

	if _, _, err := verifySession("other-key", value, time.Now()); err == nil {
		t.Fatal("expected a session signed with another key to be rejected")
	}
}

func TestVerifySessionRejectsTamperedPayload(t *testing.T) {
	value := signSession("secret-key", "user-subject", time.Now().Add(time.Hour))

	tampered := signSession("secret-key", "admin-subject", time.Now().Add(time.Hour))
	if _, _, err := verifySession("secret-key", tampered+value[len(value)-1:], time.Now()); err == nil {
		t.Fatal("expected a tampered payload to be rejected")
	}
}

func TestVerifySessionRejectsMalformedValue(t *testing.T) {
	for _, value := range []string{"", "no-separator", "a.b", "!"} {
		if _, _, err := verifySession("secret-key", value, time.Now()); err == nil {
			t.Errorf("expected %q to be rejected", value)
		}
	}
}
