//go:build tools
// +build tools

package tools

import (
	_ "github.com/golang-migrate/migrate/v4/cmd/migrate"
	_ "github.com/kyleconroy/sqlc/cmd/sqlc"
)
