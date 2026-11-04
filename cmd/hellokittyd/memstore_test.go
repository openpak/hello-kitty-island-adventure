package main

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"
)

// memStore is the store for handler tests; production only uses Postgres (store.go).
type memStore struct {
	mu      sync.Mutex
	invites map[string]invite
	saves   map[string][]byte
}

func newMemStore() *memStore {
	return &memStore{invites: map[string]invite{}, saves: map[string][]byte{}}
}

func (m *memStore) putInvite(_ context.Context, code, host string, exp time.Time) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if old, ok := m.invites[code]; ok && time.Now().Before(old.expires) {
		return false, nil
	}
	m.invites[code] = invite{host, exp}
	return true, nil
}

func (m *memStore) getInvite(_ context.Context, code string) (string, time.Time, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	inv, ok := m.invites[code]
	if !ok || time.Now().After(inv.expires) {
		delete(m.invites, code)
		return "", time.Time{}, false, nil
	}
	return inv.hostTag, inv.expires, true, nil
}

func (m *memStore) delInvite(_ context.Context, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.invites, code)
	return nil
}

func (m *memStore) putSave(_ context.Context, name, _ string, body []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saves[name] = body
	return nil
}

func (m *memStore) names() []string {
	var out []string
	for n := range m.saves {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func (m *memStore) latestSave(context.Context) ([]byte, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := m.names()
	if len(n) == 0 {
		return nil, false, nil
	}
	return m.saves[n[len(n)-1]], true, nil
}

func (m *memStore) saveNames(context.Context) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.names(), nil
}

func (m *memStore) delSaves(_ context.Context, tag string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for name := range m.saves {
		if strings.Contains(name, tag) {
			delete(m.saves, name)
			n++
		}
	}
	return n, nil
}
