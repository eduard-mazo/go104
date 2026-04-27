// Package ui embeds the compiled Vue frontend.
// The dist/ directory is populated by `make frontend` (or `make build`).
package ui

import "embed"

//go:embed all:dist
var FS embed.FS
