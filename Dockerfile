# syntax=docker/dockerfile:1
# Hello Kitty Island Adventure's first-party backend (Sunblink REST on AWS API
# Gateway + minsc.sunblink.io) for the Switch build. Plain HTTP on 23600
# behind the TLS terminator; health on 23601.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOWORK=off go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/hellokittyd ./cmd/hellokittyd

FROM gcr.io/distroless/static
COPY --from=build /out/hellokittyd /hellokittyd
USER nonroot:nonroot
EXPOSE 23600 23601
ENTRYPOINT ["/hellokittyd", "-data", "/data"]
