package auth

import (
	"crypto/subtle"
	"fmt"
	"sync"

	"go.redsock.ru/rerrors"
)

var ErrUnauthorized = rerrors.New("unauthorized: invalid credentials")

type Authenticator interface {
	Authenticate(username, password string) error
}

type CredentialStore struct {
	mu    sync.RWMutex
	creds map[string][]byte
}

func NewCredentialStore() *CredentialStore {
	return &CredentialStore{creds: make(map[string][]byte)}
}

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
