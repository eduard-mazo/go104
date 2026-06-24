.PHONY: all build backend backend-ffi frontend dev dev-backend dev-frontend run run-ffi test lint clean tidy docker \
        release release-amd64 release-ppc64le image image-save

BINARY   := ./bin/go104
WEB_DIR  := ./web

# ── DNP3 (opendnp3 via the shared goDnp3 module) ──────────────────────────────
# Default builds are pure-Go: DNP3 lines use the goDnp3 stub (they connect to
# nothing). `make backend-ffi` / `run-ffi` link the real opendnp3 for live DNP3
# polling — requires the goDnp3 sibling checkout with opendnp3 vendored
# (cd ../goDnp3 && make opendnp3-vendor). modernc.org/sqlite stays pure-Go.
GODNP3_DIR    ?= ../goDnp3
DNP3_TRIPLE   ?= x86_64-unknown-linux-gnu
DNP3_DIR      := $(abspath $(GODNP3_DIR))/third_party/opendnp3/$(DNP3_TRIPLE)
DNP3_CXXFLAGS := -std=c++17 -I$(DNP3_DIR)/include
DNP3_LDFLAGS  := -L$(DNP3_DIR)/lib -lopendnp3 -lssl -lcrypto -lstdc++ -lpthread -lm -ldl

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

# Backend with real DNP3 (cgo + opendnp3 via goDnp3); modernc sqlite stays pure-Go.
backend-ffi:
	@mkdir -p bin
	CGO_ENABLED=1 \
	CGO_CXXFLAGS="$(DNP3_CXXFLAGS)" \
	CGO_LDFLAGS="$(DNP3_LDFLAGS)" \
	go build -tags dnp3_ffi -ldflags="-s -w" -o $(BINARY) ./cmd/server

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

# Run the backend with real DNP3 (no frontend rebuild).
run-ffi: backend-ffi
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
