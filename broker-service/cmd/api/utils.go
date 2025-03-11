package main

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidSession = errors.New("invalid session")
	ErrSessionExpired = errors.New("session expired")
	ErrMissingCSRF    = errors.New("missing csrf token")
	ErrInvalidCSRF    = errors.New("invalid csrf token")
)

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	return userID, nil
}
