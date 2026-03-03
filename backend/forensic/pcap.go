package forensic

import (
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type PCAPParser struct{}

func (p *PCAPParser) Parse(filePath string) ([]ForensicEvent, error) {
	handle, err := pcap.OpenOffline(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PCAP: %v", err)
	}
	defer handle.Close()

	var events []ForensicEvent
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer != nil {
			ip, _ := ipLayer.(*layers.IPv4)
			
			evt := ForensicEvent{
				"class_name": "Network Traffic",
				"time":       packet.Metadata().Timestamp.UTC().Format(time.RFC3339),
				"src_ip":     ip.SrcIP.String(),
				"dst_ip":     ip.DstIP.String(),
				"protocol":   ip.Protocol.String(),
				"bytes":      packet.Metadata().Length,
			}

			if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
				tcp, _ := tcpLayer.(*layers.TCP)
				evt["src_port"] = int(tcp.SrcPort)
				evt["dst_port"] = int(tcp.DstPort)
			} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
				udp, _ := udpLayer.(*layers.UDP)
				evt["src_port"] = int(udp.SrcPort)
				evt["dst_port"] = int(udp.DstPort)
			}

			evt["raw_data"] = fmt.Sprintf("Packet: %s:%v -> %s:%v [%s] %d bytes", 
				evt["src_ip"], evt["src_port"], evt["dst_ip"], evt["dst_port"], evt["protocol"], evt["bytes"])

			events = append(events, evt)
		}
	}

	return events, nil
}