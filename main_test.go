package main

import (
	"net"
	"testing"

	"github.com/google/gopacket/pcap"
)

// windowsDevices simulates what Npcap returns on Windows: opaque
// \Device\NPF_{GUID} names with human-readable descriptions, plus a loopback
// adapter that is NOT named "lo" and a virtual adapter with only a link-local
// (APIPA) address.
func windowsDevices() []pcap.Interface {
	return []pcap.Interface{
		{
			Name:        `\Device\NPF_Loopback`,
			Description: "Adapter for loopback traffic capture",
			Flags:       pcapIfLoopback,
			Addresses:   []pcap.InterfaceAddress{{IP: net.IPv4(127, 0, 0, 1)}},
		},
		{
			Name:        `\Device\NPF_{11111111-1111-1111-1111-111111111111}`,
			Description: "VirtualBox Host-Only Ethernet Adapter",
			Addresses:   []pcap.InterfaceAddress{{IP: net.IPv4(169, 254, 3, 4)}}, // APIPA
		},
		{
			Name:        `\Device\NPF_{22222222-2222-2222-2222-222222222222}`,
			Description: "Intel(R) Ethernet Connection I219-V",
			Addresses:   []pcap.InterfaceAddress{{IP: net.IPv4(192, 168, 1, 50)}},
		},
	}
}

func TestAutoSelectSkipsLoopbackAndAPIPA(t *testing.T) {
	got := autoSelectInterface(windowsDevices())
	want := `\Device\NPF_{22222222-2222-2222-2222-222222222222}`
	if got != want {
		t.Errorf("autoSelectInterface() = %q, want the routable Intel adapter %q", got, want)
	}
}

func TestResolveByDescription(t *testing.T) {
	got, err := resolveInterface(windowsDevices(), "Intel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `\Device\NPF_{22222222-2222-2222-2222-222222222222}`
	if got != want {
		t.Errorf("resolveInterface(\"Intel\") = %q, want %q", got, want)
	}
}

func TestResolveByIndex(t *testing.T) {
	got, err := resolveInterface(windowsDevices(), "3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `\Device\NPF_{22222222-2222-2222-2222-222222222222}`
	if got != want {
		t.Errorf("resolveInterface(\"3\") = %q, want %q", got, want)
	}
}

func TestResolveByExactName(t *testing.T) {
	name := `\Device\NPF_{11111111-1111-1111-1111-111111111111}`
	got, err := resolveInterface(windowsDevices(), name)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != name {
		t.Errorf("resolveInterface(exact) = %q, want %q", got, name)
	}
}

func TestResolveAmbiguous(t *testing.T) {
	// "Adapter" appears in multiple descriptions.
	if _, err := resolveInterface(windowsDevices(), "Adapter"); err == nil {
		t.Error("expected ambiguity error, got nil")
	}
}

func TestResolveOutOfRange(t *testing.T) {
	if _, err := resolveInterface(windowsDevices(), "99"); err == nil {
		t.Error("expected out-of-range error, got nil")
	}
}

func TestResolveNoMatch(t *testing.T) {
	if _, err := resolveInterface(windowsDevices(), "nonexistent"); err == nil {
		t.Error("expected no-match error, got nil")
	}
}

func TestIsLoopbackByFlagAndDescription(t *testing.T) {
	devs := windowsDevices()
	if !isLoopback(devs[0]) {
		t.Error("expected loopback adapter to be detected via flag/description")
	}
	if isLoopback(devs[2]) {
		t.Error("Intel adapter wrongly detected as loopback")
	}
}
