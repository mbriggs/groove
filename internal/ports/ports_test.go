package ports

import (
	"testing"
)

func TestFindFree(t *testing.T) {
	claimed := map[int]string{}
	port, err := FindFree(claimed)
	if err != nil {
		t.Fatalf("FindFree() error: %v", err)
	}
	if port < 3000 || port > 9999 {
		t.Fatalf("FindFree() = %d, want port in range 3000-9999", port)
	}
}

func TestFindFreeAvoidsClaimedPorts(t *testing.T) {
	claimed := map[int]string{}
	// Claim a port, then find a new one — should not collide
	first, err := FindFree(claimed)
	if err != nil {
		t.Fatalf("FindFree() error: %v", err)
	}
	claimed[first] = "test:first"

	second, err := FindFree(claimed)
	if err != nil {
		t.Fatalf("FindFree() error: %v", err)
	}
	if second == first {
		t.Fatalf("FindFree() returned same port %d twice", first)
	}
}

func TestFindFreeAvoidsWellKnownPorts(t *testing.T) {
	claimed := map[int]string{}
	for i := 0; i < 50; i++ {
		port, err := FindFree(claimed)
		if err != nil {
			t.Fatalf("FindFree() error: %v", err)
		}
		if wellKnownPorts[port] {
			t.Fatalf("FindFree() returned well-known port %d", port)
		}
		claimed[port] = "test"
	}
}

func TestClaimPortsFreePorts(t *testing.T) {
	configPorts := map[string]*int{
		"web":   nil,
		"redis": nil,
	}
	claimed := map[int]string{}

	result, err := ClaimPorts(configPorts, claimed, nil)
	if err != nil {
		t.Fatalf("ClaimPorts() error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("ClaimPorts() returned %d ports, want 2", len(result))
	}

	if result["web"] == result["redis"] {
		t.Fatalf("ClaimPorts() assigned same port %d to web and redis", result["web"])
	}

	for name, port := range result {
		if port < 3000 || port > 9999 {
			t.Errorf("port %s = %d, want in range 3000-9999", name, port)
		}
	}
}

func TestClaimPortsFixedPort(t *testing.T) {
	port := 5432
	configPorts := map[string]*int{
		"db": &port,
	}
	claimed := map[int]string{}

	result, err := ClaimPorts(configPorts, claimed, nil)
	if err != nil {
		t.Fatalf("ClaimPorts() error: %v", err)
	}

	if result["db"] != 5432 {
		t.Fatalf("ClaimPorts() db = %d, want 5432", result["db"])
	}
}

func TestClaimPortsFixedPortConflictDeclined(t *testing.T) {
	port := 5432
	configPorts := map[string]*int{
		"db": &port,
	}
	claimed := map[int]string{5432: "other-wt:db"}

	result, err := ClaimPorts(configPorts, claimed, func(name string, p int, claimedBy string) bool {
		return false // decline
	})
	if err != nil {
		t.Fatalf("ClaimPorts() error: %v", err)
	}

	if _, ok := result["db"]; ok {
		t.Fatal("ClaimPorts() should not have assigned db port when declined")
	}
}

func TestClaimPortsFixedPortConflictAccepted(t *testing.T) {
	port := 5432
	configPorts := map[string]*int{
		"db": &port,
	}
	claimed := map[int]string{5432: "other-wt:db"}

	result, err := ClaimPorts(configPorts, claimed, func(name string, p int, claimedBy string) bool {
		return true // accept
	})
	if err != nil {
		t.Fatalf("ClaimPorts() error: %v", err)
	}

	if result["db"] != 5432 {
		t.Fatalf("ClaimPorts() db = %d, want 5432", result["db"])
	}
}

func TestClaimPortsMultipleFreePortsDontCollide(t *testing.T) {
	configPorts := map[string]*int{
		"a": nil,
		"b": nil,
		"c": nil,
		"d": nil,
		"e": nil,
	}
	claimed := map[int]string{}

	result, err := ClaimPorts(configPorts, claimed, nil)
	if err != nil {
		t.Fatalf("ClaimPorts() error: %v", err)
	}

	seen := map[int]string{}
	for name, port := range result {
		if other, ok := seen[port]; ok {
			t.Fatalf("port %d assigned to both %s and %s", port, other, name)
		}
		seen[port] = name
	}
}
