// Package main runs a PingFederate admin API test server for CI.
// It serves the administrative accounts endpoints with in-memory state so
// that the connector's sync, grant, and revoke paths can be exercised
// without a real PingFederate tenant.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	// testUsername and testPassword are the hardcoded credentials the CI job passes via env vars.
	testUsername = "test-admin"
	testPassword = "test-password"

	port      = "8765"
	apiPrefix = "/pf-admin-api/v1"
)

// user mirrors client.PingFederateUser so responses match the shape the connector unmarshals.
// https://docs.pingidentity.com/r/en-us/pingfederate-112/pf_admin_api_reference_admin_accounts
type user struct {
	Username          string   `json:"username"`
	Email             string   `json:"emailAddress,omitempty"`
	EncryptedPassword string   `json:"encryptedPassword"`
	PhoneNumber       string   `json:"phoneNumber,omitempty"`
	Department        string   `json:"department,omitempty"`
	Description       string   `json:"description"`
	IsAuditor         bool     `json:"auditor"`
	IsActive          bool     `json:"active"`
	Roles             []string `json:"roles"`
}

type listUsersResponse struct {
	Items []*user `json:"items"`
}

// State is the in-memory store. All reads and writes go through its methods.
type State struct {
	mu    sync.Mutex
	users map[string]*user
}

// NewState seeds the server with enough data to exercise every connector code path:
//   - admin: active, no roles initially — the grant-revoke test adds/removes EXPRESSION_ADMINISTRATOR
//   - alice: active, holds two roles — ensures multiple roles appear in GetRoles
//   - bob:   active, holds one role with alice — ensures GetRoleAssignments returns >1 user
//   - carol: active, auditor flag set — exercises the IsAuditor branch
//   - dave:  inactive — exercises STATUS_DISABLED in userResource
func NewState() *State {
	s := &State{users: make(map[string]*user)}
	for _, u := range []*user{
		{Username: "admin", Email: "admin@example.com", Description: "Administrator", IsActive: true, Roles: []string{}},
		{Username: "alice", Email: "alice@example.com", Description: "Alice", IsActive: true, Roles: []string{"EXPRESSION_ADMINISTRATOR", "USER_ADMINISTRATOR"}},
		{Username: "bob", Email: "bob@example.com", Description: "Bob", IsActive: true, Roles: []string{"EXPRESSION_ADMINISTRATOR"}},
		{Username: "carol", Email: "carol@example.com", Description: "Carol", IsActive: true, IsAuditor: true, Roles: []string{"USER_ADMINISTRATOR"}},
		{Username: "dave", Email: "dave@example.com", Description: "Dave (disabled)", IsActive: false, Roles: []string{}},
	} {
		s.users[u.Username] = u
	}
	return s
}

func (s *State) listUsers() []*user {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*user, 0, len(s.users))
	for _, u := range s.users {
		cp := *u
		out = append(out, &cp)
	}
	return out
}

func (s *State) getUser(username string) (*user, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[username]
	if !ok {
		return nil, false
	}
	cp := *u
	return &cp, true
}

func (s *State) updateUser(u *user) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *u
	s.users[u.Username] = &cp
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// authMiddleware validates Basic auth and the X-XSRF-Header the connector always sends.
// https://docs.pingidentity.com/r/en-us/pingfederate-112/pf_admin_api_about_pingfederate_admin_api
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-XSRF-Header") == "" {
			writeJSON(w, http.StatusForbidden, map[string]string{"message": "X-XSRF-Header is required"})
			return
		}
		username, password, ok := r.BasicAuth()
		if !ok || username != testUsername || password != testPassword {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "invalid credentials"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	state := NewState()
	mux := http.NewServeMux()

	// GET /pf-admin-api/v1/administrativeAccounts
	// Lists all administrative accounts. Used by GetUsers, GetRoles, and GetRoleAssignments.
	// https://docs.pingidentity.com/r/en-us/pingfederate-112/pf_admin_api_reference_admin_accounts
	mux.HandleFunc(apiPrefix+"/administrativeAccounts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, listUsersResponse{Items: state.listUsers()})
	})

	// GET /pf-admin-api/v1/administrativeAccounts/{username}
	// PUT /pf-admin-api/v1/administrativeAccounts/{username}
	// Used by AddUserToRole and RemoveUserFromRole (GET then PUT pattern).
	// https://docs.pingidentity.com/r/en-us/pingfederate-112/pf_admin_api_reference_admin_accounts
	mux.HandleFunc(apiPrefix+"/administrativeAccounts/", func(w http.ResponseWriter, r *http.Request) {
		username := strings.TrimPrefix(r.URL.Path, apiPrefix+"/administrativeAccounts/")
		if username == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "username is required"})
			return
		}

		switch r.Method {
		case http.MethodGet:
			u, ok := state.getUser(username)
			if !ok {
				writeJSON(w, http.StatusNotFound, map[string]string{"message": fmt.Sprintf("account '%s' not found", username)})
				return
			}
			writeJSON(w, http.StatusOK, u)

		case http.MethodPut:
			if _, ok := state.getUser(username); !ok {
				writeJSON(w, http.StatusNotFound, map[string]string{"message": fmt.Sprintf("account '%s' not found", username)})
				return
			}
			var updated user
			if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid request body"})
				return
			}
			updated.Username = username
			state.updateUser(&updated)
			writeJSON(w, http.StatusOK, &updated)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	addr := ":" + port
	log.Printf("test-server: listening on %s", addr)
	if err := http.ListenAndServe(addr, authMiddleware(mux)); err != nil {
		log.Fatalf("test-server: %v", err)
	}
}
