package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"project-ideas-portal/backend/internal/app"
	"project-ideas-portal/backend/internal/store"
)

func main() {
	dsn := env("DATABASE_URL", "postgres://app:app@localhost:5432/projectideas?sslmode=disable")
	port := env("PORT", "8080")
	secret := env("JWT_SECRET", "change-me")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s, err := store.NewPostgresStore(ctx, dsn)
	if err != nil { log.Fatalf("store init failed: %v", err) }
	defer s.Close()
	a, err := app.New(s, secret)
	if err != nil { log.Fatalf("app init failed: %v", err) }
	log.Printf("backend listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, a.Router()))
}

func env(k, d string) string { if v := os.Getenv(k); v != "" { return v }; return d }
