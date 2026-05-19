.PHONY: all build backend frontend dev dev-backend dev-frontend run test lint clean tidy docker \
        release release-amd64 release-ppc64le image image-save

BINARY   := ./bin/go104
WEB_DIR  := ./web

# Container image
IMAGE    ?= localhost/go104:ppc64le
TARBALL  ?= go104-ppc64le.tar

# Default: build for the host architecture
all: build

# ── Full build ────────────────────────────────────────────────────────────────
# frontend must run first so the embed picks up the latest dist

build: frontend backend

frontend: $(WEB_DIR)/node_modules
	cd $(WEB_DIR) && pnpm run build
	@echo "Copying dist → internal/ui/dist"
	rm -rf ./internal/ui/dist
	cp -r $(WEB_DIR)/dist ./internal/ui/dist

# install only when package.json is newer than node_modules
$(WEB_DIR)/node_modules: $(WEB_DIR)/package.json
	cd $(WEB_DIR) && pnpm install
	@touch $(WEB_DIR)/node_modules  # update mtime so make skips re-install

backend:
	@mkdir -p bin
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/server

# ── Cross-compilation release targets ─────────────────────────────────────────
# These targets produce statically-linked, stripped binaries for Linux.
# The frontend is compiled once (architecture-independent) and embedded by the
# Go linker into each target binary.
#
# Prerequisites: standard Go toolchain (no CGo — modernc.org/sqlite is pure Go).
#
# Usage:
#   make release              # builds both targets
#   make release-amd64        # Red Hat / x86_64 (RHEL 7+, Rocky, AlmaLinux)
#   make release-ppc64le      # IBM POWER little-endian (RHEL for POWER)

LDFLAGS := -s -w

release: frontend release-amd64 release-ppc64le
	@echo ""
	@echo "Release binaries:"
	@ls -lh bin/go104-linux-*

# Linux x86_64 — Red Hat family (RHEL 7+, Rocky Linux, AlmaLinux, CentOS Stream)
release-amd64: frontend
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 \
	  go build -ldflags="$(LDFLAGS)" \
	  -o bin/go104-linux-amd64 ./cmd/server
	@echo "Built bin/go104-linux-amd64"

# Linux ppc64le — IBM POWER little-endian (RHEL for POWER / OpenPOWER)
release-ppc64le: frontend
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=ppc64le \
	  go build -ldflags="$(LDFLAGS)" \
	  -o bin/go104-linux-ppc64le ./cmd/server
	@echo "Built bin/go104-linux-ppc64le"

# ── Container (ppc64le, air-gapped) ─────────────────────────────────────────

# Build the ppc64le container image (requires Docker + buildx on the build host).
# Steps: build Vue frontend → cross-compile Go binary → docker build.
# FROM scratch + pre-built binary: no emulation needed, --platform sets metadata only.
image: frontend
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=ppc64le \
		go build -trimpath -ldflags="-s -w" \
		-o bin/go104-linux-ppc64le ./cmd/server
	docker build --platform linux/ppc64le -t $(IMAGE) .

# Save the image to a tar for transfer to the air-gapped RHEL target host.
#   (copy via USB or other offline media, then on the target host:)
#   podman load -i go104-ppc64le.tar
image-save: image
	docker save $(IMAGE) -o $(TARBALL)
	@echo "Saved $(IMAGE) → $(TARBALL)"

# ── Development helpers ───────────────────────────────────────────────────────

dev-backend:
	go run ./cmd/server

dev-frontend:
	cd $(WEB_DIR) && pnpm run dev

# Run both dev servers (requires tmux or two terminals)
dev:
	@echo "Start dev-backend in one terminal: make dev-backend"
	@echo "Start dev-frontend in another:     make dev-frontend"

PORT ?= 8080
run: build
	HTTP_PORT=$(PORT) $(BINARY)

# ── Maintenance ───────────────────────────────────────────────────────────────

tidy:
	go mod tidy

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ $(WEB_DIR)/dist/ internal/ui/dist/

docker:
	docker build -t go104:latest .

web-install:
	cd $(WEB_DIR) && pnpm install
