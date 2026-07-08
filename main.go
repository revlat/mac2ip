package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// pcapIfLoopback mirrors libpcap's PCAP_IF_LOOPBACK flag. gopacket exposes the
// raw Flags value but not the constant, so we define it here.
const pcapIfLoopback = 0x00000001

func printHelp() {
	fmt.Println(`NAME
    mac2ip - live ARP traffic sniffer for MAC address lookup

USAGE
    sudo mac2ip [OPTIONS] [MAC_ADDRESS]
    sudo mac2ip -i INTERFACE [MAC_ADDRESS]

DESCRIPTION
    Monitors ARP traffic in real-time and displays IP addresses
    associated with the specified MAC address. Useful for troubleshooting
    network issues, especially when devices have incorrect IP configuration.

OPTIONS
    -h, --help      Show this help message
    -l, --list      List available network interfaces and exit
    -i INTERFACE    Specify network interface (default: auto-detect)
                    Accepts an interface name, a description substring
                    (useful on Windows), or an index from --list

ARGUMENTS
    MAC_ADDRESS     MAC address or pattern to monitor (optional)
                    - Full MAC: aa:bb:cc:dd:ee:ff or aa-bb-cc-dd-ee-ff
                    - Partial: aa:bb or dd:ee (matches anywhere in MAC)
                    - Omit to show ALL ARP traffic

EXAMPLES
    Monitor ALL ARP traffic (no filter):
      sudo mac2ip

    Monitor ARP traffic for a specific MAC:
      sudo mac2ip aa:bb:cc:dd:ee:ff

    Partial match (finds all MACs containing "dd:ee"):
      sudo mac2ip dd:ee

    Vendor prefix (first 3 bytes = manufacturer):
      sudo mac2ip 00:1a:2b

    Specify interface:
      sudo mac2ip -i eth0 aa:bb:cc:dd:ee:ff

    List interfaces (helpful on Windows to find the right adapter):
      mac2ip --list

    Select an interface by index or description (Windows):
      mac2ip -i 2 aa:bb:cc:dd:ee:ff
      mac2ip -i "Ethernet" aa:bb:cc:dd:ee:ff

    Works even if device has wrong IP config (Layer 2):
      sudo mac2ip 00:11:22:33:44:55

NOTES
    - Requires root/admin privileges (uses promiscuous mode)
    - Press Ctrl+C to stop
    - Only shows ARP packets involving the target MAC
    - Layer 2 protocol - works even with wrong IP configuration`)
}

func normalizeMAC(mac string) string {
	// Normalize MAC: lowercase, replace - with :
	return strings.ToLower(strings.ReplaceAll(mac, "-", ":"))
}

func formatMAC(addr []byte) string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		addr[0], addr[1], addr[2], addr[3], addr[4], addr[5])
}

func formatIP(addr []byte) string {
	return fmt.Sprintf("%d.%d.%d.%d", addr[0], addr[1], addr[2], addr[3])
}

// isLoopback reports whether a device is a loopback interface. Windows does not
// name its loopback "lo", so we also check the pcap flag and the description.
func isLoopback(device pcap.Interface) bool {
	if device.Name == "lo" {
		return true
	}
	if device.Flags&pcapIfLoopback != 0 {
		return true
	}
	return strings.Contains(strings.ToLower(device.Description), "loopback")
}

// hasRoutableIPv4 reports whether a device has a "real" IPv4 address, i.e. not
// loopback and not link-local (169.254.x.x / APIPA). Such interfaces are the
// ones actually carrying LAN traffic and make the best auto-detect candidates.
func hasRoutableIPv4(device pcap.Interface) bool {
	for _, a := range device.Addresses {
		ip := a.IP.To4()
		if ip == nil {
			continue
		}
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			continue
		}
		return true
	}
	return false
}

// ipv4Strings returns the IPv4 addresses of a device as printable strings.
func ipv4Strings(device pcap.Interface) []string {
	var ips []string
	for _, a := range device.Addresses {
		if ip := a.IP.To4(); ip != nil {
			ips = append(ips, ip.String())
		}
	}
	return ips
}

// listInterfaces prints all available capture interfaces with an index, name,
// description and IPv4 addresses. On Windows the name is an opaque
// \Device\NPF_{GUID} path, so the description and index are what users need.
func listInterfaces(devices []pcap.Interface) {
	fmt.Println("Available network interfaces:")
	fmt.Println()
	for i, device := range devices {
		fmt.Printf("  [%d] %s\n", i+1, device.Name)
		if device.Description != "" {
			fmt.Printf("      Description: %s\n", device.Description)
		}
		if ips := ipv4Strings(device); len(ips) > 0 {
			fmt.Printf("      IPv4:        %s\n", strings.Join(ips, ", "))
		}
		if isLoopback(device) {
			fmt.Printf("      (loopback)\n")
		}
		fmt.Println()
	}
	fmt.Println("Select one with:  mac2ip -i <index|name|description>")
}

// resolveInterface maps a user-supplied -i value to a concrete device name.
// It accepts, in order of precedence: an exact interface name, a 1-based index
// from --list, or a case-insensitive substring of the name or description
// (which is how users pick adapters on Windows). Returns an error describing
// the problem, including ambiguous matches.
func resolveInterface(devices []pcap.Interface, spec string) (string, error) {
	// 1) Exact name match.
	for _, device := range devices {
		if device.Name == spec {
			return device.Name, nil
		}
	}

	// 2) Numeric index (1-based, as shown by --list).
	if idx, err := strconv.Atoi(spec); err == nil {
		if idx < 1 || idx > len(devices) {
			return "", fmt.Errorf("interface index %d out of range (have %d interfaces, use --list)", idx, len(devices))
		}
		return devices[idx-1].Name, nil
	}

	// 3) Case-insensitive substring of name or description.
	needle := strings.ToLower(spec)
	var matches []pcap.Interface
	for _, device := range devices {
		if strings.Contains(strings.ToLower(device.Name), needle) ||
			strings.Contains(strings.ToLower(device.Description), needle) {
			matches = append(matches, device)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no interface matches %q (use --list to see available interfaces)", spec)
	case 1:
		return matches[0].Name, nil
	default:
		var b strings.Builder
		fmt.Fprintf(&b, "%q matches multiple interfaces:\n", spec)
		for _, m := range matches {
			fmt.Fprintf(&b, "  - %s", m.Name)
			if m.Description != "" {
				fmt.Fprintf(&b, " (%s)", m.Description)
			}
			b.WriteString("\n")
		}
		b.WriteString("Be more specific or use --list to pick by index")
		return "", fmt.Errorf("%s", b.String())
	}
}

// autoSelectInterface picks a sensible default capture interface: the first
// non-loopback device that has a routable IPv4 address, falling back to the
// first non-loopback device with any address, then to the first device.
func autoSelectInterface(devices []pcap.Interface) string {
	for _, device := range devices {
		if !isLoopback(device) && hasRoutableIPv4(device) {
			return device.Name
		}
	}
	for _, device := range devices {
		if !isLoopback(device) && len(device.Addresses) > 0 {
			return device.Name
		}
	}
	return devices[0].Name
}

func main() {
	var targetMAC string
	var iface string
	var listReq bool

	// Argument parsing
	args := []string{}
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		if arg == "-h" || arg == "--help" {
			printHelp()
			return
		} else if arg == "-l" || arg == "--list" {
			listReq = true
		} else if arg == "-i" {
			if i+1 < len(os.Args) {
				iface = os.Args[i+1]
				i++ // Skip next arg
			} else {
				fmt.Println("Error: -i requires interface name")
				printHelp()
				return
			}
		} else {
			args = append(args, arg)
		}
	}

	if len(args) > 1 {
		fmt.Println("Error: Too many arguments")
		printHelp()
		return
	}

	// If no MAC specified, show all ARP traffic
	if len(args) == 1 {
		targetMAC = normalizeMAC(args[0])
	} else {
		targetMAC = "" // Empty = show all
	}

	// Enumerate interfaces (needed for --list, auto-detect and -i resolution).
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal("Error finding devices:", err)
	}
	if len(devices) == 0 {
		log.Fatal("No network interfaces found")
	}

	if listReq {
		listInterfaces(devices)
		return
	}

	if iface == "" {
		// Auto-detect a sensible default interface.
		iface = autoSelectInterface(devices)
	} else {
		// Resolve -i value (name, index or description).
		resolved, err := resolveInterface(devices, iface)
		if err != nil {
			log.Fatal("Error selecting interface: ", err)
		}
		iface = resolved
	}

	// Open interface for sniffing
	handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
	if err != nil {
		log.Fatal("Error opening interface:", err)
	}
	defer handle.Close()

	// Set BPF filter for ARP only
	if err := handle.SetBPFFilter("arp"); err != nil {
		log.Fatal("Error setting BPF filter:", err)
	}

	if targetMAC != "" {
		fmt.Printf("🔍 Listening for ARP traffic with MAC pattern '%s' on %s...\n", targetMAC, iface)
	} else {
		fmt.Printf("🔍 Listening for ALL ARP traffic on %s...\n", iface)
	}
	fmt.Printf("   Press Ctrl+C to stop\n\n")

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		arpLayer := packet.Layer(layers.LayerTypeARP)
		if arpLayer == nil {
			continue
		}

		arp := arpLayer.(*layers.ARP)

		srcMAC := formatMAC(arp.SourceHwAddress)

		// Filter by MAC pattern if specified (only check source MAC)
		if targetMAC != "" {
			if !strings.Contains(srcMAC, targetMAC) {
				continue
			}
		}

		srcIP := formatIP(arp.SourceProtAddress)

		opType := "Request"
		if arp.Operation == 2 {
			opType = "Reply  "
		}

		timestamp := time.Now().Format("15:04:05")

		// Color codes
		green := "\033[92m"
		yellow := "\033[93m"
		blue := "\033[94m"
		reset := "\033[0m"

		// Simple format: [TIME] TYPE | MAC → IP
		// Always show source MAC/IP (most reliable data)
		opLabel := "REQ"
		if opType == "Reply  " {
			opLabel = "REP"
		}

		fmt.Printf("[%s] %s%s%s | %s%s%s → %s%s%s\n",
			timestamp,
			yellow, opLabel, reset,
			green, srcMAC, reset,
			blue, srcIP, reset)
	}
}
