# mac2ip

🇬🇧 [English](README.md) | 🇩🇪 [Deutsch](README_DE.md)

Live ARP-Traffic-Sniffer für MAC-Adressen-Lookups. Finde IP-Adressen zu MAC-Adressen in Echtzeit durch Beobachtung der Netzwerk-Kommunikation.

**Wichtig:** mac2ip sieht nur ARP-Pakete, die NACH dem Start gesendet werden. Für beste Ergebnisse das Gerät nach dem Start von mac2ip einstecken oder einschalten.

## Features

- 🔍 **Live ARP-Monitoring** - Zeigt IP-Zuweisungen in Echtzeit
- 🌐 **Layer 2 Protokoll** - Funktioniert auch bei falscher IP-Konfiguration
- 🎯 **Partial Matching** - Suche nach vollständiger MAC oder Teilmuster (z.B. `dd:ee`)
- 🖥️ **Cross-Platform** - Linux, macOS, Windows
- 📦 **Single Binary** - Benötigt nur libpcap/Npcap

## Installation

**Empfohlen:** Aus Quellcode bauen für beste Kompatibilität und Performance. Vorkompilierte Binaries sind auch verfügbar und sollten auf den meisten Systemen funktionieren.

### Aus Quellcode bauen (Empfohlen)

```bash
git clone https://github.com/revlat/mac2ip.git
cd mac2ip
make build
sudo mv mac2ip /usr/local/bin/
```

Selbst-Bauen stellt sicher, dass das Binary für dein spezifisches System optimiert ist und deine installierte libpcap-Version nutzt.

### Download Release

Lade vorkompilierte Binaries von [Releases](https://github.com/revlat/mac2ip/releases):

```bash
# Linux/macOS
wget https://github.com/revlat/mac2ip/releases/latest/download/mac2ip-linux-amd64.tar.gz
tar -xzf mac2ip-linux-amd64.tar.gz
sudo mv mac2ip /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri https://github.com/revlat/mac2ip/releases/latest/download/mac2ip-windows-amd64.zip -OutFile mac2ip.zip
Expand-Archive mac2ip.zip
```

## Verwendung

**Alle ARP-Pakete anzeigen:**
```bash
sudo mac2ip
```

**Spezifische MAC überwachen:**
```bash
sudo mac2ip aa:bb:cc:dd:ee:ff
```

**Teilübereinstimmung (findet alle MACs mit diesem Muster):**
```bash
sudo mac2ip dd:ee          # Matched aa:bb:cc:dd:ee:ff
sudo mac2ip 00:1a:2b       # Suche nach Hersteller-Prefix
```

**Verfügbare Interfaces auflisten:**
```bash
mac2ip --list
```

**Interface angeben** (per Name, Index aus `--list` oder Teil der Beschreibung):
```bash
sudo mac2ip -i eth0 aa:bb:cc:dd:ee:ff
sudo mac2ip -i 2 aa:bb:cc:dd:ee:ff          # per Index
sudo mac2ip -i "Ethernet" aa:bb:cc:dd:ee:ff # per Beschreibung (praktisch unter Windows)
```

**Hilfe:**
```bash
mac2ip --help
```

### Ohne sudo ausführen (Linux/macOS)

Statt jedes Mal `sudo` zu nutzen, kannst du dem Binary die **CAP_NET_RAW** Capability geben:

```bash
sudo setcap cap_net_raw+ep ./mac2ip
```

**Jetzt ohne sudo ausführbar:**
```bash
./mac2ip aa:bb:cc:dd:ee:ff
```

**Hinweise:**
- Sicherer als `sudo` (gibt nur Packet-Capture-Berechtigung, keine vollen Root-Rechte)
- Muss nach jedem `make build` neu gesetzt werden
- Nur Linux/macOS

### Unter Windows ausführen

Windows benötigt typischerweise **Administrator-Rechte** für Packet Capture.

**Standard: Als Administrator ausführen**
- Rechtsklick auf `mac2ip.exe` → **Als Administrator ausführen**
- Oder PowerShell/CMD als Administrator öffnen:
  ```cmd
  .\mac2ip.exe aa:bb:cc:dd:ee:ff
  ```

**Den richtigen Adapter unter Windows wählen**

Unter Windows meldet Npcap Interfaces als kryptische Gerätepfade wie
`\Device\NPF_{GUID}` statt lesbarer Namen. Mit `--list` siehst du alle Adapter
mit Beschreibung und IP-Adressen und kannst dann per Index oder per Teil der
Beschreibung auswählen:

```cmd
.\mac2ip.exe --list
.\mac2ip.exe -i 2 aa:bb:cc:dd:ee:ff
.\mac2ip.exe -i "Ethernet" aa:bb:cc:dd:ee:ff
```

Die Auto-Erkennung (ohne `-i`) überspringt Loopback- und virtuelle/APIPA-Adapter
und bevorzugt das Interface mit einer echten IPv4-Adresse.

**Optional: Ohne Admin-Rechte (Fortgeschritten)**

Du kannst Npcap so konfigurieren, dass normale User es nutzen können:

1. Während der **Npcap-Installation**, deaktiviere:
   ☐ *"Restrict Npcap driver's access to Administrators only"*
2. Dann können normale User `mac2ip.exe` ohne Admin-Rechte ausführen

**Hinweis:** Dies ist ein Sicherheitsrisiko (jede App kann Pakete sniffen) und **nicht empfohlen** für Produktivsysteme.

**Voraussetzungen:**
- [Npcap](https://npcap.com/#download) muss installiert sein (WinPcap-Nachfolger)
- **Tipp:** [Wireshark](https://www.wireshark.org/) enthält bereits Npcap!

## Beispiel-Ausgabe

```
🔍 Listening for ARP traffic with MAC pattern '18:69' on eno1...
   Press Ctrl+C to stop

[17:08:50] REP | 60:cf:84:cb:18:69 → 192.168.10.130
[17:09:06] REQ | 60:cf:84:cb:18:69 → 192.168.10.130
[17:09:36] REP | 60:cf:84:cb:18:69 → 192.168.10.130
[17:10:06] REP | 60:cf:84:cb:18:69 → 192.168.10.130
[17:10:06] REQ | 60:cf:84:cb:18:69 → 192.168.10.130
```

## Anwendungsfälle

**Einfacher als Wireshark für diese Aufgabe** - Keine komplexe UI oder Filter, einfach MAC-Adresse eingeben. Perfekt wenn du keinen Zugriff auf DHCP-Server oder Router hast!

### IP eines Geräts finden
Ein Kollege hat ein neues Gerät und braucht die IP:

```bash
# 1. MAC-Adresse vom Geräte-Aufkleber ablesen
# 2. mac2ip ZUERST starten
mac2ip a1:b2:c3:d4:e5:f6

# 3. JETZT Gerät einstecken/einschalten
# 4. Sofort sichtbar: [16:34:21] REQ | a1:b2:c3:d4:e5:f6 → 192.168.10.42
# Fertig! Gerät hat IP 192.168.10.42
```

**Funktioniert ohne:**
- DHCP-Server-Zugriff (keine Admin-Rechte nötig)
- Router/Switch-Login-Daten
- Komplexe Wireshark-Filter oder Netzwerk-Kenntnisse

### "Mystery Devices" finden
Ein Gerät ist im Netzwerk, aber du kennst die IP nicht:

```bash
# 1. MAC-Adresse vom Aufkleber checken
# 2. mac2ip starten
mac2ip 00:1a:2b:3c:4d:5e

# 3. Gerät neu starten (ausstecken/einstecken oder rebooten)
# 4. Auf ARP-Aktivität warten - zeigt IP sobald das Gerät kommuniziert
```

**Typische Szenarien:**
- Neuer Drucker, Tablet oder IoT-Gerät
- Temporäre Geräte von externen Dienstleistern
- Geräte mit falscher/statischer IP-Konfiguration

## Voraussetzungen

- **Root/Admin-Rechte** (benötigt Promiscuous Mode)
- **libpcap** (Linux/macOS) oder **WinPcap/Npcap** (Windows)

### libpcap installieren

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
- Installiere [Npcap](https://npcap.com/#download) (WinPcap-Nachfolger)
- **Hinweis:** Wenn du [Wireshark](https://www.wireshark.org/) installiert hast, ist Npcap bereits enthalten!

## Build

```bash
# Lokaler Build (aktuelles OS)
make build

# Alle Plattformen
make build-all

# Aufräumen
make clean
```

## Lizenz

MIT
