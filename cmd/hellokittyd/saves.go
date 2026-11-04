package main

// Cloud saves, kept whole in Postgres (hk_saves): an upload is stored under a timestamped name so a
// later get can serve it back and a capture shows the shape. Query and get filters are still unknown,
// so get answers the most recent blob (empty object when there is none) and query lists what is kept.
// Delete removes the caller's blobs.

import (
	"log"
	"net/http"
	"time"
)

func (s *server) saveUpload(w http.ResponseWriter, r *http.Request, body []byte) {
	ctx, cancel := dbCtx()
	defer cancel()
	tag := bearerTag(r)
	name := time.Now().UTC().Format("20060102T150405.000000") + "-" + tag + ".bin"
	if err := s.db.putSave(ctx, name, tag, body); err != nil {
		log.Printf("[hellokitty] save upload failed: %v", err)
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	log.Printf("[hellokitty] save upload %s q=%s %dB", r.Host, tagQuery(r), len(body))
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *server) saveGet(w http.ResponseWriter, r *http.Request, body []byte) {
	ctx, cancel := dbCtx()
	defer cancel()
	raw, ok, err := s.db.latestSave(ctx)
	if err != nil {
		log.Printf("[hellokitty] save get failed: %v", err)
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	log.Printf("[hellokitty] save get %s q=%s found=%t", r.Host, tagQuery(r), ok)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{})
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(raw)
}

func (s *server) saveQuery(w http.ResponseWriter, r *http.Request, body []byte) {
	ctx, cancel := dbCtx()
	defer cancel()
	names, err := s.db.saveNames(ctx)
	if err != nil {
		log.Printf("[hellokitty] save query failed: %v", err)
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	log.Printf("[hellokitty] save query %s q=%s kept=%d", r.Host, tagQuery(r), len(names))
	writeJSON(w, http.StatusOK, map[string]any{"saves": names})
}

func (s *server) saveDelete(w http.ResponseWriter, r *http.Request, body []byte) {
	ctx, cancel := dbCtx()
	defer cancel()
	removed, err := s.db.delSaves(ctx, bearerTag(r))
	if err != nil {
		log.Printf("[hellokitty] save delete failed: %v", err)
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	log.Printf("[hellokitty] save delete %s q=%s removed=%d", r.Host, tagQuery(r), removed)
	w.WriteHeader(http.StatusOK)
}
