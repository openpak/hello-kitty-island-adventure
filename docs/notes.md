# Notes: hello-kitty-island-adventure

## Status

Not deployed and not yet tested against the game. The host and path facts
in [protocol.md](protocol.md) were read from the Switch executable's
managed metadata; every response shape is provisional until a capture of a
real session confirms it. The service is built to make that capture easy:
unknown routes log their body and answer `{}`, uploads are kept under
`-data`, and `-settings` serves a captured `servers/settings` verbatim.

## Why this shape

- One process answers all four first-party deployments because the
  path-to-host mapping is unknown; splitting comes after evidence item 3.
- Invite codes live in memory with a TTL because the observed shape is a
  minutes-long rendezvous, not accounts or lobbies. Nothing is persisted
  that the protocol does not prove is needed.
- Save blobs are kept whole rather than parsed: the format is unknown and
  a save is the player's data, so get serves back exactly what upload
  received.
- Photon (`openpak/photon`) and UGS (`openpak/ugs`) are reused, not
  reimplemented: the client already speaks their protocols, and both
  services exist. This repo only reports their reachability in `/healthz`
  until resolve answers must carry room/relay pointers.

## Relation to sunblink/

`platforms/switch/games/sunblink` is the doc-only study this server grew
out of. See [sunblink-promotion.md](sunblink-promotion.md) for the plan to
promote it once this repo is live.
