package domain

// PortRange represents an inclusive port range.
type PortRange struct {
	Start uint16
	End   uint16
}

// Contains checks if a port falls within this range.
func (r PortRange) Contains(port uint16) bool {
	return port >= r.Start && port <= r.End
}

// Valid returns true if the range is well-formed.
func (r PortRange) Valid() bool {
	return r.Start <= r.End && r.Start > 0
}

// Filter specifies criteria for filtering port bindings.
type Filter struct {
	Protocols []Protocol
	Ports     []uint16
	PortRange *PortRange
	PIDs      []int32
	States    []SocketState
}

// Matches returns true if a PortBinding satisfies this filter.
func (f *Filter) Matches(pb *PortBinding) bool {
	if f == nil {
		return true
	}
	if len(f.Protocols) > 0 && !containsProtocol(f.Protocols, pb.Protocol) {
		return false
	}
	if len(f.Ports) > 0 && !containsPort(f.Ports, pb.LocalPort) {
		return false
	}
	if f.PortRange != nil && !f.PortRange.Contains(pb.LocalPort) {
		return false
	}
	if len(f.PIDs) > 0 && !containsPID(f.PIDs, pb.PID) {
		return false
	}
	if len(f.States) > 0 && !containsState(f.States, pb.State) {
		return false
	}
	return true
}

func containsProtocol(s []Protocol, v Protocol) bool {
	for _, p := range s {
		if p == v {
			return true
		}
	}
	return false
}

func containsPort(s []uint16, v uint16) bool {
	for _, p := range s {
		if p == v {
			return true
		}
	}
	return false
}

func containsPID(s []int32, v int32) bool {
	for _, p := range s {
		if p == v {
			return true
		}
	}
	return false
}

func containsState(s []SocketState, v SocketState) bool {
	for _, st := range s {
		if st == v {
			return true
		}
	}
	return false
}
