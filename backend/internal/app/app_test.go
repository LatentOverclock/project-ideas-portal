package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"project-ideas-portal/backend/internal/store"
)

func gql(t *testing.T, h http.Handler, token string, query string, vars map[string]any) map[string]any {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"query": query, "variables": vars})
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	var out map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	return out
}

func TestRequirementsFlow(t *testing.T) {
	s := store.NewMemoryStore()
	a, err := New(s, "test-secret")
	if err != nil {
		t.Fatal(err)
	}
	h := a.Router()

	// R1: Users can register with email + password.
	reg := gql(t, h, "", `mutation($email:String!,$password:String!){register(email:$email,password:$password){email token}}`, map[string]any{"email": "u@example.com", "password": "password-123"})
	if reg["errors"] != nil {
		t.Fatalf("register failed: %v", reg["errors"])
	}
	regData := reg["data"].(map[string]any)["register"].(map[string]any)
	token := regData["token"].(string)
	if regData["email"].(string) != "u@example.com" || token == "" {
		t.Fatalf("unexpected register payload: %v", regData)
	}

	// R2: Users can login with existing credentials.
	login := gql(t, h, "", `mutation($email:String!,$password:String!){login(email:$email,password:$password){email token}}`, map[string]any{"email": "u@example.com", "password": "password-123"})
	if login["errors"] != nil {
		t.Fatalf("login failed: %v", login["errors"])
	}

	// R3: Authenticated users can create ideas.
	create1 := gql(t, h, token, `mutation($title:String!,$description:String!){createIdea(title:$title,description:$description){id title userEmail createdAt}}`, map[string]any{"title": "Idea 1", "description": "Desc 1"})
	if create1["errors"] != nil {
		t.Fatalf("createIdea #1 failed: %v", create1["errors"])
	}
	time.Sleep(10 * time.Millisecond)
	create2 := gql(t, h, token, `mutation($title:String!,$description:String!){createIdea(title:$title,description:$description){id title userEmail createdAt}}`, map[string]any{"title": "Idea 2", "description": "Desc 2"})
	if create2["errors"] != nil {
		t.Fatalf("createIdea #2 failed: %v", create2["errors"])
	}

	// Register second user for ownership checks.
	regOther := gql(t, h, "", `mutation($email:String!,$password:String!){register(email:$email,password:$password){email token}}`, map[string]any{"email": "other@example.com", "password": "password-123"})
	otherToken := regOther["data"].(map[string]any)["register"].(map[string]any)["token"].(string)

	// R4+R5: List ideas newest-first and include creator + timestamp.
	list := gql(t, h, "", `query { ideas { id title userEmail createdAt } }`, nil)
	if list["errors"] != nil {
		t.Fatalf("ideas query failed: %v", list["errors"])
	}
	ideas := list["data"].(map[string]any)["ideas"].([]any)
	if len(ideas) != 2 {
		t.Fatalf("expected 2 ideas, got %d", len(ideas))
	}
	first := ideas[0].(map[string]any)
	second := ideas[1].(map[string]any)
	if first["title"].(string) != "Idea 2" || second["title"].(string) != "Idea 1" {
		t.Fatalf("ideas not newest-first: %v", ideas)
	}
	if first["userEmail"].(string) == "" || first["createdAt"].(string) == "" {
		t.Fatalf("creator/timestamp missing: %v", first)
	}

	// R6: Authenticated users can delete only their own ideas.
	firstIdeaID := first["id"].(string)
	deleteByOther := gql(t, h, otherToken, `mutation($id:ID!){deleteIdea(id:$id)}`, map[string]any{"id": firstIdeaID})
	if deleteByOther["errors"] != nil {
		t.Fatalf("delete by other returned error: %v", deleteByOther["errors"])
	}
	if deleteByOther["data"].(map[string]any)["deleteIdea"].(bool) {
		t.Fatalf("other user should not be able to delete idea")
	}

	deleteOwn := gql(t, h, token, `mutation($id:ID!){deleteIdea(id:$id)}`, map[string]any{"id": firstIdeaID})
	if deleteOwn["errors"] != nil {
		t.Fatalf("delete own failed: %v", deleteOwn["errors"])
	}
	if !deleteOwn["data"].(map[string]any)["deleteIdea"].(bool) {
		t.Fatalf("owner should be able to delete idea")
	}

	listAfterDelete := gql(t, h, "", `query { ideas { title } }`, nil)
	if len(listAfterDelete["data"].(map[string]any)["ideas"].([]any)) != 1 {
		t.Fatalf("expected 1 idea after delete")
	}
}
