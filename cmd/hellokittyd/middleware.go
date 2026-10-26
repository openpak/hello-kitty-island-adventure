package main

// Stubs for the two transports this backend hands the client to: Photon
// Realtime (Netcode-over-Photon rooms, served by openpak/photon) and Unity
// Gaming Services (player auth, Relay allocations, QoS, social — served by
// openpak/ugs). The stubs only know each service's base URL and report
// whether it answers /health; the invite resolve answer will carry the
// room/relay pointers once a capture shows their shape.

import (
	"io"
	"net/http"
	"time"
)

type middleware struct {
	name string
	url  string
	http *http.Client
}

func newMiddleware(name, url string) *middleware {
	return &middleware{name: name, url: url, http: &http.Client{Timeout: 3 * time.Second}}
}

// status reports "ok" when the service answers its health endpoint,
// otherwise a short reason. It never blocks startup: an unreachable
// transport is reported, not fatal.
func (m *middleware) status() string {
	if m.url == "" || m.url == "disabled" {
		return "disabled"
	}
	resp, err := m.http.Get(m.url + "/health")
	if err != nil {
		return "unreachable"
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
	if resp.StatusCode != http.StatusOK {
		return "unhealthy"
	}
	return "ok"
}
