//go:build linux

package linux

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/z1j1e/porthog/internal/core/domain"
)

// Enumerator implements port enumeration on Linux.
type Enumerator struct{}

func NewEnumerator() *Enumerator { return &Enumerator{} }

func (e *Enumerator) List(ctx context.Context, filter *domain.Filter) (*domain.PartialResult[[]domain.PortBinding], error) {
	var bindings []domain.PortBinding
	var warnings []string

	wantTCP := filter == nil || len(filter.Protocols) == 0 || hasProto(filter.Protocols, domain.TCP)
	wantUDP := filter == nil || len(filter.Protocols) == 0 || hasProto(filter.Protocols, domain.UDP)

	if wantTCP {
		tcp, err := enumNetlink(unix.IPPROTO_TCP)
		if err != nil {
			tcp, err = parseProcNet("/proc/net/tcp", domain.TCP)
			if err != nil {
				warnings = append(warnings, "TCP enumeration failed: "+err.Error())
			}
		}
		bindings = append(bindings, tcp...)
	}

	if wantUDP {
		udp, err := enumNetlink(unix.IPPROTO_UDP)
		if err != nil {
			udp, err = parseProcNet("/proc/net/udp", domain.UDP)
// PLACEHOLDER_LIST_CONT
			if err != nil {
				warnings = append(warnings, "UDP enumeration failed: "+err.Error())
			}
		}
		bindings = append(bindings, udp...)
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

// enumNetlink uses SOCK_DIAG netlink to enumerate sockets.
func enumNetlink(proto uint8) ([]domain.PortBinding, error) {
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_DGRAM, unix.NETLINK_SOCK_DIAG)
	if err != nil {
		return nil, err
	}
	defer unix.Close(fd)

	req := buildInetDiagReq(proto)
	sa := &unix.SockaddrNetlink{Family: unix.AF_NETLINK}
	if err := unix.Sendto(fd, req, 0, sa); err != nil {
		return nil, err
	}

	var bindings []domain.PortBinding
	buf := make([]byte, 65536)
	for {
		n, _, err := unix.Recvfrom(fd, buf, 0)
		if err != nil {
			return bindings, err
		}
		msgs, err := syscall.ParseNetlinkMessage(buf[:n])
		if err != nil {
			return bindings, fmt.Errorf("parse netlink: %w", err)
		}
		for _, msg := range msgs {
			if msg.Header.Type == syscall.NLMSG_DONE {
				return bindings, nil
			}
			if msg.Header.Type == syscall.NLMSG_ERROR {
				return nil, fmt.Errorf("netlink error response")
			}
			if len(msg.Data) >= 26 {
				b := parseInetDiagMsg(msg.Data, proto)
				bindings = append(bindings, b)
			}
		}
	}
}
// PLACEHOLDER_LINUX_BUILD_REQ

func buildInetDiagReq(proto uint8) []byte {
	const hdrLen = 16
	const msgLen = 56
	buf := make([]byte, hdrLen+msgLen)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(hdrLen+msgLen))
	binary.LittleEndian.PutUint16(buf[4:6], 20) // SOCK_DIAG_BY_FAMILY
	binary.LittleEndian.PutUint16(buf[6:8], unix.NLM_F_DUMP|unix.NLM_F_REQUEST)
	buf[hdrLen] = unix.AF_INET
	buf[hdrLen+1] = proto
	binary.LittleEndian.PutUint32(buf[hdrLen+4:hdrLen+8], 0xFFFFFFFF) // all states
	return buf
}

func parseInetDiagMsg(data []byte, proto uint8) domain.PortBinding {
	p := domain.TCP
	if proto == unix.IPPROTO_UDP {
		p = domain.UDP
	}
	// inet_diag_msg layout:
	// [0] family, [1] state, [2] timer, [3] retrans
	// [4-5] sport (big-endian), [6-7] dport (big-endian)
	// [8-11] src IP, [24-27] dst IP
	state := data[1]
	srcPort := binary.BigEndian.Uint16(data[4:6])
	dstPort := binary.BigEndian.Uint16(data[6:8])
	srcIP := net.IP(make([]byte, 4))
	copy(srcIP, data[8:12])
	dstIP := net.IP(make([]byte, 4))
	copy(dstIP, data[24:28])
	return domain.PortBinding{
		Protocol: p, LocalIP: srcIP, LocalPort: srcPort,
		RemoteIP: dstIP, RemotePort: dstPort,
		State: mapLinuxState(state),
	}
}

func parseProcNet(path string, proto domain.Protocol) ([]domain.PortBinding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var bindings []domain.PortBinding
	scanner := bufio.NewScanner(f)
	scanner.Scan() // skip header
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}
		localIP, localPort := parseHexAddr(fields[1])
		remoteIP, remotePort := parseHexAddr(fields[2])
		st, _ := strconv.ParseUint(fields[3], 16, 8)
// PLACEHOLDER_PROC_CONT
		bindings = append(bindings, domain.PortBinding{
			Protocol: proto, LocalIP: localIP, LocalPort: localPort,
			RemoteIP: remoteIP, RemotePort: remotePort,
			State: mapLinuxState(uint8(st)),
		})
	}
	return bindings, scanner.Err()
}

func parseHexAddr(s string) (net.IP, uint16) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return nil, 0
	}
	ipBytes, _ := hex.DecodeString(parts[0])
	if len(ipBytes) == 4 {
		ipBytes[0], ipBytes[3] = ipBytes[3], ipBytes[0]
		ipBytes[1], ipBytes[2] = ipBytes[2], ipBytes[1]
	}
	port, _ := strconv.ParseUint(parts[1], 16, 16)
	return net.IP(ipBytes), uint16(port)
}

func mapLinuxState(st uint8) domain.SocketState {
	switch st {
	case 0x0A:
		return domain.StateListen
	case 0x01:
		return domain.StateEstablished
	case 0x06:
		return domain.StateTimeWait
	case 0x08:
		return domain.StateCloseWait
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
