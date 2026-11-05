# Account erasure (card openpak/prds#120)

This service has no erase endpoint, on purpose: what it keeps cannot be tied to an OpenPak account.

## What is kept

| Table | Columns | Key |
| --- | --- | --- |
| `hk_saves` | `name`, `tag`, `body`, `created_at` | `name` = timestamp + `tag` |
| `hk_invites` | `code`, `host_tag`, `expires_at` | `code` |

`tag` / `host_tag` is a keyed hash (HMAC-SHA256, truncated to 3 bytes) of the opaque
`Authorization` bearer string the client sends to Sunblink's endpoints. Three facts make it
unattributable:

1. The bearer is a Sunblink account token. The service never validates it, never calls nx-baas or
   the account core, and never sees a BAAS user id, NSA id or console PID (no Nintendo identity
   appears in the traffic, see `protocol.md`).
2. The hash key is random per process (`tagKey`), so the same player gets a different `tag` after
   every restart. There is no stable value to look up from outside, not even for the service itself.
3. `savegames/get` and `savegames/query` do not filter by caller: they serve the most recent blob
   and all names. The self-delete (`savegames/delete`) only matches blobs whose `name` contains the
   current process's tag for that bearer.

So no mapping from an OpenPak account to these rows exists, and none is invented here.

## Retention

- `hk_invites`: a code lives for the invite TTL (minutes); rows expired for over an hour are purged
  on the next allocation. No personal data beyond the per-process host tag.
- `hk_saves`: kept until the player deletes them in game (self-delete, same process run only).
  These are game save blobs without any account identifier.

## If this changes

Erasure becomes possible once the title's token flow is captured and the service resolves the
bearer to a Nintendo or OpenPak identity through nx-baas (`/internal/switch/identity`), storing that
BAAS user id with each row. At that point add the standard
`DELETE /internal/accounts/{account_id}/data` route (see plants-vs-zombies `cmd/pvzd/erase.go`).
