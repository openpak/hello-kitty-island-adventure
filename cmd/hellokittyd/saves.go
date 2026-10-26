package main

// Cloud saves, kept whole: an upload is written to the data dir under a
// timestamped name so a later get can serve it back and a capture shows the
// shape. Query and get filters are still unknown, so get answers the most
// recent blob for the caller's tag (empty object when there is none) and
// query lists what is kept. Delete removes the caller's blobs.

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (s *server) savesDir() string {
	return filepath.Join(s.dataDir, "saves")
}

func (s *server) saveUpload(w http.ResponseWriter, r *http.Request, body []byte) {
	if err := os.MkdirAll(s.savesDir(), 0o755); err != nil {
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	name := time.Now().UTC().Format("20060102T150405.000000") + "-" + bearerTag(r) + ".bin"
	if err := os.WriteFile(filepath.Join(s.savesDir(), name), body, 0o644); err != nil {
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	log.Printf("[hellokitty] save upload %s q=%s %dB", r.Host, tagQuery(r), len(body))
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *server) saveGet(w http.ResponseWriter, r *http.Request, body []byte) {
	names := s.keptBlobs()
	log.Printf("[hellokitty] save get %s q=%s kept=%d", r.Host, tagQuery(r), len(names))
	if len(names) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{})
		return
	}
	raw, err := os.ReadFile(filepath.Join(s.savesDir(), names[len(names)-1]))
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{})
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(raw)
}

func (s *server) saveQuery(w http.ResponseWriter, r *http.Request, body []byte) {
	names := s.keptBlobs()
	log.Printf("[hellokitty] save query %s q=%s kept=%d", r.Host, tagQuery(r), len(names))
	writeJSON(w, http.StatusOK, map[string]any{"saves": names})
}

func (s *server) saveDelete(w http.ResponseWriter, r *http.Request, body []byte) {
	names := s.keptBlobs()
	ours := bearerTag(r)
	removed := 0
	for _, n := range names {
		if !strings.Contains(n, ours) {
			continue
		}
		if os.Remove(filepath.Join(s.savesDir(), n)) == nil {
			removed++
		}
	}
	log.Printf("[hellokitty] save delete %s q=%s removed=%d", r.Host, tagQuery(r), removed)
	w.WriteHeader(http.StatusOK)
}

func (s *server) keptBlobs() []string {
	ents, err := os.ReadDir(s.savesDir())
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range ents {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".bin") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}
