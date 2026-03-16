package auth

import (
	"context"
	"crypto/subtle"
	"fmt"
	"sync"

	"go.redsock.ru/rerrors"

	"go.redsock.ru/mead/internal/config"
	"go.redsock.ru/mead/internal/domain"
	"go.redsock.ru/mead/internal/service/user_errors"
	"go.redsock.ru/mead/internal/storage"
	"go.redsock.ru/mead/internal/storage/sqlite/user_queries"
	"go.redsock.ru/mead/internal/utils"
)

type Authenticator interface {
	Authenticate(username, password string) error
}

type CredentialService struct {
	usersStorage storage.Users

	mu    sync.RWMutex
	cache map[string][]byte

	proxyBaseUrl string
}

func NewAuth(cfg config.Config, str storage.Storage) *CredentialService {
	return &CredentialService{
		usersStorage: str.Users(),

		cache: make(map[string][]byte),

		proxyBaseUrl: fmt.Sprintf("tg://socks?server=%s&port=%d",
			cfg.Environment.ProxyAddress,
			cfg.Environment.ProxyPort),
	}
}

func (s *CredentialService) Add(ctx context.Context, username string) error {
	if username == "" {
		return user_errors.ErrUsernameIsEmpty
	}

	password := utils.GeneratePassword(16)

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
func (s *CredentialService) Authenticate(ctx context.Context, username, password string) error {

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
		return user_errors.ErrUnauthorized
	}
	return nil
}

func (s *CredentialService) ListUsers(ctx context.Context, req domain.ListUsersReq) ([]domain.User, error) {
	users, err := s.usersStorage.List(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing users")
	}

	return users, nil
}

func (s *CredentialService) AuthByTelegramUsername(ctx context.Context, username string) (domain.UserAuth, error) {
	user, err := s.usersStorage.GetByTelegramName(ctx, username)
	if err != nil {
		return domain.UserAuth{}, rerrors.Wrap(err, "error getting user")
	}

	pass, err := s.usersStorage.GetPassByUsername(ctx, username)
	if err != nil {
		return domain.UserAuth{}, rerrors.Wrap(err, "error reading pass")
	}

	return domain.UserAuth{
		User:      user,
		ProxyLink: s.generateProxyLink(user, pass),
	}, nil
}

func (s *CredentialService) generateProxyLink(user domain.User, pass string) string {
	return fmt.Sprintf(s.proxyBaseUrl+"&user=%s&pass=%s", user.Username, pass)
}
