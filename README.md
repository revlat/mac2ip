# mac2ip

🇬🇧 [English](README.md) | 🇩🇪 [Deutsch](README_DE.md)

Live ARP traffic sniffer for MAC address lookup. Find IP addresses associated with a MAC address in real-time by monitoring network communication.

**Important:** mac2ip only sees ARP packets sent AFTER you start it. For best results, plug in or power on the device after starting mac2ip.

## Features

- 🔍 **Live ARP monitoring** - Shows IP assignments as they happen
- 🌐 **Layer 2 protocol** - Works even if device has wrong IP configuration
- 🎯 **Partial matching** - Search by full MAC or partial pattern (e.g., `dd:ee`)
- 🖥️ **Cross-platform** - Linux, macOS, Windows
- 📦 **Single binary** - Just needs libpcap/Npcap

## Installation

### Download Release

Download pre-built binaries from [Releases](https://github.com/revlat/mac2ip/releases):

```bash
# Linux/macOS
wget https://github.com/revlat/mac2ip/releases/latest/download/mac2ip-linux-amd64.tar.gz
tar -xzf mac2ip-linux-amd64.tar.gz
sudo mv mac2ip /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri https://github.com/revlat/mac2ip/releases/latest/download/mac2ip-windows-amd64.zip -OutFile mac2ip.zip
Expand-Archive mac2ip.zip
```

### Build from Source

```bash
git clone https://github.com/revlat/mac2ip.git
cd mac2ip
make build
```

## Usage

**Monitor ALL ARP traffic:**
```bash
sudo mac2ip
```

**Monitor specific MAC:**
```bash
sudo mac2ip aa:bb:cc:dd:ee:ff
```

**Partial match (finds all MACs containing pattern):**
```bash
sudo mac2ip dd:ee          # Matches aa:bb:cc:dd:ee:ff
sudo mac2ip 00:1a:2b       # Find by vendor prefix
```

**Specify interface:**
```bash
sudo mac2ip -i eth0 aa:bb:cc:dd:ee:ff
```

**Help:**
```bash
mac2ip --help
```

### Running without sudo (Linux/macOS)

Instead of using `sudo` every time, you can grant the binary the **CAP_NET_RAW** capability:

```bash
sudo setcap cap_net_raw+ep ./mac2ip
```

**Now you can run without sudo:**
```bash
./mac2ip aa:bb:cc:dd:ee:ff
```

**Notes:**
- More secure than `sudo` (only grants packet capture permission, not full root)
- Must be reapplied after each `make build`
- Linux/macOS only

### Running on Windows

Windows typically **requires Administrator privileges** for packet capture.

**Standard: Run as Administrator**
- Right-click `mac2ip.exe` → **Run as administrator**
- Or open PowerShell/CMD as Administrator:
  ```cmd
  .\mac2ip.exe aa:bb:cc:dd:ee:ff
  ```

**Optional: Non-Admin Access (Advanced)**

You can configure Npcap to allow non-admin users:

1. During **Npcap installation**, uncheck:
   ☐ *"Restrict Npcap driver's access to Administrators only"*
2. Then normal users can run `mac2ip.exe` without admin rights

**Note:** This is a security risk (any app can sniff packets) and **not recommended** for production systems.

**Requirements:**
- [Npcap](https://npcap.com/#download) must be installed (WinPcap successor)
- **Tip:** [Wireshark](https://www.wireshark.org/) already includes Npcap!

## Example Output

```
🔍 Listening for ARP traffic with MAC pattern '18:69' on eno1...
   Press Ctrl+C to stop

[17:08:50] REP | 60:cf:84:cb:18:69 → 192.168.10.130
[17:09:06] REQ | 60:cf:84:cb:18:69 → 192.168.10.130
[17:09:36] REP | 60:cf:84:cb:18:69 → 192.168.10.130
[17:10:06] REP | 60:cf:84:cb:18:69 → 192.168.10.130
[17:10:06] REQ | 60:cf:84:cb:18:69 → 192.168.10.130
```

## Use Cases

**Simpler than Wireshark for this task** - No complex UI or filters, just enter a MAC address. Perfect when you don't have access to the DHCP server or router!

### Finding a Device's IP
A colleague has a new device and needs to know its IP:

```bash
# 1. Read the MAC address from the device label
# 2. Start mac2ip FIRST
mac2ip a1:b2:c3:d4:e5:f6

# 3. NOW plug in/power on the device
# 4. Instantly see: [16:34:21] REQ | a1:b2:c3:d4:e5:f6 → 192.168.10.42
# Done! Device has IP 192.168.10.42
```

**Works without:**
- DHCP server access (no admin rights needed)
- Router/switch login credentials
- Complex Wireshark filters or network knowledge

### Finding "Mystery Devices"
A device is on the network but you don't know its IP:

```bash
# 1. Check the MAC address label
# 2. Start mac2ip
mac2ip 00:1a:2b:3c:4d:5e

# 3. Power cycle the device (unplug/replug or reboot)
# 4. Wait for ARP activity - shows IP when device communicates
```

**Common scenarios:**
- New printer, tablet, or IoT device
- Temporary equipment from external vendors
- Devices with wrong/static IP configuration

## Requirements

- **Root/Admin privileges** (requires promiscuous mode)
- **libpcap** (Linux/macOS) or **WinPcap/Npcap** (Windows)

### Install libpcap

**Debian/Ubuntu:**
```bash
sudo apt install libpcap-dev
```

**Fedora/RHEL:**
```bash
sudo dnf install libpcap-devel
```

**OpenSUSE:**
```bash
sudo zypper install libpcap-devel
```

**macOS:**
```bash
brew install libpcap
```

**Windows:**
- Install [Npcap](https://npcap.com/#download) (WinPcap successor)
- **Note:** If you have [Wireshark](https://www.wireshark.org/) installed, Npcap is already included!

## Build

```bash
# Local build (current OS)
make build

# All platforms
make build-all

# Clean
make clean
```

## License

MIT
