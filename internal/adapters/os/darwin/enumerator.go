//go:build darwin

package darwin

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/porthog/porthog/internal/core/domain"
)

// Enumerator implements port enumeration on macOS via lsof.
type Enumerator struct{}

func NewEnumerator() *Enumerator { return &Enumerator{} }

func (e *Enumerator) List(ctx context.Context, filter *domain.Filter) (*domain.PartialResult[[]domain.PortBinding], error) {
	var warnings []string
	bindings, err := runLsof(ctx)
	if err != nil {
		warnings = append(warnings, "lsof enumeration error: "+err.Error())
	}

	if filter != nil {
		var filtered []domain.PortBinding
		for i := range bindings {
			if filter.Matches(&bindings[i]) {
				filtered = append(filtered, bindings[i])
			}
		}
		bindings = filtered
	}

	return &domain.PartialResult[[]domain.PortBinding]{
		Data: bindings, Warnings: warnings,
	}, nil
}
// PLACEHOLDER_DARWIN_LSOF

func runLsof(ctx context.Context) ([]domain.PortBinding, error) {
	cmd := exec.CommandContext(ctx, "lsof", "-i", "-n", "-P", "-F", "pcnPtTs")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("lsof: %w", err)
	}
	return parseLsofOutput(string(out))
}

func parseLsofOutput(output string) ([]domain.PortBinding, error) {
	var bindings []domain.PortBinding
	var pid int32
	var pname string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 {
			continue
		}
		key, val := line[0], line[1:]
		switch key {
		case 'p':
			p, _ := strconv.ParseInt(val, 10, 32)
			pid = int32(p)
		case 'c':
			pname = val
		case 'n':
			b, ok := parseNameField(val, pid, pname)
			if ok {
				bindings = append(bindings, b)
			}
		}
	}
	return bindings, nil
}

func parseNameField(name string, pid int32, pname string) (domain.PortBinding, bool) {
	// lsof -F n format: "host:port" or "host:port->remote:port"
	parts := strings.SplitN(name, "->", 2)
	localIP, localPort, ok := parseHostPort(parts[0])
	if !ok {
		return domain.PortBinding{}, false
	}

	b := domain.PortBinding{
		Protocol:  domain.TCP,
		LocalIP:   localIP,
		LocalPort: localPort,
		State:     domain.StateListen,
		PID:       pid,
		Process:   &domain.ProcessIdentity{PID: pid, Name: pname},
	}

	if len(parts) == 2 {
		rIP, rPort, ok := parseHostPort(parts[1])
		if ok {
			b.RemoteIP = rIP
			b.RemotePort = rPort
			b.State = domain.StateEstablished
		}
	}
	return b, true
}

func parseHostPort(s string) (net.IP, uint16, bool) {
	idx := strings.LastIndex(s, ":")
	if idx < 0 {
		return nil, 0, false
	}
	host := s[:idx]
	port, err := strconv.ParseUint(s[idx+1:], 10, 16)
	if err != nil {
		return nil, 0, false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		ip = net.IPv4zero
	}
	return ip, uint16(port), true
}
