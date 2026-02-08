//go:build windows

package platform

import (
	winEnum "github.com/z1j1e/porthog/internal/adapters/os/windows"
	"github.com/z1j1e/porthog/internal/core/ports"
)

// NewEnumerator returns the platform-specific port enumerator.
func NewEnumerator() ports.Enumerator {
	return winEnum.NewEnumerator()
}
