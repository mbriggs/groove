package ports

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
)

// wellKnownPorts that should be avoided when assigning free ports.
var wellKnownPorts = map[int]bool{
	3306: true, // mysql
	5432: true, // postgres
	6379: true, // redis
	8080: true, // common http alt
	8443: true, // common https alt
}

// IsAvailable checks if a port is available by attempting to listen on it.
func IsAvailable(port int) bool {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// FindFree finds an available port in 3000-9999 that isn't claimed or well-known.
func FindFree(claimed map[int]string) (int, error) {
	// Build set of excluded ports
	excluded := make(map[int]bool)
	for p := range claimed {
		excluded[p] = true
	}
	for p := range wellKnownPorts {
		excluded[p] = true
	}

	// Try random ports first for speed
	for i := 0; i < 100; i++ {
		port := 3000 + rand.Intn(7000)
		if excluded[port] {
			continue
		}
		if IsAvailable(port) {
			return port, nil
		}
	}

	// Fall back to sequential scan
	for port := 3000; port <= 9999; port++ {
		if excluded[port] {
			continue
		}
		if IsAvailable(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no free ports available in range 3000-9999")
}

// ClaimPorts assigns ports based on project config.
// configPorts maps name -> *int (nil means free, non-nil means fixed).
// claimed is the current set of claimed ports from state.
// Returns assigned ports and any warnings.
func ClaimPorts(configPorts map[string]*int, claimed map[int]string, confirmFixed func(name string, port int, claimedBy string) bool) (map[string]int, error) {
	result := make(map[string]int)

	for name, fixedPort := range configPorts {
		if fixedPort == nil {
			// Free port — find an available one
			port, err := FindFree(claimed)
			if err != nil {
				return nil, fmt.Errorf("claiming port for %s: %w", name, err)
			}
			result[name] = port
			claimed[port] = "pending:" + name
		} else {
			// Fixed port — check for conflicts
			port := *fixedPort
			if claimedBy, ok := claimed[port]; ok {
				if confirmFixed != nil && !confirmFixed(name, port, claimedBy) {
					continue
				}
			}
			result[name] = port
			claimed[port] = "pending:" + name
		}
	}

	return result, nil
}
