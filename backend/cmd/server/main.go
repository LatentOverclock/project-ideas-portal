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
	databaseURL := getenv("DATABASE_URL", "postgres://app:app@localhost:5432/projectideas?sslmode=disable")
	port := getenv("PORT", "8080")
	jwtSecret := getenv("JWT_SECRET", "change-me")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := store.NewPostgresStore(ctx, databaseURL)
	if err != nil {
		log.Fatalf("store init failed: %v", err)
	}
	defer s.Close()

	a, err := app.New(s, jwtSecret)
	if err != nil {
		log.Fatalf("app init failed: %v", err)
	}

	log.Printf("backend listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, a.Router()))
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
