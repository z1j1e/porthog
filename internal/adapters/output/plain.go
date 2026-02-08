package output

import (
	"fmt"

	"github.com/porthog/porthog/internal/core/domain"
)

func (r *Renderer) renderPlain(bindings []domain.PortBinding) error {
	for _, b := range bindings {
		name := "-"
		if b.Process != nil && b.Process.Name != "" {
			name = b.Process.Name
		}
		fmt.Fprintf(r.w, "%s\t%s:%d\t%d\t%s\t%s\n",
			b.Protocol, b.LocalIP, b.LocalPort, b.PID, name, b.State)
	}
	return nil
}
