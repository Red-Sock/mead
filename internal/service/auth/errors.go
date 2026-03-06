package auth

import (
	"go.redsock.ru/rerrors"
)

var (
	ErrUsernameIsEmpy = rerrors.New("username must not be empty")
	ErrPasswordIsEmpy = rerrors.New("password must not be empty")
	ErrUnauthorized   = rerrors.New("unauthorized: invalid credentials")
)
