// Command hellokittyd answers Hello Kitty Island Adventure's first-party
// backend (Sunblink REST on three AWS API-Gateway deployments plus
// minsc.sunblink.io) for the Switch build (010027901C89C000, Unity
// 2021.3.43f1 IL2CPP) when a console is redirected to OpenPak.
//
// Gameplay itself is Unity Netcode over Photon Realtime (served by
// platforms/switch/games/photon) with Unity Relay as the alternate transport
// (served by platforms/switch/games/ugs); this service is only the
// developer's small HTTPS API the client calls around it:
//
//	invitecode/allocate|dispose|resolve   2-player invite-code rendezvous
//	savegames/delete|get|query|upload     cloud saves
//	servers/settings                      boot config (Photon AppId/region, Relay)
//	support/claim-save-data|offer-save-data|send-device-logs
//	session-token/...                     token flow (shape still unknown)
//
// Each deployment carries an API-Gateway stage prefix (/live or /prod) that
// the service strips before routing. The service listens on plain HTTP
// (-addr) behind whatever terminates TLS. Routes it does not know yet are
// logged with their body and answered with an empty JSON object, so a
// capture shows what the client asks next.
//
// The contract was read out of the Switch executable's managed metadata;
// see docs/protocol.md. Response shapes are provisional until a live
// capture confirms them.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	addr       = flag.String("addr", ":23560", "plain HTTP listen address (behind the TLS terminator)")
	healthAddr = flag.String("health-addr", ":23561", "health listen address")
	dataDir    = flag.String("data", "./data", "directory for kept uploads and save blobs")
	settings   = flag.String("settings", "", "servers/settings JSON to serve instead of the built-in one")
	codeTTL    = flag.Duration("code-ttl", 5*time.Minute, "how long an allocated invite code stays resolvable")
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	flag.Parse()
	s, err := newServer(*dataDir, *settings, *codeTTL)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		h := http.NewServeMux()
		h.HandleFunc("/healthz", s.health)
		h.HandleFunc("/health", s.health)
		log.Printf("[hellokitty] health on %s", *healthAddr)
		log.Fatal(http.ListenAndServe(*healthAddr, h))
	}()
	log.Printf("[hellokitty] serving HTTP on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, s))
}

func (s *server) health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","photon":%q,"ugs":%q}`+"\n", s.photon.status(), s.ugs.status())
}
