//go:build linux

package linux

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// mapInodeToPID scans /proc/<pid>/fd to find which PID owns a given socket inode.
func mapInodeToPID(inode uint64) (int32, error) {
	target := fmt.Sprintf("socket:[%d]", inode)

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.ParseInt(entry.Name(), 10, 32)
		if err != nil {
			continue
		}

		fdDir := filepath.Join("/proc", entry.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			// hidepid or permission denied — skip silently
			continue
		}

		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}
			if link == target {
				return int32(pid), nil
			}
		}
	}
	return 0, fmt.Errorf("no process found for inode %d", inode)
}

// mapInodesToPIDs batch-maps socket inodes to PIDs.
// Returns a map of inode -> PID. Handles hidepid gracefully by skipping
// inaccessible /proc entries.
func mapInodesToPIDs(inodes []uint64) map[uint64]int32 {
	if len(inodes) == 0 {
		return nil
	}

	inodeSet := make(map[uint64]bool, len(inodes))
	for _, ino := range inodes {
		inodeSet[ino] = true
	}

	result := make(map[uint64]int32, len(inodes))

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return result
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.ParseInt(entry.Name(), 10, 32)
		if err != nil {
			continue
		}

		fdDir := filepath.Join("/proc", entry.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			// hidepid mount option or permission denied — skip gracefully
			continue
		}

		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}
			if strings.HasPrefix(link, "socket:[") {
				inoStr := link[8 : len(link)-1]
				ino, err := strconv.ParseUint(inoStr, 10, 64)
				if err != nil {
					continue
				}
				if inodeSet[ino] {
					result[ino] = int32(pid)
					delete(inodeSet, ino)
					if len(inodeSet) == 0 {
						return result
					}
				}
			}
		}
	}
	return result
}
