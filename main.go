package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

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
    -i INTERFACE    Specify network interface (default: auto-detect)

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

func main() {
	var targetMAC string
	var iface string

	// Argument parsing
	args := []string{}
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		if arg == "-h" || arg == "--help" {
			printHelp()
			return
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

	// Find interface if not specified
	if iface == "" {
		devices, err := pcap.FindAllDevs()
		if err != nil {
			log.Fatal("Error finding devices:", err)
		}
		if len(devices) == 0 {
			log.Fatal("No network interfaces found")
		}

		// Find first non-loopback interface with addresses
		for _, device := range devices {
			if len(device.Addresses) > 0 && device.Name != "lo" {
				iface = device.Name
				break
			}
		}

		if iface == "" {
			iface = devices[0].Name
		}
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
