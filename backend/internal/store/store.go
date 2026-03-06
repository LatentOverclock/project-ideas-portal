package store

import (
	"context"
	"errors"
	"time"

	"project-ideas-portal/backend/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailExists = errors.New("email already exists")

type Store interface {
	Health(ctx context.Context) error
	Close()
	CreateUser(ctx context.Context, email, hash string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	CreateIdea(ctx context.Context, userID int64, title, description string) (*model.Idea, error)
	ListIdeas(ctx context.Context) ([]model.Idea, error)
}

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	p, err := pgxpool.New(ctx, dsn)
	if err != nil { return nil, err }
	s := &PostgresStore{pool: p}
	if err := s.migrate(ctx); err != nil { p.Close(); return nil, err }
	return s, nil
}

func (s *PostgresStore) migrate(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users(id BIGSERIAL PRIMARY KEY,email TEXT NOT NULL UNIQUE,password_hash TEXT NOT NULL,created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
		`CREATE TABLE IF NOT EXISTS ideas(id BIGSERIAL PRIMARY KEY,user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,title TEXT NOT NULL,description TEXT NOT NULL,created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`,
	}
	for _, q := range queries { if _, err := s.pool.Exec(ctx, q); err != nil { return err } }
	return nil
}

func (s *PostgresStore) Health(ctx context.Context) error { c, cancel := context.WithTimeout(ctx, 3*time.Second); defer cancel(); return s.pool.Ping(c) }
func (s *PostgresStore) Close()                           { s.pool.Close() }

func (s *PostgresStore) CreateUser(ctx context.Context, email, hash string) (*model.User, error) {
	var u model.User
	err := s.pool.QueryRow(ctx, `INSERT INTO users(email,password_hash) VALUES($1,$2) RETURNING id,email,password_hash,created_at`, email, hash).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" { return nil, ErrEmailExists }
		return nil, err
	}
	return &u, nil
}

func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	if err := s.pool.QueryRow(ctx, `SELECT id,email,password_hash,created_at FROM users WHERE email=$1`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt); err != nil { return nil, err }
	return &u, nil
}

func (s *PostgresStore) CreateIdea(ctx context.Context, userID int64, title, description string) (*model.Idea, error) {
	var i model.Idea
	if err := s.pool.QueryRow(ctx, `INSERT INTO ideas(user_id,title,description) VALUES($1,$2,$3) RETURNING id,user_id,title,description,created_at`, userID, title, description).Scan(&i.ID, &i.UserID, &i.Title, &i.Description, &i.CreatedAt); err != nil { return nil, err }
	if err := s.pool.QueryRow(ctx, `SELECT email FROM users WHERE id=$1`, userID).Scan(&i.UserEmail); err != nil { return nil, err }
	return &i, nil
}

func (s *PostgresStore) ListIdeas(ctx context.Context) ([]model.Idea, error) {
	rows, err := s.pool.Query(ctx, `SELECT i.id,i.user_id,u.email,i.title,i.description,i.created_at FROM ideas i JOIN users u ON u.id=i.user_id ORDER BY i.created_at DESC`)
	if err != nil { return nil, err }
	defer rows.Close()
	out := make([]model.Idea, 0)
	for rows.Next() {
		var i model.Idea
		if err := rows.Scan(&i.ID, &i.UserID, &i.UserEmail, &i.Title, &i.Description, &i.CreatedAt); err != nil { return nil, err }
		out = append(out, i)
	}
	return out, rows.Err()
}
