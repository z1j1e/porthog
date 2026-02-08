//go:build darwin

package platform

import (
	darwinEnum "github.com/z1j1e/porthog/internal/adapters/os/darwin"
	"github.com/z1j1e/porthog/internal/core/ports"
)

func NewEnumerator() ports.Enumerator {
	return darwinEnum.NewEnumerator()
}
