package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testServer(t *testing.T) *server {
	t.Helper()
	s, err := newServer(t.TempDir(), "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func serve(s *server, host, target string, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(http.MethodGet, target, nil)
	} else {
		r = httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	}
	r.Host = host
	s.ServeHTTP(w, r)
	return w
}

const gwHost = "gkbkwtb0ad.execute-api.us-east-1.amazonaws.com"

func TestUnknownHostRejected(t *testing.T) {
	w := serve(testServer(t), "evil.example.com", "/prod/servers/settings", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown host: %d", w.Code)
	}
}

func TestStagePrefixesRoute(t *testing.T) {
	s := testServer(t)
	for _, target := range []string{"/live/servers/settings", "/prod/servers/settings", "/servers/settings"} {
		w := serve(s, gwHost, target, "")
		if w.Code != 200 {
			t.Errorf("%s: %d", target, w.Code)
		}
	}
	// minsc answers too: it is the fourth first-party host.
	if w := serve(s, "minsc.sunblink.io", "/prod/invitecode/allocate", ""); w.Code != 200 {
		t.Errorf("minsc allocate: %d", w.Code)
	}
}

func TestInviteRoundTrip(t *testing.T) {
	s := testServer(t)
	w := serve(s, gwHost, "/prod/invitecode/allocate", "")
	var got struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got.Code == "" {
		t.Fatalf("allocate: %d %q", w.Code, w.Body.String())
	}
	w = serve(s, gwHost, "/prod/invitecode/resolve?code="+got.Code, "")
	var found struct {
		Found bool `json:"found"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &found); err != nil || !found.Found {
		t.Fatalf("resolve: %d %q", w.Code, w.Body.String())
	}
	if w := serve(s, gwHost, "/prod/invitecode/resolve?code=NOPE99", ""); w.Code != http.StatusNotFound {
		t.Fatalf("resolve unknown: %d", w.Code)
	}
	if w := serve(s, gwHost, "/prod/invitecode/dispose?code="+got.Code, ""); w.Code != 200 {
		t.Fatalf("dispose: %d", w.Code)
	}
	if w := serve(s, gwHost, "/prod/invitecode/resolve?code="+got.Code, ""); w.Code != http.StatusNotFound {
		t.Fatalf("resolve after dispose: %d", w.Code)
	}
}

func TestInviteCodesExpire(t *testing.T) {
	s, err := newServer(t.TempDir(), "", time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	w := serve(s, gwHost, "/prod/invitecode/allocate", "")
	var got struct {
		Code string `json:"code"`
	}
	json.Unmarshal(w.Body.Bytes(), &got)
	time.Sleep(5 * time.Millisecond)
	if w := serve(s, gwHost, "/prod/invitecode/resolve?code="+got.Code, ""); w.Code != http.StatusNotFound {
		t.Fatalf("expired code still resolves: %d", w.Code)
	}
}

func TestSaveRoundTrip(t *testing.T) {
	s := testServer(t)
	payload := `{"slot":1,"data":"aGVsbG8="}`
	if w := serve(s, gwHost, "/prod/savegames/upload", payload); w.Code != 200 {
		t.Fatalf("upload: %d", w.Code)
	}
	if w := serve(s, gwHost, "/prod/savegames/get", ""); w.Code != 200 || w.Body.String() != payload {
		t.Fatalf("get: %d %q", w.Code, w.Body.String())
	}
	if w := serve(s, gwHost, "/prod/savegames/query", ""); w.Code != 200 {
		t.Fatalf("query: %d", w.Code)
	}
	if w := serve(s, gwHost, "/prod/savegames/delete", ""); w.Code != 200 {
		t.Fatalf("delete: %d", w.Code)
	}
}

func TestUnknownRouteIsEmptyJSON(t *testing.T) {
	s := testServer(t)
	for _, p := range []string{"/prod/support/send-device-logs", "/prod/session-token/refresh", "/prod/something/new"} {
		w := serve(s, gwHost, p, `{"a":1}`)
		if w.Code != 200 || strings.TrimSpace(w.Body.String()) != "{}" {
			t.Errorf("%s: %d %q", p, w.Code, w.Body.String())
		}
	}
}

func TestHealth(t *testing.T) {
	s := testServer(t)
	w := httptest.NewRecorder()
	s.health(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	var h struct {
		Status string `json:"status"`
		Photon string `json:"photon"`
		UGS    string `json:"ugs"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &h); err != nil || h.Status != "ok" {
		t.Fatalf("health: %q", w.Body.String())
	}
	if h.Photon == "" || h.UGS == "" {
		t.Fatalf("health omits transports: %q", w.Body.String())
	}
}
