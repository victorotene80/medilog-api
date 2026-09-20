.PHONY: test test-unit test-integration test-coverage test-build lint vet migrate-test \
        docker-build gcp-deploy gcp-logs

MIGRATE_TEST_CONTAINER ?= medilog-migrate-test
MIGRATE_TEST_PORT      ?= 55432
MIGRATE_TEST_URL       ?= postgres://test:test@localhost:$(MIGRATE_TEST_PORT)/medilog_migrate_test?sslmode=disable

test-build:
	@echo "=== Verifying test build ==="
	go build ./...

test:
	@echo "=== Running unit tests (skipping integration) ==="
	go test -short -count=1 -race ./internal/...

test-verbose:
	@echo "=== Running unit tests (verbose) ==="
	go test -short -count=1 -race -v ./internal/...

test-integration:
	@echo "=== Running integration tests ==="
	go test -count=1 -race ./internal/... -run "Integration"

test-coverage:
	@echo "=== Running tests with coverage ==="
	go test -short -count=1 -race -coverprofile=coverage.out ./internal/...
	go tool cover -func=coverage.out
	@echo "=== Coverage summary ==="
	@go tool cover -func=coverage.out | tail -1

test-coverage-html:
	@echo "=== Generating HTML coverage report ==="
	go tool cover -html=coverage.out -o coverage.html

test-mocks:
	@echo "=== Verifying mock build ==="
	go build ./test/testutil/...

# Applies every migration to a throwaway database, rolls the whole stack back
# down, then re-applies it — exercising every up and every down file. The
# integration suite runs against an already-migrated database with
# DB_AUTO_MIGRATE=false, so a broken migration file is otherwise invisible to
# CI, which is how 000002 shipped adding a column 000001 already declared.
#
# Note `migrate down` here rolls back ALL migrations (cmd/migrate calls m.Down()
# and takes no count argument), which is exactly what this target wants.
migrate-test:
	@echo "=== Verifying migrations against a scratch database ==="
	@docker rm -f $(MIGRATE_TEST_CONTAINER) >/dev/null 2>&1 || true
	@docker run -d --name $(MIGRATE_TEST_CONTAINER) \
		-e POSTGRES_PASSWORD=test -e POSTGRES_USER=test -e POSTGRES_DB=medilog_migrate_test \
		-p $(MIGRATE_TEST_PORT):5432 postgres:17-alpine >/dev/null
	@echo "waiting for postgres..."
	@for i in $$(seq 1 30); do \
		docker exec $(MIGRATE_TEST_CONTAINER) pg_isready -U test -d medilog_migrate_test >/dev/null 2>&1 && break; \
		sleep 1; \
	done
	@set -e; \
	trap 'docker rm -f $(MIGRATE_TEST_CONTAINER) >/dev/null 2>&1 || true' EXIT; \
	DATABASE_URL="$(MIGRATE_TEST_URL)" go run ./cmd/migrate up; \
	DATABASE_URL="$(MIGRATE_TEST_URL)" go run ./cmd/migrate down; \
	DATABASE_URL="$(MIGRATE_TEST_URL)" go run ./cmd/migrate up; \
	DATABASE_URL="$(MIGRATE_TEST_URL)" go run ./cmd/migrate status
	@echo "=== Migrations OK ==="

# Regenerates docs/swagger.{json,yaml} and docs/docs.go from the @Router
# annotations on the handlers. The generated spec is the contract the mobile
# team codes against, and it had drifted badly enough that whole route groups
# (the notification endpoints) appeared not to exist at all — so regenerate it
# whenever a route or a request/response struct changes.
# swag is a build tool, not an import. It is deliberately not a go.mod
# dependency — that would pull urfave/cli and sigs.k8s.io/yaml into the module
# graph purely for doc generation — and a globally installed binary is neither
# guaranteed to be on PATH nor guaranteed to match the swaggo/swag version the
# code is built against. Running the pinned version directly fixes both: no
# setup step, same output on every machine and in CI.
SWAG := go run github.com/swaggo/swag/cmd/swag@v1.16.6

swagger:
	@echo "=== Regenerating swagger spec ==="
	$(SWAG) init -g main.go --parseDependency --parseInternal -o docs
	@echo "=== Swagger regenerated: docs/swagger.json ==="

# Fails when the committed spec is out of date with the annotations.
swagger-check: swagger
	@git diff --quiet -- docs/swagger.json docs/swagger.yaml docs/docs.go \
		|| { echo "ERROR: swagger spec is stale — commit the regenerated docs/"; exit 1; }
	@echo "=== Swagger spec is up to date ==="

lint:
	@echo "=== Running linter ==="
	golangci-lint run ./...

vet:
	@echo "=== Running go vet ==="
	go vet ./...

clean:
	@echo "=== Cleaning coverage files ==="
	rm -f coverage.out coverage.html

help:
	@echo "Usage:"
	@echo "  make test-build         Verify test infrastructure compiles"
	@echo "  make test               Run unit tests (skips integration)"
	@echo "  make test-verbose       Run unit tests with verbose output"
	@echo "  make test-integration   Run integration tests (requires TEST_DB_URL)"
	@echo "  make test-coverage      Run tests with coverage report"
	@echo "  make test-coverage-html Generate HTML coverage report"
	@echo "  make test-mocks         Verify mock build compiles"
	@echo "  make migrate-test       Apply all migrations to a scratch DB (needs docker)"
	@echo "  make swagger            Regenerate the OpenAPI spec from handler annotations"
	@echo "  make swagger-check      Fail if the committed spec is stale (for CI)"
	@echo "  make lint               Run golangci-lint"
	@echo "  make vet                Run go vet"
	@echo "  make clean              Remove coverage files"

# ---------------------------------------------------------------------------
# Container / Google Cloud
# ---------------------------------------------------------------------------

GCP_REGION   ?= us-central1
GCP_SERVICE  ?= medilog-api
TAG          ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo latest)

# Cloud Run is amd64-only, so --platform is not optional on an Apple Silicon
# machine: without it you get an arm64 image that Cloud Run refuses to start.
docker-build:
	@echo "=== Building linux/amd64 image ==="
	docker build --platform linux/amd64 -t $(GCP_SERVICE):$(TAG) .

# One pipeline: build -> push -> `migrate up` as a Cloud Run job -> deploy.
# _CLOUDSQL_INSTANCE and _REDIS_HOST have no sensible defaults, so they are
# required here rather than silently wrong.
gcp-deploy:
	@test -n "$(CLOUDSQL_INSTANCE)" || { echo "ERROR: set CLOUDSQL_INSTANCE=PROJECT:REGION:INSTANCE"; exit 1; }
	@test -n "$(REDIS_HOST)" || { echo "ERROR: set REDIS_HOST=<memorystore private ip>"; exit 1; }
	@echo "=== Submitting build $(TAG) to Cloud Build ==="
	gcloud builds submit --config cloudbuild.yaml \
		--substitutions=_TAG=$(TAG),_REGION=$(GCP_REGION),_SERVICE=$(GCP_SERVICE),_CLOUDSQL_INSTANCE=$(CLOUDSQL_INSTANCE),_REDIS_HOST=$(REDIS_HOST)

gcp-logs:
	gcloud run services logs tail $(GCP_SERVICE) --region $(GCP_REGION)
