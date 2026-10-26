# Evidence: hello-kitty-island-adventure

## Confirmed

- Static RECON of the owner's base v0 US dump only (see
  [provenance](provenance.md)). No traffic has been captured yet.

## Evidence still needed

1. A boot capture in Harbor with OpenPak DNS: resolution order of the 4
   Sunblink/AWS hosts + UGS hosts at title screen vs Multiplayer-app vs
   host-code vs guest-join, and which ports open to each.
2. TLS-stub results for the Sunblink hosts (SNI/ALPN/pinning).
3. Which `execute-api` deployment serves `invitecode/*` vs `savegames/*`
   (path-to-host mapping: traffic or `*WithHttpInfo` call-site disassembly).
4. The Photon AppId/region (runtime `servers/settings` response or
   `PhotonInviteServiceSettings` defaults).

## Test results

Nothing recorded yet.
