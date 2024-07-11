package storage

import "errors"

var (
	ErrUserNotFound = errors.New("client not found")
	ErrExists       = errors.New("client already exists")
)
