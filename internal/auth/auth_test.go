package auth

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/amvid/vanillastone/internal/store"
)

func newAuth(t *testing.T) *Auth {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return New(st)
}

// TestRegisterUniqueness encodes the core account rule: a username can be taken
// exactly once. The second registration must fail so two people can't share an
// identity.
func TestRegisterUniqueness(t *testing.T) {
	a := newAuth(t)
	if err := a.Register("alice", "password123"); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if err := a.Register("alice", "different456"); !errors.Is(err, ErrTaken) {
		t.Fatalf("duplicate username must be rejected, got %v", err)
	}
}

// TestRegisterValidation rejects malformed input before hashing/storage.
func TestRegisterValidation(t *testing.T) {
	a := newAuth(t)
	cases := []struct{ user, pass string }{
		{"ab", "password123"}, // username too short
		{"alice", "short"},    // password too short
		{"", "password123"},   // empty username
	}
	for _, c := range cases {
		if err := a.Register(c.user, c.pass); !errors.Is(err, ErrValidation) {
			t.Fatalf("Register(%q,%q) should be ErrValidation, got %v", c.user, c.pass, err)
		}
	}
}

// TestLogin verifies a good login yields a working token and that wrong
// passwords / unknown users are rejected with the same opaque error (no user
// enumeration).
func TestLogin(t *testing.T) {
	a := newAuth(t)
	if err := a.Register("alice", "password123"); err != nil {
		t.Fatalf("register: %v", err)
	}

	token, err := a.Login("alice", "password123")
	if err != nil {
		t.Fatalf("valid login failed: %v", err)
	}
	if name, ok := a.Username(token); !ok || name != "alice" {
		t.Fatalf("token should resolve to alice, got %q ok=%v", name, ok)
	}

	if _, err := a.Login("alice", "wrongpass"); !errors.Is(err, ErrBadCreds) {
		t.Fatalf("wrong password must be ErrBadCreds, got %v", err)
	}
	if _, err := a.Login("nobody", "password123"); !errors.Is(err, ErrBadCreds) {
		t.Fatalf("unknown user must be ErrBadCreds, got %v", err)
	}
}

// TestUnknownToken rejects a token that was never issued.
func TestUnknownToken(t *testing.T) {
	a := newAuth(t)
	if _, ok := a.Username("deadbeef"); ok {
		t.Fatal("unknown token must not resolve")
	}
}
