//go:build windows

package platform

import (
	winEnum "github.com/porthog/porthog/internal/adapters/os/windows"
	"github.com/porthog/porthog/internal/core/ports"
)

// NewEnumerator returns the platform-specific port enumerator.
func NewEnumerator() ports.Enumerator {
	return winEnum.NewEnumerator()
}
