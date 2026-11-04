package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

// Needs a scratch Postgres: HELLOKITTY_TEST_DATABASE_URL; skipped otherwise.
func TestStateSurvivesRestartInPostgres(t *testing.T) {
	url := os.Getenv("HELLOKITTY_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("HELLOKITTY_TEST_DATABASE_URL unset")
	}
	ctx := context.Background()
	open := func(ttl time.Duration) *server {
		db, err := openStore(ctx, url)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(db.pool.Close)
		s, err := newServer(db, "", ttl)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	a := open(time.Minute)
	a.db.(*pgStore).pool.Exec(ctx, `DELETE FROM hk_invites; DELETE FROM hk_saves`)
	w := serve(a, gwHost, "/prod/invitecode/allocate", "")
	var got struct{ Code string }
	json.Unmarshal(w.Body.Bytes(), &got)
	if got.Code == "" {
		t.Fatalf("allocate: %d %s", w.Code, w.Body.String())
	}
	short := serve(a, gwHost, "/prod/savegames/upload", `{"slot":1}`)
	if short.Code != 200 {
		t.Fatalf("upload: %d", short.Code)
	}
	b := open(time.Minute) // restart: a new server on a new pool
	if w := serve(b, gwHost, "/prod/invitecode/resolve?code="+got.Code, ""); w.Code != 200 {
		t.Fatalf("code lost on restart: %d", w.Code)
	}
	if w := serve(b, gwHost, "/prod/savegames/get", ""); w.Code != 200 || w.Body.String() != `{"slot":1}` {
		t.Fatalf("save lost on restart: %d %q", w.Code, w.Body.String())
	}
	// The TTL is kept in the database: a code past its expiry stops resolving after a restart too.
	exp := open(time.Millisecond)
	w = serve(exp, gwHost, "/prod/invitecode/allocate", "")
	var old struct{ Code string }
	json.Unmarshal(w.Body.Bytes(), &old)
	time.Sleep(20 * time.Millisecond)
	c := open(time.Minute)
	if w := serve(c, gwHost, "/prod/invitecode/resolve?code="+old.Code, ""); w.Code != http.StatusNotFound {
		t.Fatalf("expired code resolves: %d", w.Code)
	}
	serve(c, gwHost, "/prod/savegames/delete", "")
}
