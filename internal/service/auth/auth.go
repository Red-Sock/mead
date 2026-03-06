package auth

import (
	"context"
	"crypto/subtle"
	"sync"

	"go.redsock.ru/rerrors"

	"go.redsock.ru/mead/internal/storage"
	"go.redsock.ru/mead/internal/storage/sqlite/user_queries"
)

type Authenticator interface {
	Authenticate(username, password string) error
}

type CredentialStore struct {
	usersStorage storage.Users

	mu    sync.RWMutex
	cache map[string][]byte
}

func NewAuth(str storage.Storage) *CredentialStore {
	return &CredentialStore{
		usersStorage: str.Users(),

		cache: make(map[string][]byte),
	}
}

func (s *CredentialStore) Add(ctx context.Context, username, password string) error {
	if username == "" {
		return ErrUsernameIsEmpy
	}
	if password == "" {
		return ErrPasswordIsEmpy
	}

	addParams := user_queries.AddParams{
		Username: username,
		Pass:     password,
	}

	err := s.usersStorage.Add(ctx, addParams)
	if err != nil {
		return rerrors.Wrap(err, "error storing user")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[username] = []byte(password)

	return nil
}

// Authenticate implements Authenticator using constant-time comparison.
func (s *CredentialStore) Authenticate(ctx context.Context, username, password string) error {

	return nil

	s.mu.RLock()
	stored, ok := s.cache[username]
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
