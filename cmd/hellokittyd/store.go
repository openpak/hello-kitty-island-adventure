package main

// Durable state in Postgres (HELLOKITTY_DATABASE_URL, required): invite codes with their expiry and the
// cloud save blobs. Nothing is kept in memory or in files, so a restart loses neither. The pool is
// capped at three connections (openpak/prds#85).

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schema = `CREATE TABLE IF NOT EXISTS hk_invites (
	code       text PRIMARY KEY,
	host_tag   text NOT NULL,
	expires_at timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS hk_invites_expires ON hk_invites (expires_at);
CREATE TABLE IF NOT EXISTS hk_saves (
	name       text PRIMARY KEY,
	tag        text NOT NULL,
	body       bytea NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now()
)`

// store is what the handlers keep their state in.
type store interface {
	// putInvite stores a fresh code; false means the code is taken (still live).
	putInvite(ctx context.Context, code, hostTag string, expires time.Time) (bool, error)
	// getInvite returns a live code's host and expiry; expired codes are dropped.
	getInvite(ctx context.Context, code string) (hostTag string, expires time.Time, ok bool, err error)
	delInvite(ctx context.Context, code string) error
	putSave(ctx context.Context, name, tag string, body []byte) error
	// latestSave is the most recent blob by name; ok false when there is none.
	latestSave(ctx context.Context) ([]byte, bool, error)
	saveNames(ctx context.Context) ([]string, error)
	// delSaves removes the blobs whose name contains tag and returns how many.
	delSaves(ctx context.Context, tag string) (int, error)
}

type pgStore struct{ pool *pgxpool.Pool }

func openStore(ctx context.Context, url string) (*pgStore, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, errors.New("HELLOKITTY_DATABASE_URL: invalid")
	}
	cfg.MaxConns = 3
	var pool *pgxpool.Pool
	for i := 0; ; i++ { // Postgres may come up after this unit
		pool, err = pgxpool.NewWithConfig(ctx, cfg)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				break
			}
			pool.Close()
		}
		if i >= 30 {
			return nil, fmt.Errorf("database unreachable: %w", err)
		}
		time.Sleep(2 * time.Second)
	}
	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("schema: %w", err)
	}
	return &pgStore{pool}, nil
}

func (p *pgStore) putInvite(ctx context.Context, code, hostTag string, expires time.Time) (bool, error) {
	// An expired row under the same code is replaced; a live one makes the code taken.
	tag, err := p.pool.Exec(ctx, `INSERT INTO hk_invites (code, host_tag, expires_at) VALUES ($1,$2,$3)
		ON CONFLICT (code) DO UPDATE SET host_tag = EXCLUDED.host_tag, expires_at = EXCLUDED.expires_at
		WHERE hk_invites.expires_at <= now()`, code, hostTag, expires)
	if err != nil {
		return false, err
	}
	// Opportunistic sweep of long-expired codes.
	_, _ = p.pool.Exec(ctx, `DELETE FROM hk_invites WHERE expires_at < now() - interval '1 hour'`)
	return tag.RowsAffected() == 1, nil
}

func (p *pgStore) getInvite(ctx context.Context, code string) (string, time.Time, bool, error) {
	var host string
	var exp time.Time
	err := p.pool.QueryRow(ctx, `SELECT host_tag, expires_at FROM hk_invites WHERE code = $1`, code).Scan(&host, &exp)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", time.Time{}, false, nil
	}
	if err != nil {
		return "", time.Time{}, false, err
	}
	if time.Now().After(exp) {
		_ = p.delInvite(ctx, code)
		return "", time.Time{}, false, nil
	}
	return host, exp, true, nil
}

func (p *pgStore) delInvite(ctx context.Context, code string) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM hk_invites WHERE code = $1`, code)
	return err
}

func (p *pgStore) putSave(ctx context.Context, name, tag string, body []byte) error {
	_, err := p.pool.Exec(ctx, `INSERT INTO hk_saves (name, tag, body) VALUES ($1,$2,$3)
		ON CONFLICT (name) DO UPDATE SET body = EXCLUDED.body`, name, tag, body)
	return err
}

func (p *pgStore) latestSave(ctx context.Context) ([]byte, bool, error) {
	var body []byte
	err := p.pool.QueryRow(ctx, `SELECT body FROM hk_saves ORDER BY name DESC LIMIT 1`).Scan(&body)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	return body, err == nil, err
}

func (p *pgStore) saveNames(ctx context.Context) ([]string, error) {
	rows, err := p.pool.Query(ctx, `SELECT name FROM hk_saves ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (p *pgStore) delSaves(ctx context.Context, tag string) (int, error) {
	var t pgconn.CommandTag
	t, err := p.pool.Exec(ctx, `DELETE FROM hk_saves WHERE position($1 in name) > 0`, tag)
	return int(t.RowsAffected()), err
}
