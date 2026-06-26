.PHONY: all build backend backend-ffi frontend dev dev-backend dev-frontend run run-ffi test lint clean tidy docker \
        release release-amd64 release-ppc64le release-ppc64le-ffi release-icr release-icr-ffi \
        verify-arm verify-ppc64le check-dnp3-arm check-arm-toolchain \
        check-dnp3-ppc64le check-ppc64le-toolchain image image-ffi image-save

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

# ARM/v7 cross (Advantech ICR-3232). The armv7 opendnp3 is vendored WITHOUT TLS,
# so the link omits -lssl/-lcrypto; libstdc++ is pulled statically (-l:libstdc++.a)
# so the fully-static binary needs no C++ runtime on the device.
DNP3_ARM_TRIPLE   ?= armv7-unknown-linux-gnueabihf
DNP3_ARM_DIR      := $(abspath $(GODNP3_DIR))/third_party/opendnp3/$(DNP3_ARM_TRIPLE)
DNP3_ARM_CXXFLAGS := -std=c++17 -I$(DNP3_ARM_DIR)/include
DNP3_ARM_LDFLAGS  := -L$(DNP3_ARM_DIR)/lib -lopendnp3 -l:libstdc++.a -lpthread -lm -ldl -static-libgcc

# ppc64le cross (IBM POWER container). Same no-TLS, static-libstdc++ link as ARM.
DNP3_PPC64LE_TRIPLE   ?= powerpc64le-unknown-linux-gnu
DNP3_PPC64LE_DIR      := $(abspath $(GODNP3_DIR))/third_party/opendnp3/$(DNP3_PPC64LE_TRIPLE)
DNP3_PPC64LE_CXXFLAGS := -std=c++17 -I$(DNP3_PPC64LE_DIR)/include
DNP3_PPC64LE_LDFLAGS  := -L$(DNP3_PPC64LE_DIR)/lib -lopendnp3 -l:libstdc++.a -lpthread -lm -ldl -static-libgcc

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

# ── ICR-3232 (Advantech, linux/arm/v7) ────────────────────────────────────────
# Two flavors mirror the edge gateway: a pure-Go stub (IEC-104 master only) and a
# fully static cgo build with the real opendnp3 DNP3 master.

# Pure-Go stub: IEC-104 master is fully functional; DNP3 lines use the goDnp3 stub
# (they connect to nothing). CGO off → statically linked automatically.
release-icr: frontend
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
	  go build -trimpath -ldflags="$(LDFLAGS)" \
	  -o bin/go104-linux-armv7 ./cmd/server
	@echo "Built bin/go104-linux-armv7 (stub DNP3 — IEC-104 master only)"

# Real opendnp3 DNP3 master, FULLY static-linked for the ICR-3232. cgo would
# otherwise link the build-host glibc (Debian GLIBC_2.38), far newer than the ICR
# firmware userland — a dynamic binary dies with "GLIBC_2.xx not found". The
# -extldflags -static bakes glibc in; netgo gives a pure-Go DNS resolver so name
# lookups don't need glibc NSS (configure outstations by IP — asio's getaddrinfo
# still wants NSS otherwise). modernc.org/sqlite stays pure-Go.
release-icr-ffi: check-dnp3-arm check-arm-toolchain frontend
	@mkdir -p bin
	CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 \
	CC=arm-linux-gnueabihf-gcc \
	CXX=arm-linux-gnueabihf-g++ \
	CGO_CXXFLAGS="$(DNP3_ARM_CXXFLAGS)" \
	CGO_LDFLAGS="$(DNP3_ARM_LDFLAGS)" \
	go build -tags dnp3_ffi,netgo -trimpath -ldflags="-s -w -extldflags '-static'" \
	  -o bin/go104-linux-armv7-ffi ./cmd/server
	@echo "Built bin/go104-linux-armv7-ffi (real opendnp3 DNP3 master, static)"
	@echo "Deploy: scp bin/go104-linux-armv7-ffi root@<icr-ip>:/root/  (no .so needed — opendnp3 is static-linked)"

# Assert the cross-built FFI binary is a static ARM ELF before shipping.
verify-arm:
	@file bin/go104-linux-armv7-ffi 2>/dev/null | grep -q "ARM" || { echo "ERROR: bin/go104-linux-armv7-ffi missing or not ARM — run 'make release-icr-ffi' first"; exit 1; }
	@file bin/go104-linux-armv7-ffi | grep -q "statically linked" || echo "WARN: not statically linked — check CGO/toolchain"
	@file bin/go104-linux-armv7-ffi; ls -lh bin/go104-linux-armv7-ffi

# Linux ppc64le with REAL opendnp3, fully static — for the FROM-scratch POWER
# container. Same rationale as the ICR FFI build: static so the scratch image
# carries no glibc/libstdc++; netgo gives a pure-Go DNS resolver.
release-ppc64le-ffi: check-dnp3-ppc64le check-ppc64le-toolchain frontend
	@mkdir -p bin
	CGO_ENABLED=1 GOOS=linux GOARCH=ppc64le \
	CC=powerpc64le-linux-gnu-gcc \
	CXX=powerpc64le-linux-gnu-g++ \
	CGO_CXXFLAGS="$(DNP3_PPC64LE_CXXFLAGS)" \
	CGO_LDFLAGS="$(DNP3_PPC64LE_LDFLAGS)" \
	go build -tags dnp3_ffi,netgo -trimpath -ldflags="-s -w -extldflags '-static'" \
	  -o bin/go104-linux-ppc64le-ffi ./cmd/server
	@echo "Built bin/go104-linux-ppc64le-ffi (real opendnp3 DNP3 master, static)"

verify-ppc64le:
	@file bin/go104-linux-ppc64le-ffi 2>/dev/null | grep -q "PowerPC" || { echo "ERROR: bin/go104-linux-ppc64le-ffi missing or not ppc64le — run 'make release-ppc64le-ffi' first"; exit 1; }
	@file bin/go104-linux-ppc64le-ffi | grep -q "statically linked" || echo "WARN: not statically linked — check CGO/toolchain"
	@file bin/go104-linux-ppc64le-ffi; ls -lh bin/go104-linux-ppc64le-ffi

# ── Cross-build preflight checks ──────────────────────────────────────────────
check-dnp3-arm:
	@if [ ! -f $(DNP3_ARM_DIR)/include/opendnp3/DNP3Manager.h ] || [ ! -f $(DNP3_ARM_DIR)/lib/libopendnp3.a ]; then \
		echo "ERROR: missing $(DNP3_ARM_DIR)/{include/opendnp3/DNP3Manager.h,lib/libopendnp3.a}"; \
		echo "  Vendor it: cd $(GODNP3_DIR) && make opendnp3-vendor-arm"; \
		exit 1; \
	fi

check-arm-toolchain:
	@if ! command -v arm-linux-gnueabihf-gcc >/dev/null 2>&1; then \
		echo "ERROR: arm-linux-gnueabihf-gcc not found."; \
		echo "  Install on Debian/Ubuntu: sudo apt install gcc-arm-linux-gnueabihf g++-arm-linux-gnueabihf"; \
		exit 1; \
	fi

check-dnp3-ppc64le:
	@if [ ! -f $(DNP3_PPC64LE_DIR)/include/opendnp3/DNP3Manager.h ] || [ ! -f $(DNP3_PPC64LE_DIR)/lib/libopendnp3.a ]; then \
		echo "ERROR: missing $(DNP3_PPC64LE_DIR)/{include/opendnp3/DNP3Manager.h,lib/libopendnp3.a}"; \
		echo "  Vendor it: cd $(GODNP3_DIR) && make opendnp3-vendor-ppc64le"; \
		exit 1; \
	fi

check-ppc64le-toolchain:
	@if ! command -v powerpc64le-linux-gnu-g++ >/dev/null 2>&1; then \
		echo "ERROR: powerpc64le-linux-gnu-g++ not found."; \
		echo "  Install on Debian/Ubuntu: sudo apt install gcc-powerpc64le-linux-gnu g++-powerpc64le-linux-gnu"; \
		exit 1; \
	fi

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

# Same scratch container, but with the REAL DNP3 master (static FFI binary) so
# the POWER deployment can poll DNP3 outstations, not just IEC-104.
image-ffi: release-ppc64le-ffi
	docker build --platform linux/ppc64le -f Dockerfile.ffi -t $(IMAGE) .

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
