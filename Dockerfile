# Cross-compiling builder: runs natively on the build machine but emits a
# binary for the *target* platform. Deployed via `docker compose build` run
# directly on the Oracle arm64 VM, so TARGETARCH is auto-populated as arm64
# there with no flags needed; --platform still lets you cross-build for a
# different target (e.g. testing an amd64 image on a Mac) when needed.
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Dependencies first so the module layer is cached across source-only changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
# CGO off => fully static binary, so the runner stage needs no libc.
# -trimpath strips local filesystem paths out of the binary.
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/medilog-api . \
 && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.20 AS runner

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /out/medilog-api ./medilog-api

# cmd/migrate resolves its source as "file://migrations", relative to the
# working directory — it does NOT use the embedded FS (that is only how the
# DB_VERIFY_SCHEMA drift check learns the expected version). So the SQL files
# have to be in the image next to the binary, or `migrate up` finds nothing.
# Shipping both in one image lets the `migrate` and `api` compose services
# share it: the migrate service overrides the entrypoint with ./migrate.
COPY --from=builder /out/migrate ./migrate
COPY migrations ./migrations

RUN adduser --disabled-password --gecos "" --uid 10001 appuser
USER appuser

# 8080 is the port Compose publishes (127.0.0.1:8080 -> Caddy proxies to it).
ENV PORT=8080
EXPOSE 8080

# Used by Compose's own healthcheck and any plain `docker run`.
#
# -O /dev/null rather than --spider: --spider issues a HEAD, and the route is
# registered with chi's Get(), so HEAD /health answers 405 and the container
# never leaves "starting".
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 -O /dev/null "http://127.0.0.1:${PORT}/health" || exit 1

ENTRYPOINT ["./medilog-api"]
