package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"project-ideas-portal/backend/internal/auth"
	"project-ideas-portal/backend/internal/store"

	"github.com/graphql-go/graphql"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const userContextKey contextKey = "authUser"

type App struct {
	store     store.Store
	schema    graphql.Schema
	jwtSecret string
}

type authUser struct { ID int64; Email string }

func New(s store.Store, jwtSecret string) (*App, error) {
	schema, err := buildSchema(s, jwtSecret)
	if err != nil { return nil, err }
	return &App{store: s, schema: schema, jwtSecret: jwtSecret}, nil
}

func (a *App) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/graphql", a.handleGraphQL)
	return withCORS(a.withAuthContext(mux))
}

func withCORS(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	if r.Method == http.MethodOptions { w.WriteHeader(http.StatusNoContent); return }
	next.ServeHTTP(w, r)
})}

func (a *App) withAuthContext(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	head := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(head), "bearer ") {
		token := strings.TrimSpace(head[7:])
		if c, err := auth.Parse(a.jwtSecret, token); err == nil {
			ctx = context.WithValue(ctx, userContextKey, authUser{ID: c.UserID, Email: c.Email})
		}
	}
	next.ServeHTTP(w, r.WithContext(ctx))
})}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	if err := a.store.Health(r.Context()); err != nil { http.Error(w, err.Error(), http.StatusServiceUnavailable); return }
	writeJSON(w, map[string]any{"ok": true})
}

type gqlReq struct { Query string `json:"query"`; Variables map[string]any `json:"variables"` }
func (a *App) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
	var req gqlReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
	res := graphql.Do(graphql.Params{Schema:a.schema,RequestString:req.Query,VariableValues:req.Variables,Context:r.Context()})
	writeJSON(w,res)
}
func writeJSON(w http.ResponseWriter, payload any) { w.Header().Set("Content-Type", "application/json"); _ = json.NewEncoder(w).Encode(payload) }

func buildSchema(s store.Store, secret string) (graphql.Schema, error) {
	ideaType := graphql.NewObject(graphql.ObjectConfig{Name:"ProjectIdea",Fields:graphql.Fields{
		"id":&graphql.Field{Type:graphql.ID},"title":&graphql.Field{Type:graphql.String},"description":&graphql.Field{Type:graphql.String},"userEmail":&graphql.Field{Type:graphql.String},"createdAt":&graphql.Field{Type:graphql.String},
	}})
	authType := graphql.NewObject(graphql.ObjectConfig{Name:"AuthPayload",Fields:graphql.Fields{"token":&graphql.Field{Type:graphql.String},"email":&graphql.Field{Type:graphql.String}}})

	query := graphql.NewObject(graphql.ObjectConfig{Name:"Query",Fields:graphql.Fields{
		"ideas": &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(ideaType))), Resolve: func(p graphql.ResolveParams) (any, error) {
			ideas, err := s.ListIdeas(ctx(p.Context)); if err != nil { return nil, err }
			out := make([]map[string]any,0,len(ideas)); for _, i := range ideas { out = append(out, map[string]any{"id":fmt.Sprintf("%d",i.ID),"title":i.Title,"description":i.Description,"userEmail":i.UserEmail,"createdAt":i.CreatedAt.UTC().Format(time.RFC3339)}) }
			return out, nil
		}},
	}})

	mutation := graphql.NewObject(graphql.ObjectConfig{Name:"Mutation",Fields:graphql.Fields{
		"register": &graphql.Field{Type: graphql.NewNonNull(authType), Args: graphql.FieldConfigArgument{"email":&graphql.ArgumentConfig{Type:graphql.NewNonNull(graphql.String)},"password":&graphql.ArgumentConfig{Type:graphql.NewNonNull(graphql.String)}}, Resolve: func(p graphql.ResolveParams) (any, error) {
			email := strings.ToLower(strings.TrimSpace(p.Args["email"].(string))); pass := p.Args["password"].(string)
			if len(pass) < 8 { return nil, errors.New("password too short") }
			hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
			u, err := s.CreateUser(ctx(p.Context), email, string(hash)); if err != nil { return nil, err }
			t, err := auth.Sign(secret, u.ID, u.Email); if err != nil { return nil, err }
			return map[string]any{"token":t,"email":u.Email}, nil
		}},
		"login": &graphql.Field{Type: graphql.NewNonNull(authType), Args: graphql.FieldConfigArgument{"email":&graphql.ArgumentConfig{Type:graphql.NewNonNull(graphql.String)},"password":&graphql.ArgumentConfig{Type:graphql.NewNonNull(graphql.String)}}, Resolve: func(p graphql.ResolveParams) (any, error) {
			email := strings.ToLower(strings.TrimSpace(p.Args["email"].(string))); pass := p.Args["password"].(string)
			u, err := s.GetUserByEmail(ctx(p.Context), email); if err != nil { return nil, errors.New("invalid credentials") }
			if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pass)) != nil { return nil, errors.New("invalid credentials") }
			t, err := auth.Sign(secret, u.ID, u.Email); if err != nil { return nil, err }
			return map[string]any{"token":t,"email":u.Email}, nil
		}},
		"createIdea": &graphql.Field{Type: graphql.NewNonNull(ideaType), Args: graphql.FieldConfigArgument{"title":&graphql.ArgumentConfig{Type:graphql.NewNonNull(graphql.String)},"description":&graphql.ArgumentConfig{Type:graphql.NewNonNull(graphql.String)}}, Resolve: func(p graphql.ResolveParams) (any, error) {
			u, ok := p.Context.Value(userContextKey).(authUser); if !ok { return nil, errors.New("unauthorized") }
			title := strings.TrimSpace(p.Args["title"].(string)); desc := strings.TrimSpace(p.Args["description"].(string))
			if title == "" || desc == "" { return nil, errors.New("title and description required") }
			i, err := s.CreateIdea(ctx(p.Context), u.ID, title, desc); if err != nil { return nil, err }
			return map[string]any{"id":fmt.Sprintf("%d",i.ID),"title":i.Title,"description":i.Description,"userEmail":i.UserEmail,"createdAt":i.CreatedAt.UTC().Format(time.RFC3339)}, nil
		}},
	}})
	return graphql.NewSchema(graphql.SchemaConfig{Query: query, Mutation: mutation})
}
func ctx(c context.Context) context.Context { if c==nil { return context.Background() }; return c }
