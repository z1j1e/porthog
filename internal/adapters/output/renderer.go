package output

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"golang.org/x/term"

	"github.com/z1j1e/porthog/internal/core/domain"
)

// Format represents an output format.
type Format int

const (
	FormatAuto Format = iota
	FormatTable
	FormatJSON
	FormatPlain
)

// Renderer writes port bindings to an output stream.
type Renderer struct {
	w      io.Writer
	format Format
}

func NewRenderer(w io.Writer, format Format) *Renderer {
	return &Renderer{w: w, format: format}
}

// Render outputs port bindings in the configured format.
func (r *Renderer) Render(result *domain.PartialResult[[]domain.PortBinding], cmd string) error {
	f := r.resolveFormat()
	switch f {
	case FormatJSON:
		return r.renderJSON(result, cmd)
	case FormatPlain:
		return r.renderPlain(result.Data)
	default:
		return r.renderTable(result.Data)
	}
}

func (r *Renderer) resolveFormat() Format {
	if r.format != FormatAuto {
		return r.format
	}
	if f, ok := r.w.(*os.File); ok && isatty.IsTerminal(f.Fd()) {
		return FormatTable
	}
	return FormatPlain
}

// termWidth returns the terminal width, defaulting to 120 if detection fails.
func (r *Renderer) termWidth() int {
	if f, ok := r.w.(*os.File); ok {
		w, _, err := term.GetSize(int(f.Fd()))
		if err == nil && w > 0 {
			return w
		}
	}
	return 120
}
