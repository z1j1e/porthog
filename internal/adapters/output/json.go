package output

import (
	"encoding/json"
	"time"

	"github.com/z1j1e/porthog/internal/core/domain"
)

type jsonEnvelope struct {
	SchemaVersion string        `json:"schema_version"`
	Command       string        `json:"command"`
	Timestamp     string        `json:"timestamp"`
	Data          []jsonBinding `json:"data"`
	Errors        []string      `json:"errors,omitempty"`
	Partial       bool          `json:"partial,omitempty"`
}

type jsonBinding struct {
	Protocol   string       `json:"protocol"`
	LocalAddr  string       `json:"local_addr"`
	LocalPort  uint16       `json:"local_port"`
	RemoteAddr string       `json:"remote_addr,omitempty"`
	RemotePort uint16       `json:"remote_port,omitempty"`
	State      string       `json:"state"`
	PID        int32        `json:"pid"`
	Process    *jsonProcess `json:"process,omitempty"`
}

type jsonProcess struct {
	Name     string `json:"name,omitempty"`
	Exe      string `json:"exe,omitempty"`
	Username string `json:"username,omitempty"`
}

func (r *Renderer) renderJSON(result *domain.PartialResult[[]domain.PortBinding], cmd string) error {
	env := jsonEnvelope{
		SchemaVersion: "1.0",
		Command:       cmd,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Partial:       result.Partial,
		Errors:        result.Warnings,
	}

	for _, b := range result.Data {
		jb := jsonBinding{
			Protocol:  b.Protocol.String(),
			LocalAddr: b.LocalIP.String(),
			LocalPort: b.LocalPort,
			State:     b.State.String(),
			PID:       b.PID,
		}
		if b.RemoteIP != nil {
			jb.RemoteAddr = b.RemoteIP.String()
			jb.RemotePort = b.RemotePort
		}
		if b.Process != nil {
			jb.Process = &jsonProcess{
				Name: b.Process.Name, Exe: b.Process.Exe, Username: b.Process.Username,
			}
		}
		env.Data = append(env.Data, jb)
	}

	enc := json.NewEncoder(r.w)
	enc.SetIndent("", "  ")
	return enc.Encode(env)
}
