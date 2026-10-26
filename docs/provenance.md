# Provenance

Where each fact in [protocol.md](protocol.md) came from, per the
[clean-room policy](https://docs.openpak.org/contributors/clean-room/). Only the owner's own game
client was examined. No server code of the developer was seen or used.

| Fact | Source |
| --- | --- |
| Title `010027901C89C000`, base v0 US NSP; Program NCA SDK 18.3.3.0 | `scratch/decompile-hello-kitty-island-adventure/RECON.md` ("Format / layout"), `findings.json` (`dump`, `format`) |
| Unity 2021.3.43f1 IL2CPP; `global-metadata.dat` 23 MB | `RECON.md` ("Format / layout"), `findings.json` (`engine`), `strings.txt` (`SDK MW+UnityTechnologies+Unity-2021_3_43_f1`) |
| Four first-party base URLs + `/live`, `/prod` stages | `findings.json` (`backend_hosts`), `RECON.md` ("Backend hosts") |
| REST paths `invitecode/allocate\|dispose\|resolve`, `savegames/delete\|get\|query\|upload`, `servers/settings`, `support/claim-save-data\|offer-save-data\|send-device-logs`, `session-token/...` | `findings.json` (`network.rest_paths`), `RECON.md` ("Middleware verdict", "Session model") |
| Sunblink assemblies (`Sunblink.Networking`, `InviteCodeService`, `Savegamesservice`, `SupportService`) | `RECON.md` ("Middleware verdict") |
| Photon Realtime integrated; no hardcoded AppId; `PhotonHostRpcTransport`, `wss://{0}:{1}` | `findings.json` (`network.photon*`), `RECON.md` ("Middleware verdict") |
| Unity Netcode linked (`Unity.NetCode.dll`) | `findings.json` (`network.unity_netcode`), `strings.txt` |
| UGS DLLs + hosts (`player-auth`, `relay-allocations`, `qos-discovery`, `social`, `userreporting`); project id `bd29838b…` | `findings.json` (`backend_hosts`, `ugs_project_id`), `strings.txt` |
| S3/CloudFront content paths (`devinbox/v3`, `gamenews/v2` + staging) | `findings.json` (`backend_hosts`), `RECON.md` ("Backend hosts") |
| Not NEX/NPLN (nnSdk symbols only), not PlayFab (bare tokens), not EOS/Demonware | `findings.json` (`network.nex_*`, `playfab*`, `eos`), `RECON.md` ("Middleware verdict") |
| Auth is Sunblink account + tokens (`AccessToken`/`RefreshToken`); no Nintendo dauth strings | `RECON.md` ("Backend hosts") |
| Invite-code session model; Unity Relay alternate (`RelayGameSessionInviteCodeData`) | `RECON.md` ("Session model"), `findings.json` (`network.transport_model`) |
| OS services incl. `ldn:u` local wireless | `findings.json` (`npdm_services`), `RECON.md` ("OS service surface") |
| 2-player visit-a-friend, minutes-long codes, same platform/version | public wiki via `platforms/switch/games/sunblink/docs/design.md` §2 (leads only, not wire facts) |
