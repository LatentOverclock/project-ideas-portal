package store

import (
	"context"
	"errors"

	"project-ideas-portal/backend/internal/model"
)

var ErrEmailExists = errors.New("email already exists")

type Store interface {
	Health(ctx context.Context) error
	Close()
	CreateUser(ctx context.Context, email, passwordHash string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	CreateIdea(ctx context.Context, userID int64, title, description string) (*model.Idea, error)
	DeleteIdea(ctx context.Context, userID int64, ideaID int64) (bool, error)
	ListIdeas(ctx context.Context) ([]model.Idea, error)
}
