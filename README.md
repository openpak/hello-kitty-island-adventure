# Hello Kitty Island Adventure

OpenPak's Hello Kitty Island Adventure Server.

| | |
| --- | --- |
| Game | [Hello Kitty Island Adventure](https://openpak.org/titles#t-72360) |
| Network | Nintendo Switch |
| Title ID | `010027901C89C000` |
| Licence | AGPL-3.0-only |

## What it serves

| Host | What | Served by |
| --- | --- | --- |
| `*.execute-api.us-east-1.amazonaws.com` (3 Sunblink deployments) | Invite codes, cloud saves, boot settings, support, session tokens | this repo (`hellokittyd`) |
| `minsc.sunblink.io` | Same first-party API | this repo (`hellokittyd`) |
| Photon Realtime | 2-player sessions (Netcode-over-Photon rooms) | [`photon`](https://github.com/openpak/photon) |
| Unity Gaming Services (`player-auth`, `relay-allocations`, `qos-discovery`, `social`) | Auth, Relay transport, QoS, social | [`ugs`](https://github.com/openpak/ugs) |

Not NEX, not PlayFab: the build carries neither (see [protocol](docs/protocol.md)).

## Run

```sh
go run ./cmd/hellokittyd
```

| Flag / env | Meaning |
| --- | --- |
| `-data` | directory for kept uploads and save blobs (default `./data`) |
| `-settings` | a `servers/settings` JSON to serve instead of the built-in one |
| `-code-ttl` | how long an allocated invite code stays resolvable (default `5m`) |
| `HKIA_PHOTON_APPID` / `HKIA_PHOTON_REGION` | Photon pointers served in the built-in settings |
| `PHOTON_URL` | Photon service base for the health report (default `http://127.0.0.1:8080`) |
| `UGS_URL` | UGS service base for the health report (default `http://127.0.0.1:22681`) |

## Docs

- [Protocol](docs/protocol.md)
- [Evidence](docs/evidence.md)
- [Deploy](docs/deploy.md)
- [Notes](docs/notes.md)
- [Ports](docs/ports.md)
- [Provenance](docs/provenance.md)
- [Sunblink promotion plan](docs/sunblink-promotion.md)

Live compatibility status: [openpak.org/titles#t-72360](https://openpak.org/titles#t-72360).

## Licence

AGPL-3.0-only. See [LICENSE](LICENSE).
