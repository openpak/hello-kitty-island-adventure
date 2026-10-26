# sunblink/ promotion plan

`platforms/switch/games/sunblink` (repo `openpak/sunblink`) is the doc-only
study for this exact title: hypothesis, design, and evidence-needed steps
written before the dump existed. This repo (`openpak/hello-kitty-island-adventure`)
is the live server the RECON scoped. Two repos must not own one title, so
once this server is deployed and serving a real session, `sunblink/` is
promoted out of the way. Nothing below touches `sunblink/` itself; it is a
plan for a later serial step.

## What moves where

| sunblink/ content | Disposition |
| --- | --- |
| `docs/design.md` §§1–2 (game facts, wiki-observed multiplayer shape) | Fold the still-true lines into this repo's `docs/notes.md`; the wiki shape is already cited in `docs/provenance.md` |
| `docs/design.md` §§3–5, `docs/evidence-needed.md` (hypothesis + capture steps) | Superseded by this repo's `docs/evidence.md` + RECON; archive, do not copy |
| `docs/protocol.md`, `docs/evidence.md` ("Nothing recorded yet") | Delete with the repo; this repo's versions replace them |
| `docs/deploy.md`, `docs/notes.md`, `prds/`, `next-session.md` | Read once for anything not yet captured here, then archive |
| Kill conditions (NEX premise wrong → delete) | Retired: the RECON answered them — not NEX, Sunblink+Photon+UGS confirmed |

## Steps

1. Deploy this repo and confirm one real session against it (evidence item
   1 in this repo's `docs/evidence.md`).
2. Fold any surviving `sunblink/docs/design.md` game facts into this
   repo's docs; link this repo from the titles page entry (`t-72360`)
   instead of `sunblink`.
3. Archive `openpak/sunblink`: final commit pointing at
   `openpak/hello-kitty-island-adventure`, then archive the GitHub repo.
   Do not delete history — the study is the audit trail for the design.
4. Remove the local `platforms/switch/games/sunblink/` checkout only after
   the archive commit is pushed.

## Why archive instead of reuse

The study's name is the vendor, the server's name is the title — matching
the per-title layout of the other game repos (`unrailed`,
`stumble-guys`). A second Sunblink title would get its own repo and share
nothing here except lessons. Keeping `sunblink` as a live name would
invite scope creep into a vendor umbrella the workspace does not use.
