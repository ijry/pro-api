package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func mockGitHub(t *testing.T, opts mockOpts) (baseURL string, apiBaseURL string) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if opts.badCode && r.Form.Get("code") == "bad" {
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "bad_verification_code"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"access_token": "tok123",
			"token_type":   "bearer",
			"scope":        "read:user user:email",
		})
	})
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			http.Error(w, "no auth", 401)
			return
		}
		user := map[string]any{
			"id":         12345,
			"login":      "alice",
			"name":       "Alice",
			"avatar_url": "https://gh.com/a.png",
		}
		if !opts.noEmail {
			user["email"] = "alice@example.com"
		}
		_ = json.NewEncoder(w).Encode(user)
	})
	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"email": "alice@personal.com", "primary": false, "verified": true},
			{"email": "alice@fallback.com", "primary": true, "verified": true},
		})
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts.URL, ts.URL
}

type mockOpts struct {
	badCode bool
	noEmail bool
}

func TestBuildAuthURL_ContainsExpectedParams(t *testing.T) {
	p := New(Config{ClientID: "cid", BaseURL: "https://gh.example"})
	got, err := p.BuildAuthURL(context.Background(), "state-x", "https://app.example/cb")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(got)
	if u.Query().Get("client_id") != "cid" {
		t.Fatalf("client_id missing: %s", got)
	}
	if u.Query().Get("state") != "state-x" {
		t.Fatalf("state missing: %s", got)
	}
	if u.Query().Get("scope") != "read:user user:email" {
		t.Fatalf("scope wrong: %s", u.Query().Get("scope"))
	}
}

func TestBuildAuthURL_NoClientID(t *testing.T) {
	p := New(Config{})
	if _, err := p.BuildAuthURL(context.Background(), "s", "r"); err == nil {
		t.Fatal("want err when client_id missing")
	}
}

func TestExchange_Success(t *testing.T) {
	base, api := mockGitHub(t, mockOpts{})
	p := New(Config{ClientID: "cid", ClientSecret: "sec", BaseURL: base, APIBaseURL: api})
	info, tok, err := p.Exchange(context.Background(), "good", "https://app.example/cb")
	if err != nil {
		t.Fatal(err)
	}
	if tok != "tok123" {
		t.Fatalf("token wrong: %q", tok)
	}
	if info.ID != 12345 || info.Login != "alice" || info.Email != "alice@example.com" {
		t.Fatalf("info wrong: %+v", info)
	}
}

func TestExchange_NoEmailFallsBackToUserEmails(t *testing.T) {
	base, api := mockGitHub(t, mockOpts{noEmail: true})
	p := New(Config{ClientID: "cid", ClientSecret: "sec", BaseURL: base, APIBaseURL: api})
	info, _, err := p.Exchange(context.Background(), "good", "")
	if err != nil {
		t.Fatal(err)
	}
	if info.Email != "alice@fallback.com" {
		t.Fatalf("want primary fallback email, got %q", info.Email)
	}
}

func TestExchange_BadCodeReturnsError(t *testing.T) {
	base, api := mockGitHub(t, mockOpts{badCode: true})
	p := New(Config{ClientID: "cid", ClientSecret: "sec", BaseURL: base, APIBaseURL: api})
	if _, _, err := p.Exchange(context.Background(), "bad", ""); err == nil {
		t.Fatal("want err")
	}
}

func TestExchange_NoClientID(t *testing.T) {
	p := New(Config{})
	if _, _, err := p.Exchange(context.Background(), "good", ""); err == nil {
		t.Fatal("want err")
	}
}
