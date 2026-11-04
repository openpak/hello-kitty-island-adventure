package main

// The invite-code rendezvous: the host allocates a short code, the guest
// resolves it to learn where to connect, either side disposes it when done.
// Codes expire after the configured TTL (the wiki-observed shape is minutes).
// Request and response shapes are provisional — method, query names and JSON
// keys come from no capture yet, so every spelling the client might use is
// read and the answers stay minimal. See docs/protocol.md.

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type invite struct {
	hostTag string
	expires time.Time
}

// inviteStore hands out codes; the codes and their expiry live in the store (Postgres).
type inviteStore struct {
	db  store
	ttl time.Duration
}

func newInviteStore(db store, ttl time.Duration) *inviteStore {
	return &inviteStore{db: db, ttl: ttl}
}

const codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

func newCode() string {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	for i, v := range b {
		b[i] = codeAlphabet[v%uint8(len(codeAlphabet))]
	}
	return string(b[:])
}

func normCode(code string) string { return strings.ToUpper(strings.TrimSpace(code)) }

func dbCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (st *inviteStore) allocate(hostTag string) (string, time.Time, error) {
	ctx, cancel := dbCtx()
	defer cancel()
	for {
		code := newCode()
		exp := time.Now().Add(st.ttl)
		ok, err := st.db.putInvite(ctx, code, hostTag, exp)
		if err != nil {
			return "", time.Time{}, err
		}
		if ok {
			return code, exp, nil
		}
	}
}

func (st *inviteStore) resolve(code string) (invite, bool, error) {
	ctx, cancel := dbCtx()
	defer cancel()
	host, exp, ok, err := st.db.getInvite(ctx, normCode(code))
	return invite{hostTag: host, expires: exp}, ok, err
}

func (st *inviteStore) dispose(code string) error {
	ctx, cancel := dbCtx()
	defer cancel()
	return st.db.delInvite(ctx, normCode(code))
}

// inviteCode reads the code from wherever the client put it: query names
// first, then a JSON body. All spellings are guesses until a capture lands.
func inviteCode(r *http.Request, body []byte) string {
	for _, k := range []string{"code", "inviteCode", "invite_code", "joinCode", "id"} {
		if v := r.URL.Query().Get(k); v != "" {
			return v
		}
	}
	var m map[string]any
	if json.Unmarshal(body, &m) == nil {
		for _, k := range []string{"code", "inviteCode", "invite_code", "joinCode", "id"} {
			if v, _ := m[k].(string); v != "" {
				return v
			}
		}
	}
	return ""
}

func bearerTag(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		h = r.Header.Get("X-Authorization")
	}
	return tag(strings.TrimPrefix(strings.TrimSpace(h), "Bearer "))
}

func (s *server) inviteAllocate(w http.ResponseWriter, r *http.Request, body []byte) {
	code, exp, err := s.invites.allocate(bearerTag(r))
	if err != nil {
		log.Printf("[hellokitty] invite allocate failed: %v", err)
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	log.Printf("[hellokitty] invite allocate -> %s ttl=%s", code, time.Until(exp).Round(time.Second))
	writeJSON(w, http.StatusOK, map[string]any{
		"code":      code,
		"expiresIn": int64(time.Until(exp).Seconds()),
	})
}

func (s *server) inviteResolve(w http.ResponseWriter, r *http.Request, body []byte) {
	code := inviteCode(r, body)
	inv, ok, err := s.invites.resolve(code)
	if err != nil {
		log.Printf("[hellokitty] invite resolve failed: %v", err)
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	log.Printf("[hellokitty] invite resolve %s -> found=%t", tag(code), ok)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]any{"found": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"found":     true,
		"host":      inv.hostTag,
		"expiresIn": int64(time.Until(inv.expires).Seconds()),
	})
}

func (s *server) inviteDispose(w http.ResponseWriter, r *http.Request, body []byte) {
	code := inviteCode(r, body)
	if err := s.invites.dispose(code); err != nil {
		log.Printf("[hellokitty] invite dispose failed: %v", err)
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	log.Printf("[hellokitty] invite dispose %s", tag(code))
	writeJSON(w, http.StatusOK, map[string]any{})
}
