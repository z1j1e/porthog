//go:build windows

package windows

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/porthog/porthog/internal/core/domain"
)

const (
	tcpTableOwnerPIDAll = 5
	udpTableOwnerPID    = 1
	afINET              = 2
)

var (
	modIPHlpAPI             = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetExtendedTcpTable = modIPHlpAPI.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable = modIPHlpAPI.NewProc("GetExtendedUdpTable")
)

// Enumerator implements port enumeration via Windows IP Helper APIs.
type Enumerator struct{}

func NewEnumerator() *Enumerator { return &Enumerator{} }

func (e *Enumerator) List(ctx context.Context, filter *domain.Filter) (*domain.PartialResult[[]domain.PortBinding], error) {
	var bindings []domain.PortBinding
	var warnings []string

	wantTCP := filter == nil || len(filter.Protocols) == 0 || hasProto(filter.Protocols, domain.TCP)
	wantUDP := filter == nil || len(filter.Protocols) == 0 || hasProto(filter.Protocols, domain.UDP)

	if wantTCP {
		tcp, err := enumTCP()
		if err != nil {
			warnings = append(warnings, "TCP enumeration failed: "+err.Error())
		} else {
			bindings = append(bindings, tcp...)
		}
	}
	if wantUDP {
		udp, err := enumUDP()
		if err != nil {
			warnings = append(warnings, "UDP enumeration failed: "+err.Error())
		} else {
			bindings = append(bindings, udp...)
		}
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

	return &domain.PartialResult[[]domain.PortBinding]{Data: bindings, Warnings: warnings}, nil
}

// --- TCP ---

type mibTCPRowOwnerPID struct {
	State      uint32
	LocalAddr  uint32
	LocalPort  uint32
	RemoteAddr uint32
	RemotePort uint32
	OwningPID  uint32
}
// PLACEHOLDER_TCP_ENUM

func enumTCP() ([]domain.PortBinding, error) {
	var size uint32
	ret, _, _ := procGetExtendedTcpTable.Call(0, uintptr(unsafe.Pointer(&size)), 1, afINET, tcpTableOwnerPIDAll, 0)
	if ret != uintptr(windows.ERROR_INSUFFICIENT_BUFFER) && ret != 0 {
		return nil, fmt.Errorf("GetExtendedTcpTable size query failed: %d", ret)
	}

	buf := make([]byte, size)
	ret, _, _ = procGetExtendedTcpTable.Call(
		uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)),
		1, afINET, tcpTableOwnerPIDAll, 0,
	)
	if ret != 0 {
		return nil, fmt.Errorf("GetExtendedTcpTable failed: %d", ret)
	}

	count := *(*uint32)(unsafe.Pointer(&buf[0]))
	rows := unsafe.Slice((*mibTCPRowOwnerPID)(unsafe.Pointer(&buf[4])), count)

	out := make([]domain.PortBinding, 0, count)
	for _, r := range rows {
		out = append(out, domain.PortBinding{
			Protocol: domain.TCP, LocalIP: u32ToIP(r.LocalAddr), LocalPort: ntohs(r.LocalPort),
			RemoteIP: u32ToIP(r.RemoteAddr), RemotePort: ntohs(r.RemotePort),
			State: mapTCPState(r.State), PID: int32(r.OwningPID),
		})
	}
	return out, nil
}

// --- UDP ---

type mibUDPRowOwnerPID struct {
	LocalAddr uint32
	LocalPort uint32
	OwningPID uint32
}
// PLACEHOLDER_UDP_ENUM

func enumUDP() ([]domain.PortBinding, error) {
	var size uint32
	ret, _, _ := procGetExtendedUdpTable.Call(0, uintptr(unsafe.Pointer(&size)), 1, afINET, udpTableOwnerPID, 0)
	if ret != uintptr(windows.ERROR_INSUFFICIENT_BUFFER) && ret != 0 {
		return nil, fmt.Errorf("GetExtendedUdpTable size query failed: %d", ret)
	}

	buf := make([]byte, size)
	ret, _, _ = procGetExtendedUdpTable.Call(
		uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)),
		1, afINET, udpTableOwnerPID, 0,
	)
	if ret != 0 {
		return nil, fmt.Errorf("GetExtendedUdpTable failed: %d", ret)
	}

	count := *(*uint32)(unsafe.Pointer(&buf[0]))
	rows := unsafe.Slice((*mibUDPRowOwnerPID)(unsafe.Pointer(&buf[4])), count)

	out := make([]domain.PortBinding, 0, count)
	for _, r := range rows {
		out = append(out, domain.PortBinding{
			Protocol: domain.UDP, LocalIP: u32ToIP(r.LocalAddr), LocalPort: ntohs(r.LocalPort),
			State: domain.StateListen, PID: int32(r.OwningPID),
		})
	}
	return out, nil
}

// --- Helpers ---

func ntohs(port uint32) uint16 {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, port)
	return binary.BigEndian.Uint16(b[:2])
}

func u32ToIP(addr uint32) net.IP {
	ip := make(net.IP, 4)
	binary.LittleEndian.PutUint32(ip, addr)
	return ip
}

func mapTCPState(state uint32) domain.SocketState {
	switch state {
	case 2:
		return domain.StateListen
	case 5:
		return domain.StateEstablished
	case 8:
		return domain.StateCloseWait
	case 11:
		return domain.StateTimeWait
	default:
		return domain.StateUnknown
	}
}

func hasProto(protos []domain.Protocol, p domain.Protocol) bool {
	for _, v := range protos {
		if v == p {
			return true
		}
	}
	return false
}
