# Ports

| Port | Use |
| --- | --- |
| 23600 | plain HTTP for the four Sunblink first-party hostnames, routed by `Host` (behind the TLS terminator) |
| 23601 | `/healthz` (and `/health`), incl. Photon/UGS reachability |
| 23602–23609 | reserved |

23600–23609 continues the 23500–23999 range after unrailed (23500–23509),
stumble-guys (23550–23559), astroneer (23560–23569), rivals-of-aether
(23570–23579), grounded (23580–23589) and gang-beasts (23590–23599).

Sessions use the photon server's ports and UGS uses the ugs server's
ports; this service publishes none for them.
