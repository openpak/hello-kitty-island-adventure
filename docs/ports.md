# Ports

| Port | Use |
| --- | --- |
| 23560 | plain HTTP for the four Sunblink first-party hostnames, routed by `Host` (behind the TLS terminator) |
| 23561 | `/healthz` (and `/health`), incl. Photon/UGS reachability |
| 23562–23569 | reserved |

23560–23569 continues the 23500–23999 range after unrailed (23500–23509)
and stumble-guys (23550–23559).

Sessions use the photon server's ports and UGS uses the ugs server's
ports; this service publishes none for them.
