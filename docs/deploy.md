# Deploy: hello-kitty-island-adventure

## Running it

```
go run ./cmd/hellokittyd                                  # plain HTTP on :23560, health on :23561
go run ./cmd/hellokittyd -settings ./settings.json -data ./data
```

| Flag / env | Meaning |
| --- | --- |
| `-data` | directory for kept uploads and save blobs (default `./data`) |
| `-settings` | a `servers/settings` JSON to serve instead of the built-in one |
| `-code-ttl` | invite-code lifetime (default `5m`) |
| `HKIA_PHOTON_APPID` / `HKIA_PHOTON_REGION` | Photon pointers in the built-in settings |
| `PHOTON_URL` | Photon service base for the health report (default `http://127.0.0.1:8080`) |
| `UGS_URL` | UGS service base for the health report (default `http://127.0.0.1:22681`) |

Ports are in [ports.md](ports.md). Logs never carry a token, code, save id
or account: query values and bodies are tagged per process.

The four first-party hostnames must resolve to this service with TLS
terminated in front (Traefik); the service only speaks plain HTTP.

## Container

The Dockerfile builds a static `hellokittyd` (`CGO_ENABLED=0`, `-trimpath`) into a
`gcr.io/distroless/static` image that runs as `nonroot:nonroot`, with `EXPOSE 23560 23561`
and `-data /data`.
