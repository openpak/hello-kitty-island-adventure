package main

// The invite-code rendezvous: the host allocates a short code, the guest
// resolves it to learn where to connect, either side disposes it when done.
// Codes expire after the configured TTL (the wiki-observed shape is minutes).
// Request and response shapes are provisional — method, query names and JSON
// keys come from no capture yet, so every spelling the client might use is
// read and the answers stay minimal. See docs/protocol.md.

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type invite struct {
	hostTag string
	expires time.Time
}

type inviteStore struct {
	mu    sync.Mutex
	ttl   time.Duration
	codes map[string]invite
}

func newInviteStore(ttl time.Duration) *inviteStore {
	return &inviteStore{ttl: ttl, codes: map[string]invite{}}
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

func (st *inviteStore) allocate(hostTag string) (string, time.Time) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.sweepLocked()
	for {
		code := newCode()
		if _, taken := st.codes[code]; taken {
			continue
		}
		exp := time.Now().Add(st.ttl)
		st.codes[code] = invite{hostTag: hostTag, expires: exp}
		return code, exp
	}
}

func (st *inviteStore) resolve(code string) (invite, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	inv, ok := st.codes[strings.ToUpper(strings.TrimSpace(code))]
	if !ok || time.Now().After(inv.expires) {
		delete(st.codes, strings.ToUpper(strings.TrimSpace(code)))
		return invite{}, false
	}
	return inv, true
}

func (st *inviteStore) dispose(code string) {
	st.mu.Lock()
	defer st.mu.Unlock()
	delete(st.codes, strings.ToUpper(strings.TrimSpace(code)))
}

func (st *inviteStore) sweepLocked() {
	now := time.Now()
	for code, inv := range st.codes {
		if now.After(inv.expires) {
			delete(st.codes, code)
		}
	}
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
	code, exp := s.invites.allocate(bearerTag(r))
	log.Printf("[hellokitty] invite allocate -> %s ttl=%s", code, time.Until(exp).Round(time.Second))
	writeJSON(w, http.StatusOK, map[string]any{
		"code":      code,
		"expiresIn": int64(time.Until(exp).Seconds()),
	})
}

func (s *server) inviteResolve(w http.ResponseWriter, r *http.Request, body []byte) {
	code := inviteCode(r, body)
	inv, ok := s.invites.resolve(code)
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
	s.invites.dispose(code)
	log.Printf("[hellokitty] invite dispose %s", tag(code))
	writeJSON(w, http.StatusOK, map[string]any{})
}
