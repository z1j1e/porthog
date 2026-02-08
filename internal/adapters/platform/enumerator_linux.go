//go:build linux

package platform

import (
	linuxEnum "github.com/porthog/porthog/internal/adapters/os/linux"
	"github.com/porthog/porthog/internal/core/ports"
)

func NewEnumerator() ports.Enumerator {
	return linuxEnum.NewEnumerator()
}
