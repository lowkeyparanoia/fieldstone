// Package api — integration tests.
// Tests real HTTP round-trips using httptest.Server + MemoryBackend.
// No external deps required. Run with: go test -v -race ./internal/api/...
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fieldstone/fieldstone/internal/auth"
	"github.com/fieldstone/fieldstone/internal/backend"
	"github.com/fieldstone/fieldstone/internal/jobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Test helpers ─────────────────────────────────────────────────────────────

func newTestServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	b := backend.NewMemoryBackend()
	authSvc := auth.NewService("test-secret-32-chars-minimum-key!!", 1*time.Hour, "test")
	store := jobs.NewMemoryStore()
	q := jobs.NewQueue(store, 2)
	// nil cache = no middleware caching = tests always see live backend state
	srv := NewServer(b, authSvc, q, nil)
	ts := httptest.NewServer(srv.Router())
	t.Cleanup(func() {
		ts.Close()
		q.Shutdown(context.Background()) //nolint:errcheck
	})
	return srv, ts
}

func doJSON(t *testing.T, method, url string, body interface{}, headers map[string]string) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req, err := http.NewRequest(method, url, &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, dst interface{}) {
	t.Helper()
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(dst))
}

func registerAndLogin(t *testing.T, ts *httptest.Server, email, password string) string {
	t.Helper()
	resp := doJSON(t, "POST", ts.URL+"/api/auth/register", map[string]string{
		"email": email, "password": password,
	}, nil)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var body map[string]interface{}
	decodeJSON(t, resp, &body)
	token, ok := body["token"].(string)
	require.True(t, ok, "no token in register response")
	return token
}

func bearer(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// ── Auth flow tests ───────────────────────────────────────────────────────────

func TestAuthFlow_RegisterLoginRefreshLogout(t *testing.T) {
	_, ts := newTestServer(t)

	// Register
	resp := doJSON(t, "POST", ts.URL+"/api/auth/register", map[string]string{
		"email": "user@test.com", "password": "password123!",
	}, nil)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var authResp map[string]interface{}
	decodeJSON(t, resp, &authResp)
	token := authResp["token"].(string)
	assert.NotEmpty(t, token)

	// Duplicate register → 409
	resp = doJSON(t, "POST", ts.URL+"/api/auth/register", map[string]string{
		"email": "user@test.com", "password": "password123!",
	}, nil)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	resp.Body.Close()

	// Login
	resp = doJSON(t, "POST", ts.URL+"/api/auth/login", map[string]string{
		"email": "user@test.com", "password": "password123!",
	}, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var loginResp map[string]interface{}
	decodeJSON(t, resp, &loginResp)
	loginToken := loginResp["token"].(string)
	assert.NotEmpty(t, loginToken)

	// Wrong password → 401
	resp = doJSON(t, "POST", ts.URL+"/api/auth/login", map[string]string{
		"email": "user@test.com", "password": "wrongpass",
	}, nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// Refresh
	resp = doJSON(t, "POST", ts.URL+"/api/auth/refresh", nil, bearer(loginToken))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var refreshResp map[string]interface{}
	decodeJSON(t, resp, &refreshResp)
	newToken := refreshResp["token"].(string)
	assert.NotEmpty(t, newToken)
	assert.NotEqual(t, loginToken, newToken, "refresh should issue a NEW token")

	// Old token should fail after refresh (key rotated)
	resp = doJSON(t, "POST", ts.URL+"/api/auth/refresh", nil, bearer(loginToken))
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "old token must be invalidated after refresh")
	resp.Body.Close()

	// Logout
	resp = doJSON(t, "POST", ts.URL+"/api/auth/logout", nil, bearer(newToken))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Refresh after logout → 401
	resp = doJSON(t, "POST", ts.URL+"/api/auth/refresh", nil, bearer(newToken))
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "token must be invalid after logout")
	resp.Body.Close()
}

func TestAuthMiddleware_ProtectsAllRoutes(t *testing.T) {
	_, ts := newTestServer(t)

	protectedGETs := []string{
		"/api/collections",
		"/api/users",
	}

	for _, path := range protectedGETs {
		t.Run("no_token_"+path, func(t *testing.T) {
			resp, err := http.Get(ts.URL + path)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
				"GET %s must require auth", path)
		})
	}
}

// ── Collection CRUD ──────────────────────────────────────────────────────────

func TestCollections_CRUD(t *testing.T) {
	_, ts := newTestServer(t)
	token := registerAndLogin(t, ts, "coll@test.com", "pass123!")

	// Create
	resp := doJSON(t, "POST", ts.URL+"/api/collections", map[string]interface{}{
		"name":   "articles",
		"fields": []map[string]string{{"name": "title", "type": "text"}},
	}, bearer(token))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var coll map[string]interface{}
	decodeJSON(t, resp, &coll)
	collID := coll["id"].(string)
	assert.Equal(t, "articles", coll["name"])

	// List
	req, _ := http.NewRequest("GET", ts.URL+"/api/collections", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	listResp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	var listBody map[string]interface{}
	decodeJSON(t, listResp, &listBody)
	items := listBody["items"].([]interface{})
	assert.Len(t, items, 1)

	// Duplicate name → 409
	resp = doJSON(t, "POST", ts.URL+"/api/collections", map[string]interface{}{
		"name": "articles", "fields": []interface{}{},
	}, bearer(token))
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	resp.Body.Close()

	// Get by ID
	req, _ = http.NewRequest("GET", ts.URL+"/api/collections/"+collID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	getResp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	getResp.Body.Close()

	// Delete
	req, _ = http.NewRequest("DELETE", ts.URL+"/api/collections/"+collID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	delResp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusOK, delResp.StatusCode)
	delResp.Body.Close()

	// List after delete → empty
	req, _ = http.NewRequest("GET", ts.URL+"/api/collections", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	listResp2, _ := http.DefaultClient.Do(req)
	var listBody2 map[string]interface{}
	decodeJSON(t, listResp2, &listBody2)
	items2 := listBody2["items"].([]interface{})
	assert.Len(t, items2, 0)
}

// ── Record CRUD ───────────────────────────────────────────────────────────────

func TestRecords_CRUD_And_Pagination(t *testing.T) {
	_, ts := newTestServer(t)
	token := registerAndLogin(t, ts, "rec@test.com", "pass123!")

	// Create collection
	resp := doJSON(t, "POST", ts.URL+"/api/collections", map[string]interface{}{
		"name":   "songs",
		"fields": []map[string]string{{"name": "title", "type": "text"}, {"name": "bpm", "type": "number"}},
	}, bearer(token))
	var coll map[string]interface{}
	decodeJSON(t, resp, &coll)
	collID := coll["id"].(string)

	// Create 5 records
	var recIDs []string
	for i := 0; i < 5; i++ {
		r := doJSON(t, "POST", ts.URL+"/api/collections/"+collID+"/records",
			map[string]interface{}{"data": map[string]interface{}{
				"title": fmt.Sprintf("Song %d", i+1), "bpm": 100 + i,
			}}, bearer(token))
		require.Equal(t, http.StatusCreated, r.StatusCode)
		var rec map[string]interface{}
		decodeJSON(t, r, &rec)
		recIDs = append(recIDs, rec["id"].(string))
	}
	assert.Len(t, recIDs, 5)

	// List all
	req, _ := http.NewRequest("GET", ts.URL+"/api/collections/"+collID+"/records", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	listResp, _ := http.DefaultClient.Do(req)
	var listBody map[string]interface{}
	decodeJSON(t, listResp, &listBody)
	assert.Equal(t, float64(5), listBody["total"])

	// Paginate — page 1, perPage 2
	req, _ = http.NewRequest("GET", ts.URL+"/api/collections/"+collID+"/records?page=1&perPage=2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	p1Resp, _ := http.DefaultClient.Do(req)
	var p1Body map[string]interface{}
	decodeJSON(t, p1Resp, &p1Body)
	p1Items := p1Body["items"].([]interface{})
	assert.Len(t, p1Items, 2)
	assert.Equal(t, float64(3), p1Body["totalPages"])

	// Get single
	req, _ = http.NewRequest("GET", ts.URL+"/api/collections/"+collID+"/records/"+recIDs[0], nil)
	req.Header.Set("Authorization", "Bearer "+token)
	getResp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	getResp.Body.Close()

	// Update
	upResp := doJSON(t, "PUT", ts.URL+"/api/collections/"+collID+"/records/"+recIDs[0],
		map[string]interface{}{"data": map[string]interface{}{"title": "Updated Song", "bpm": 999}},
		bearer(token))
	assert.Equal(t, http.StatusOK, upResp.StatusCode)
	var upBody map[string]interface{}
	decodeJSON(t, upResp, &upBody)
	data := upBody["data"].(map[string]interface{})
	assert.Equal(t, "Updated Song", data["title"])

	// Delete
	req, _ = http.NewRequest("DELETE", ts.URL+"/api/collections/"+collID+"/records/"+recIDs[0], nil)
	req.Header.Set("Authorization", "Bearer "+token)
	delResp, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusOK, delResp.StatusCode)
	delResp.Body.Close()

	// Verify deleted → 404
	req, _ = http.NewRequest("GET", ts.URL+"/api/collections/"+collID+"/records/"+recIDs[0], nil)
	req.Header.Set("Authorization", "Bearer "+token)
	gone, _ := http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusNotFound, gone.StatusCode)
	gone.Body.Close()
}

// ── Multi-tenancy isolation ───────────────────────────────────────────────────

func TestMultiTenancy_Isolation(t *testing.T) {
	_, ts := newTestServer(t)

	// Register users in two different tenants
	t1Token := func() string {
		req, _ := http.NewRequest("POST", ts.URL+"/api/auth/register", bytes.NewBufferString(`{"email":"a@t1.com","password":"pass123!"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant1")
		resp, _ := http.DefaultClient.Do(req)
		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		return body["token"].(string)
	}()

	t2Token := func() string {
		req, _ := http.NewRequest("POST", ts.URL+"/api/auth/register", bytes.NewBufferString(`{"email":"b@t2.com","password":"pass123!"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", "tenant2")
		resp, _ := http.DefaultClient.Do(req)
		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		return body["token"].(string)
	}()

	// Create collection in tenant1
	resp := doJSON(t, "POST", ts.URL+"/api/collections", map[string]interface{}{
		"name": "secret-t1", "fields": []interface{}{},
	}, bearer(t1Token))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// tenant2 sees ZERO collections — full isolation
	req, _ := http.NewRequest("GET", ts.URL+"/api/collections", nil)
	req.Header.Set("Authorization", "Bearer "+t2Token)
	listResp, _ := http.DefaultClient.Do(req)
	var listBody map[string]interface{}
	json.NewDecoder(listResp.Body).Decode(&listBody)
	listResp.Body.Close()
	items := listBody["items"].([]interface{})
	assert.Len(t, items, 0, "tenant2 must NOT see tenant1 collections")

	// tenant1 sees its own collection
	req, _ = http.NewRequest("GET", ts.URL+"/api/collections", nil)
	req.Header.Set("Authorization", "Bearer "+t1Token)
	ownResp, _ := http.DefaultClient.Do(req)
	var ownBody map[string]interface{}
	json.NewDecoder(ownResp.Body).Decode(&ownBody)
	ownResp.Body.Close()
	ownItems := ownBody["items"].([]interface{})
	assert.Len(t, ownItems, 1, "tenant1 must see its own collection")
}

// ── Validation / error cases ─────────────────────────────────────────────────

func TestValidation_ErrorCodes(t *testing.T) {
	_, ts := newTestServer(t)

	tests := []struct {
		name       string
		method     string
		path       string
		body       interface{}
		wantStatus int
	}{
		{"empty_email", "POST", "/api/auth/register", map[string]string{"email": "", "password": "x"}, http.StatusBadRequest},
		{"missing_password", "POST", "/api/auth/register", map[string]string{"email": "x@x.com"}, http.StatusBadRequest},
		{"bad_content", "POST", "/api/auth/login", map[string]string{"email": "nobody@x.com", "password": "wrong"}, http.StatusUnauthorized},
		{"no_auth_collections", "GET", "/api/collections", nil, http.StatusUnauthorized},
		{"no_auth_users", "GET", "/api/users", nil, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := doJSON(t, tt.method, ts.URL+tt.path, tt.body, nil)
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

// ── Health endpoint ───────────────────────────────────────────────────────────

func TestHealth_NoAuthRequired(t *testing.T) {
	_, ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "healthy", body["status"])
}

// ── Concurrency / race test ───────────────────────────────────────────────────

func TestConcurrentRecordCreation_NoRace(t *testing.T) {
	_, ts := newTestServer(t)
	token := registerAndLogin(t, ts, "race@test.com", "pass123!")

	// Create collection
	resp := doJSON(t, "POST", ts.URL+"/api/collections", map[string]interface{}{
		"name": "concurrent", "fields": []interface{}{},
	}, bearer(token))
	var coll map[string]interface{}
	decodeJSON(t, resp, &coll)
	collID := coll["id"].(string)

	// Hammer 20 concurrent record creates
	done := make(chan bool, 20)
	for i := 0; i < 20; i++ {
		go func(n int) {
			r := doJSON(t, "POST", ts.URL+"/api/collections/"+collID+"/records",
				map[string]interface{}{"data": map[string]interface{}{"n": n}},
				bearer(token))
			r.Body.Close()
			done <- true
		}(i)
	}
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify all 20 created
	req, _ := http.NewRequest("GET", ts.URL+"/api/collections/"+collID+"/records?perPage=100", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	listResp, _ := http.DefaultClient.Do(req)
	var listBody map[string]interface{}
	decodeJSON(t, listResp, &listBody)
	assert.Equal(t, float64(20), listBody["total"])
}
