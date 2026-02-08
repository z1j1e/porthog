package domain

import (
	"fmt"
	"net"
)

// Protocol represents a network protocol type.
type Protocol uint8

const (
	TCP Protocol = iota
	UDP
)

func (p Protocol) String() string {
	switch p {
	case TCP:
		return "tcp"
	case UDP:
		return "udp"
	default:
		return "unknown"
	}
}

// SocketState represents the state of a network socket.
type SocketState uint8

const (
	StateUnknown SocketState = iota
	StateListen
	StateEstablished
	StateTimeWait
	StateCloseWait
	StateClosed
)

func (s SocketState) String() string {
	switch s {
	case StateListen:
		return "LISTEN"
	case StateEstablished:
		return "ESTABLISHED"
	case StateTimeWait:
		return "TIME_WAIT"
	case StateCloseWait:
		return "CLOSE_WAIT"
	case StateClosed:
		return "CLOSED"
	default:
		return "UNKNOWN"
	}
}

// PortBinding represents a network socket bound to a port with its owner process.
type PortBinding struct {
	Protocol  Protocol
	LocalIP   net.IP
	LocalPort uint16
	RemoteIP  net.IP
	RemotePort uint16
	State     SocketState
	PID       int32
	Process   *ProcessIdentity
}

// IsListening returns true if the socket is in LISTEN state.
func (pb *PortBinding) IsListening() bool {
	return pb.State == StateListen
}

// LocalAddr returns the local address as "ip:port".
func (pb *PortBinding) LocalAddr() string {
	return net.JoinHostPort(pb.LocalIP.String(), fmt.Sprintf("%d", pb.LocalPort))
}

// RemoteAddr returns the remote address as "ip:port".
func (pb *PortBinding) RemoteAddr() string {
	if pb.RemoteIP == nil {
		return ""
	}
	return net.JoinHostPort(pb.RemoteIP.String(), fmt.Sprintf("%d", pb.RemotePort))
}
