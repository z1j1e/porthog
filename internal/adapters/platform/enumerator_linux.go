//go:build linux

package platform

import (
	linuxEnum "github.com/z1j1e/porthog/internal/adapters/os/linux"
	"github.com/z1j1e/porthog/internal/core/ports"
)

func NewEnumerator() ports.Enumerator {
	return linuxEnum.NewEnumerator()
}
