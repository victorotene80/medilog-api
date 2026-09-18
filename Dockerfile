# Cross-compiling builder: runs natively on the build machine (fast on Apple
# Silicon) but emits a binary for the *target* platform. Cloud Run only runs
# linux/amd64, so `docker build --platform linux/amd64` on a Mac must not
# produce an arm64 binary.
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
    go build -trimpath -ldflags="-s -w" -o /out/medilog-api .

FROM alpine:3.20 AS runner

RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /out/medilog-api ./medilog-api

RUN adduser --disabled-password --gecos "" --uid 10001 appuser
USER appuser

# Cloud Run injects PORT and routes traffic to it; 8080 is its default and the
# port Compose publishes.
ENV PORT=8080
EXPOSE 8080

# Ignored by Cloud Run (it uses its own startup/liveness probes) but used by
# Compose and any plain `docker run`.
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider "http://127.0.0.1:${PORT}/health" || exit 1

ENTRYPOINT ["./medilog-api"]
