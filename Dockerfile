# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.25

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-bookworm AS deps
WORKDIR /src

ENV CGO_ENABLED=0 \
	GOFLAGS="-mod=readonly -buildvcs=false"

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
	go mod download

FROM deps AS builder
WORKDIR /src

ARG TARGETOS
ARG TARGETARCH

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
	go build -trimpath -ldflags="-s -w" -o /out/dbot .

FROM builder AS test
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go test ./...

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app

COPY --from=builder /out/dbot /app/dbot
COPY opt/config/db.yaml /app/opt/config/db.yaml

USER nonroot:nonroot
ENTRYPOINT ["/app/dbot"]
