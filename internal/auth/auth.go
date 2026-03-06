// Package auth implements SOCKS5 username/password authentication
// as specified in RFC 1929.
package auth

import (
	"crypto/subtle"
	"fmt"
	"sync"
)

// ErrUnauthorized is returned when credentials are invalid.
var ErrUnauthorized = fmt.Errorf("unauthorized: invalid credentials")

// Authenticator validates SOCKS5 username/password credentials.
type Authenticator interface {
	Authenticate(username, password string) error
}

// CredentialStore is a thread-safe, in-memory username/password store.
// Comparisons are performed in constant time to prevent timing attacks.
type CredentialStore struct {
	mu    sync.RWMutex
	creds map[string][]byte // username -> hashed/raw password bytes
}

// NewCredentialStore creates an empty store.
func NewCredentialStore() *CredentialStore {
	return &CredentialStore{creds: make(map[string][]byte)}
}

// Add registers a credential pair. Calling Add with the same username
// overwrites the previous password.
func (s *CredentialStore) Add(username, password string) error {
	if username == "" {
		return fmt.Errorf("username must not be empty")
	}
	if password == "" {
		return fmt.Errorf("password must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.creds[username] = []byte(password)
	return nil
}

// Remove deletes a credential pair. It is a no-op if the user is unknown.
func (s *CredentialStore) Remove(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.creds, username)
}

// Authenticate implements Authenticator using constant-time comparison.
func (s *CredentialStore) Authenticate(username, password string) error {
	s.mu.RLock()
	stored, ok := s.creds[username]
	s.mu.RUnlock()

	// Always do the comparison even on miss to prevent timing side-channels
	// on username enumeration.
	candidate := []byte(password)
	dummy := []byte("$dummy$")
	reference := stored
	if !ok {
		reference = dummy
	}

	// subtle.ConstantTimeCompare returns 1 only when lengths AND content match.
	match := subtle.ConstantTimeCompare(reference, candidate) == 1
	if !ok || !match {
		return ErrUnauthorized
	}
	return nil
}

// Len returns the number of registered credentials.
func (s *CredentialStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.creds)
}
