# minsc.sunblink.io + AWS API Gateway — the client's contract

Read from the Switch executable's managed metadata (Hello Kitty Island
Adventure `010027901C89C000`, base v0 US, Unity 2021.3.43f1 IL2CPP), see
[provenance](provenance.md). Keep this file in step with `cmd/hellokittyd`.

Wire facts come only from the RECON capture notes
(`scratch/decompile-hello-kitty-island-adventure/RECON.md`,
`scratch/decompile-hello-kitty-island-adventure/findings.json`,
`scratch/decompile-hello-kitty-island-adventure/strings.txt`).
Everything else — request/response shapes, methods, auth headers — is
**provisional** until a live capture confirms it, and is marked so below.

## Base URLs

Four first-party deployments, all HTTPS, each with its API-Gateway stage:

- `https://61cg23vc38.execute-api.us-east-1.amazonaws.com/live`
- `https://gkbkwtb0ad.execute-api.us-east-1.amazonaws.com/prod`
- `https://shlax276jb.execute-api.us-east-1.amazonaws.com/prod`
- `https://minsc.sunblink.io/prod`

Which deployment serves which path family is **not known** (RECON next
step 3): the service routes all four identically and strips one leading
`/live` or `/prod` segment before matching.

## Requests

Measured fact: the path families below exist in managed metadata
(`Sunblink.Networking`, `InviteCodeService`, `Savegamesservice`,
`SupportService`, `Switch*` assemblies). Method, query, headers and bodies
are provisional.

| Request | Client reads | hellokittyd answers |
| --- | --- | --- |
| `servers/settings` | boot config: presumably the Photon AppId/region and Relay config (no hardcoded AppId or `*.exitgamescloud.com` host exists in the build) | built-in JSON with `photonAppId`/`photonRegion` from `HKIA_PHOTON_APPID`/`HKIA_PHOTON_REGION`, or `-settings` file verbatim |
| `invitecode/allocate` | a short code for the host's island | fresh 6-char code + `expiresIn` seconds (TTL `-code-ttl`, default 5m) |
| `invitecode/resolve` | where to connect for a guest's code | `found` + host tag + `expiresIn`, or 404 `{"found":false}` |
| `invitecode/dispose` | success only | `200 {}`, code forgotten |
| `savegames/upload` | success only | `200 {}`; the raw body is kept under `-data/saves/` |
| `savegames/get` | the caller's save | most recent kept blob, or `{}` when none |
| `savegames/query` | the caller's save list | `{"saves":[...]}` of kept blob names |
| `savegames/delete` | success only | `200`; the caller's blobs are removed |
| `support/claim-save-data`, `support/offer-save-data`, `support/send-device-logs` | unknown | `200 {}`, logged |
| `session-token/...` | unknown (token flow; `AccessToken`/`RefreshToken` types exist) | `200 {}`, logged |
| anything else | unknown | `200 {}`, logged with body so a capture shows what the client asks next |

Every query value is treated as identifying or authenticating and is never
logged raw (per-process tags only). No Nintendo dauth/NSA token strings
appear in managed metadata: auth is Sunblink account + tokens.

## What is not here

- Sessions: Unity Netcode over Photon Realtime rooms
  (`PhotonHostRpcTransport`, `LoadBalancingClient`), served by
  `platforms/switch/games/photon`. The AppId is runtime-supplied, still
  unrecovered (RECON next step 4).
- Alternate transport + auth: Unity Relay, player auth, QoS discovery and
  social (`relay-allocations`, `player-auth`, `qos-discovery`, `social`
  under `services.api.unity.com`, UGS project id
  `bd29838b372f4b7c9e4b846e1fb8a096`), served by
  `platforms/switch/games/ugs`. `PhotonGameSessionInviteCodeData` and
  `RelayGameSessionInviteCodeData` both exist, so both transports are live
  options in the client.
- Content/news: `lynx-live.s3.amazonaws.com/.../devinbox/v3`,
  `.../gamenews/v2` (+ staging variants), mirrored at
  `d2gx77gza65ci6.cloudfront.net`. Static JSON, not served yet.
- Not NEX/NPLN/Pia (only nnSdk-static-lib symbols, no tenants), not
  PlayFab (two bare tokens, no API types), not EOS/Demonware.

The game is 2-player visit-a-friend: the host allocates a code in the
in-game Multiplayer app, the guest resolves it. Same platform, same
version, minutes-long code expiry. Nothing here simulates the game.
