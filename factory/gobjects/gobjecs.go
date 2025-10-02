// Package gobjects provides a prototype for generic objects.
// This package also handles IGoBE interface to avoid circular dependencies.
package gobjects

import (
	gb "github.com/kubex-ecosystem/gobe"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
)

type GobJect interface{}

type IGoBE interface {
	ci.IGoBE
}
type GoBE = gb.GoBE
