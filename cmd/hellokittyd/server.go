package main

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// sunblinkHosts is the first-party surface measured in the Switch build's
// managed metadata (docs/provenance.md). Paths only — the stage each
// deployment serves under (/live on the first, /prod on the rest) is
// stripped before routing.
var sunblinkHosts = map[string]bool{
	"61cg23vc38.execute-api.us-east-1.amazonaws.com": true,
	"gkbkwtb0ad.execute-api.us-east-1.amazonaws.com": true,
	"shlax276jb.execute-api.us-east-1.amazonaws.com": true,
	"minsc.sunblink.io": true,
}

type server struct {
	dataDir  string
	settings json.RawMessage
	invites  *inviteStore
	photon   *middleware
	ugs      *middleware
}

func newServer(dataDir, settingsFile string, ttl time.Duration) (*server, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	settings := builtinSettings()
	if settingsFile != "" {
		raw, err := os.ReadFile(settingsFile)
		if err != nil {
			return nil, err
		}
		if !json.Valid(raw) {
			return nil, fmt.Errorf("settings file is not valid JSON")
		}
		settings = raw
	}
	return &server{
		dataDir:  dataDir,
		settings: settings,
		invites:  newInviteStore(ttl),
		photon:   newMiddleware("photon", envOr("PHOTON_URL", "http://127.0.0.1:8080")),
		ugs:      newMiddleware("ugs", envOr("UGS_URL", "http://127.0.0.1:22681")),
	}, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := strings.ToLower(r.Host)
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}
	if !sunblinkHosts[host] {
		log.Printf("[hellokitty] !! unknown host %s%s", r.Host, r.URL.RequestURI())
		http.NotFound(w, r)
		return
	}
	body, err := readBody(r)
	if err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	path := stripStage(r.URL.Path)
	log.Printf("[hellokitty] <- %s %s%s len=%d", r.Method, host, r.URL.RequestURI(), len(body))
	switch {
	case path == "servers/settings":
		s.serveSettings(w, r)
	case path == "invitecode/allocate":
		s.inviteAllocate(w, r, body)
	case path == "invitecode/resolve":
		s.inviteResolve(w, r, body)
	case path == "invitecode/dispose":
		s.inviteDispose(w, r, body)
	case path == "savegames/upload":
		s.saveUpload(w, r, body)
	case path == "savegames/get":
		s.saveGet(w, r, body)
	case path == "savegames/query":
		s.saveQuery(w, r, body)
	case path == "savegames/delete":
		s.saveDelete(w, r, body)
	case strings.HasPrefix(path, "support/claim-save-data") ||
		strings.HasPrefix(path, "support/offer-save-data") ||
		strings.HasPrefix(path, "support/send-device-logs"):
		logRequest(r, body)
		writeJSON(w, http.StatusOK, map[string]any{})
	case strings.HasPrefix(path, "session-token"):
		// The token flow's shape is still unknown (see docs/protocol.md):
		// log the call and answer empty so a capture shows what comes next.
		logRequest(r, body)
		writeJSON(w, http.StatusOK, map[string]any{})
	default:
		logRequest(r, body)
		writeJSON(w, http.StatusOK, map[string]any{})
	}
}

// stripStage drops the one API-Gateway stage segment (/live, /prod) that
// prefixes every path on these deployments.
func stripStage(p string) string {
	p = strings.TrimPrefix(p, "/")
	if rest, ok := strings.CutPrefix(p, "live/"); ok {
		return rest
	}
	if rest, ok := strings.CutPrefix(p, "prod/"); ok {
		return rest
	}
	return p
}

// builtinSettings is the provisional servers/settings body: the keys the
// client reads are still unknown, so it carries the Photon and Relay
// pointers the RECON says this endpoint hands out, from the environment.
func builtinSettings() json.RawMessage {
	raw, _ := json.Marshal(map[string]any{
		"photonAppId":  envOr("HKIA_PHOTON_APPID", ""),
		"photonRegion": envOr("HKIA_PHOTON_REGION", ""),
		"relayConfig":  map[string]any{},
	})
	return raw
}

func (s *server) serveSettings(w http.ResponseWriter, r *http.Request) {
	log.Printf("[hellokitty] settings host=%s q=%s", r.Host, tagQuery(r))
	w.Header().Set("Content-Type", "application/json")
	w.Write(s.settings)
}

// readBody returns the request body, inflating it when the client compressed it.
func readBody(r *http.Request) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(r.Header.Get("Content-Encoding"), "gzip") || (len(b) > 2 && b[0] == 0x1f && b[1] == 0x8b) {
		zr, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		return io.ReadAll(io.LimitReader(zr, 32<<20))
	}
	return b, nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// logRequest prints one request with its body, cut short and with tokens
// replaced by tags: bodies may carry access and refresh tokens.
func logRequest(r *http.Request, body []byte) {
	b := string(body)
	if len(b) > 600 {
		b = b[:600] + "..."
	}
	log.Printf("[hellokitty] %s %s%s q=%s body=%q", r.Method, r.Host, r.URL.Path, tagQuery(r), b)
}

// tagKey is random per process: tags correlate lines within one run and
// cannot be reversed or compared across restarts.
var tagKey = func() []byte {
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		panic(err)
	}
	return k
}()

func tag(v string) string {
	if v == "" {
		return "-"
	}
	m := hmac.New(sha256.New, tagKey)
	m.Write([]byte(v))
	return "~" + hex.EncodeToString(m.Sum(nil)[:3])
}

// tagQuery renders the request's query with every value replaced by a tag:
// codes, tokens and save ids must never land in logs raw.
func tagQuery(r *http.Request) string {
	q := r.URL.Query()
	if len(q) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(q))
	for k, vs := range q {
		for _, v := range vs {
			parts = append(parts, k+"="+tag(v))
		}
	}
	return strings.Join(parts, "&")
}
