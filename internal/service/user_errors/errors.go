package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	ErrUsernameIsEmpty = rerrors.NewUserError("username must not be empty", codes.InvalidArgument)
	ErrPasswordIsEmpty = rerrors.NewUserError("password must not be empty", codes.InvalidArgument)
	ErrUnauthorized    = rerrors.NewUserError("unauthorized: invalid credentials", codes.Unauthenticated)

	ErrAlreadyExists = rerrors.NewUserError("already exists", codes.AlreadyExists)
)
