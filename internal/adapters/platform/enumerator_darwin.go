//go:build darwin

package platform

import (
	darwinEnum "github.com/porthog/porthog/internal/adapters/os/darwin"
	"github.com/porthog/porthog/internal/core/ports"
)

func NewEnumerator() ports.Enumerator {
	return darwinEnum.NewEnumerator()
}
