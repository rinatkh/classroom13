package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/rinatkh/classroom13/internal/auth"
	"github.com/rinatkh/classroom13/internal/httpapi"
	"github.com/rinatkh/classroom13/internal/httpapi/handler"
	"github.com/rinatkh/classroom13/internal/httpapi/middleware"
	"github.com/rinatkh/classroom13/internal/note"
	"github.com/rinatkh/classroom13/internal/repository/memory"
	"github.com/rinatkh/classroom13/internal/user"
)

func newRouter() http.Handler {
	users := memory.NewUserRepository()
	notes := memory.NewNoteRepository()
	tokens := auth.NewJWTManager("test-secret", time.Hour)
	authUseCase := user.NewUseCase(users, auth.NewBcryptHasher(bcrypt.MinCost), tokens)
	noteUseCase := note.NewUseCase(notes)
	return httpapi.NewRouter(
		handler.NewAuthHandler(authUseCase),
		handler.NewNoteHandler(noteUseCase),
		middleware.NewAuth(tokens),
	)
}

func TestProtectedRoutesAndOwnership(t *testing.T) {
	router := newRouter()

	withoutToken := request(t, router, http.MethodGet, "/notes", "", "")
	assertStatus(t, withoutToken, http.StatusUnauthorized)

	alice := registerAndLogin(t, router, "alice@example.com")
	bob := registerAndLogin(t, router, "bob@example.com")

	created := request(t, router, http.MethodPost, "/notes", alice, `{"title":"private","text":"Alice only"}`)
	assertStatus(t, created, http.StatusCreated)
	if got := created.Header().Get("Location"); got != "/notes/1" {
		t.Fatalf("Location = %q", got)
	}

	foreign := request(t, router, http.MethodGet, "/notes/1", bob, "")
	assertStatus(t, foreign, http.StatusForbidden)

	aliceList := request(t, router, http.MethodGet, "/notes", alice, "")
	assertStatus(t, aliceList, http.StatusOK)
	var aliceNotes []map[string]any
	if err := json.Unmarshal(aliceList.Body.Bytes(), &aliceNotes); err != nil {
		t.Fatalf("decode Alice list: %v", err)
	}
	if len(aliceNotes) != 1 {
		t.Fatalf("Alice list length = %d", len(aliceNotes))
	}

	bobList := request(t, router, http.MethodGet, "/notes", bob, "")
	assertStatus(t, bobList, http.StatusOK)
	var bobNotes []map[string]any
	if err := json.Unmarshal(bobList.Body.Bytes(), &bobNotes); err != nil {
		t.Fatalf("decode Bob list: %v", err)
	}
	if len(bobNotes) != 0 {
		t.Fatalf("Bob can see foreign notes: %s", bobList.Body.String())
	}
}

func registerAndLogin(t *testing.T, router http.Handler, email string) string {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"password":"secret123"}`, email)
	created := request(t, router, http.MethodPost, "/register", "", body)
	assertStatus(t, created, http.StatusCreated)
	loggedIn := request(t, router, http.MethodPost, "/login", "", body)
	assertStatus(t, loggedIn, http.StatusOK)
	var response struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(loggedIn.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode login: %v", err)
	}
	if response.Token == "" {
		t.Fatal("empty token")
	}
	return response.Token
}

func request(t *testing.T, router http.Handler, method, target, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func assertStatus(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()
	if response.Code != want {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, want, response.Body.String())
	}
}
