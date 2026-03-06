package store

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"project-ideas-portal/backend/internal/model"
)

type MemoryStore struct {
	mu         sync.Mutex
	nextUserID int64
	nextIdeaID int64
	users      map[string]model.User
	ideas      []model.Idea
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{nextUserID: 1, nextIdeaID: 1, users: map[string]model.User{}, ideas: []model.Idea{}}
}

func (s *MemoryStore) Health(context.Context) error { return nil }
func (s *MemoryStore) Close()                       {}

func (s *MemoryStore) CreateUser(_ context.Context, email, passwordHash string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[email]; exists {
		return nil, ErrEmailExists
	}
	u := model.User{ID: s.nextUserID, Email: email, PasswordHash: passwordHash, CreatedAt: time.Now().UTC()}
	s.nextUserID++
	s.users[email] = u
	return &u, nil
}

func (s *MemoryStore) GetUserByEmail(_ context.Context, email string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[email]
	if !ok {
		return nil, errors.New("not found")
	}
	return &u, nil
}

func (s *MemoryStore) CreateIdea(_ context.Context, userID int64, title, description string) (*model.Idea, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var email string
	for _, u := range s.users {
		if u.ID == userID {
			email = u.Email
			break
		}
	}
	if email == "" {
		return nil, errors.New("user not found")
	}
	i := model.Idea{ID: s.nextIdeaID, UserID: userID, UserEmail: email, Title: title, Description: description, CreatedAt: time.Now().UTC()}
	s.nextIdeaID++
	s.ideas = append(s.ideas, i)
	return &i, nil
}

func (s *MemoryStore) ListIdeas(_ context.Context) ([]model.Idea, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := append([]model.Idea{}, s.ideas...)
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
