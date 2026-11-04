# Changelog — hello-kitty-island-adventure

## v0.2.0 — 2026-11-04

- Invite codes (with their TTL) and cloud saves move to Postgres (`hk_invites`, `hk_saves`, `HELLOKITTY_DATABASE_URL`, pool of 3); the service refuses to start without it and nothing is kept in memory or in files. `-data` is now unused. Restart test against a real Postgres. (openpak/prds#85)

## v0.1.0 — 2026-11-04

- `hellokittyd`: the client's first-party Sunblink REST API on the three AWS
  API-Gateway deployments and `minsc.sunblink.io`. Invite-code rendezvous
  (allocate/resolve/dispose with TTL), file-backed cloud saves, provisional
  `servers/settings`, empty-JSON support and session-token stubs, unknown
  routes logged and answered empty for capture. Photon and UGS reachability
  in `/healthz`. Response shapes are provisional until a live capture lands.
