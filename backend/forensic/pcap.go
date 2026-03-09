package forensic

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

type PCAPParser struct {
	ContentMaxLen int
}

func NewPCAPParser() *PCAPParser {
	return &PCAPParser{ContentMaxLen: 4096}
}

type ForensicEventJSON struct {
	Timestamp    float64  `json:"_time"`
	TimestampISO string   `json:"timestamp,omitempty"`
	SrcIP        string   `json:"src_ip,omitempty"`
	DstIP        string   `json:"dst_ip,omitempty"`
	SrcPort      string   `json:"src_port,omitempty"`
	DstPort      string   `json:"dst_port,omitempty"`
	Network      string   `json:"network,omitempty"`
	Transport    string   `json:"transport,omitempty"`
	Protocol     string   `json:"protocol"`
	Content      string   `json:"content,omitempty"`
	TCPFlags     string   `json:"tcp_flags,omitempty"`
	SYN          bool     `json:"tcp_syn,omitempty"`
	ACK          bool     `json:"tcp_ack,omitempty"`
	FIN          bool     `json:"tcp_fin,omitempty"`
	RST          bool     `json:"tcp_rst,omitempty"`
	HTTPMethod   string   `json:"http_method,omitempty"`
	HTTPURL      string   `json:"http_url,omitempty"`
	HTTPHost     string   `json:"http_host,omitempty"`
	HTTPStatus   int      `json:"http_status,omitempty"`
	DNSQR        bool     `json:"dns_qr,omitempty"`
	DNSRCode     string   `json:"dns_rcode,omitempty"`
	DNSQueries   []string `json:"dns_queries,omitempty"`
	DNSAnswers   []string `json:"dns_answers,omitempty"`
	// USB 相关
	USBEventType string `json:"usb_event_type,omitempty"` // keyboard, mouse, storage
	USBKeyPress  string `json:"usb_keypress,omitempty"`   // 键盘按键
	USBMouseMove string `json:"usb_mouse_move,omitempty"` // 鼠标移动
	// WiFi 相关
	WiFiFrameType string `json:"wifi_frame_type,omitempty"` // beacon, data, control
	WiFiSSID      string `json:"wifi_ssid,omitempty"`       // SSID
	WiFiBSSID     string `json:"wifi_bssid,omitempty"`      // AP MAC
	WiFiChannel   int    `json:"wifi_channel,omitempty"`
	WiFiSignal    int    `json:"wifi_signal_dbm,omitempty"`
	// EAPOL
	EAPOLType string `json:"eapol_type,omitempty"` // 1=Request, 2=Response, 3=Success, 4=Fail
	EAPOLKey  string `json:"eapol_key,omitempty"`  // Key信息
	Tags      string `json:"tags,omitempty"`
}

func (p *PCAPParser) Parse(filePath string) ([]ForensicEvent, error) {
	handle, err := pcap.OpenOffline(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PCAP: %v", err)
	}
	defer handle.Close()

	var events []ForensicEvent
	eventChan := make(chan ForensicEventJSON, 10000)
	var wg sync.WaitGroup

	streamFactory := &pcapStreamFactory{eventChan: eventChan, wg: &wg, contentMaxLen: p.ContentMaxLen}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetSource.DecodeOptions.Lazy = true
	packetSource.DecodeOptions.NoCopy = true

	streamStartTimes := make(map[string]time.Time)
	var timeMu sync.RWMutex

	// USB 键盘映射表 (HID Usage Table)
	usbKeyMap := map[uint8]string{
		0x04: "a", 0x05: "b", 0x06: "c", 0x07: "d", 0x08: "e",
		0x09: "f", 0x0A: "g", 0x0B: "h", 0x0C: "i", 0x0D: "j",
		0x0E: "k", 0x0F: "l", 0x10: "m", 0x11: "n", 0x12: "o",
		0x13: "p", 0x14: "q", 0x15: "r", 0x16: "s", 0x17: "t",
		0x18: "u", 0x19: "v", 0x1A: "w", 0x1B: "x", 0x1C: "y",
		0x1D: "z", 0x1E: "1", 0x1F: "2", 0x20: "3", 0x21: "4",
		0x22: "5", 0x23: "6", 0x24: "7", 0x25: "8", 0x26: "9",
		0x27: "0", 0x28: "\n", 0x29: "esc", 0x2A: "\b",
		0x2C: "space", 0x2D: "-", 0x2E: "=", 0x2F: "[", 0x30: "]",
		0x31: "\\", 0x33: ";", 0x34: "'", 0x35: "`", 0x36: ",",
		0x37: ".", 0x38: "/",
	}

	go func() {
		for ev := range eventChan {
			events = append(events, ForensicEvent{
				"class_name":    "Network Traffic",
				"_time":         ev.Timestamp,
				"timestamp":     ev.TimestampISO,
				"src_ip":        ev.SrcIP,
				"dst_ip":        ev.DstIP,
				"src_port":      ev.SrcPort,
				"dst_port":      ev.DstPort,
				"network":       ev.Network,
				"transport":     ev.Transport,
				"protocol":      ev.Protocol,
				"content":       ev.Content,
				"tcp_flags":     ev.TCPFlags,
				"tcp_syn":       ev.SYN,
				"tcp_ack":       ev.ACK,
				"tcp_fin":       ev.FIN,
				"tcp_rst":       ev.RST,
				"http_method":   ev.HTTPMethod,
				"http_url":      ev.HTTPURL,
				"http_host":     ev.HTTPHost,
				"http_status":   ev.HTTPStatus,
				"dns_qr":        ev.DNSQR,
				"dns_rcode":     ev.DNSRCode,
				"dns_queries":   ev.DNSQueries,
				"dns_answers":   ev.DNSAnswers,
				"usb_event_type": ev.USBEventType,
				"usb_keypress":  ev.USBKeyPress,
				"usb_mouse_move": ev.USBMouseMove,
				"wifi_frame_type": ev.WiFiFrameType,
				"wifi_ssid":     ev.WiFiSSID,
				"wifi_bssid":    ev.WiFiBSSID,
				"wifi_channel":  ev.WiFiChannel,
				"eapol_type":    ev.EAPOLType,
				"eapol_key":     ev.EAPOLKey,
				"tags":          ev.Tags,
			})
		}
	}()

	for packet := range packetSource.Packets() {
		netLayer := packet.NetworkLayer()
		transLayer := packet.TransportLayer()
		ts := packet.Metadata().Timestamp

		// WiFi 802.11 处理
		if wifiLayer := packet.Layer(layers.LayerTypeDot11); wifiLayer != nil {
			wifi, _ := wifiLayer.(*layers.Dot11)
			ev := parseWiFiFrame(wifi, ts)
			if ev.Protocol != "" {
				eventChan <- ev
			}
			continue
		}

		// EAPOL (WiFi 握手) 处理
		if eapolLayer := packet.Layer(layers.LayerTypeEAPOL); eapolLayer != nil {
			eapol, _ := eapolLayer.(*layers.EAPOL)
			ev := parseEAPOL(eapol, ts)
			if ev.Protocol != "" {
				eventChan <- ev
			}
			continue
		}

		// USB 处理
		if usbLayer := packet.Layer(layers.LayerTypeUSB); usbLayer != nil {
			usb, _ := usbLayer.(*layers.USB)
			ev := parseUSB(usb, ts, usbKeyMap)
			if ev.Protocol != "" {
				eventChan <- ev
			}
			continue
		}

		if netLayer == nil || transLayer == nil {
			continue
		}

		netFlow := netLayer.NetworkFlow()
		transFlow := transLayer.TransportFlow()
		srcIP := netFlow.Src().String()
		dstIP := netFlow.Dst().String()
		srcPort := transFlow.Src().String()
		dstPort := transFlow.Dst().String()

		streamID := srcIP + ":" + srcPort + "->" + dstIP + ":" + dstPort

		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)

			timeMu.Lock()
			if _, ok := streamStartTimes[streamID]; !ok {
				streamStartTimes[streamID] = ts
			}
			startTime := streamStartTimes[streamID]
			timeMu.Unlock()

			ev := ForensicEventJSON{
				Timestamp:    float64(startTime.UnixNano()) / 1e9,
				TimestampISO: startTime.Format(time.RFC3339Nano),
				SrcIP:        srcIP,
				DstIP:        dstIP,
				SrcPort:      srcPort,
				DstPort:      dstPort,
				Network:      netFlow.EndpointType().String(),
				Transport:    "TCP",
				Protocol:     "TCP",
				TCPFlags:     buildTCPFlags(tcp),
				SYN:          tcp.SYN,
				ACK:          tcp.ACK,
				FIN:          tcp.FIN,
				RST:          tcp.RST,
			}
			eventChan <- ev
			assembler.AssembleWithTimestamp(netFlow, tcp, ts)
			continue
		}

		if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
			ev := ForensicEventJSON{
				Timestamp:    float64(ts.UnixNano()) / 1e9,
				TimestampISO: ts.Format(time.RFC3339Nano),
				SrcIP:        srcIP,
				DstIP:        dstIP,
				SrcPort:      srcPort,
				DstPort:      dstPort,
				Network:      netFlow.EndpointType().String(),
				Transport:    "UDP",
				Protocol:     "UDP",
			}

			if srcPort == "53" || dstPort == "53" {
				if dnsLayer := packet.Layer(layers.LayerTypeDNS); dnsLayer != nil {
					dns, _ := dnsLayer.(*layers.DNS)
					ev.Protocol = "DNS"
					ev.DNSQR = dns.QR
					ev.DNSRCode = dns.ResponseCode.String()
					for _, q := range dns.Questions {
						ev.DNSQueries = append(ev.DNSQueries, string(q.Name))
					}
					for _, a := range dns.Answers {
						if a.Type == layers.DNSTypeA || a.Type == layers.DNSTypeAAAA {
							ev.DNSAnswers = append(ev.DNSAnswers, a.IP.String())
						} else if a.Type == layers.DNSTypeCNAME {
							ev.DNSAnswers = append(ev.DNSAnswers, string(a.CNAME))
						} else if a.Type == layers.DNSTypeTXT {
							ev.DNSAnswers = append(ev.DNSAnswers, string(a.TXT))
						}
					}
					ev.Tags = "dns"
					if len(ev.DNSAnswers) > 0 {
						for _, ans := range ev.DNSAnswers {
							if len(ans) > 50 {
								ev.Tags = "dns_tunnel"
								break
							}
						}
					}
				}
			}

			if ev.Protocol == "UDP" {
				ev.Protocol = "UDP_STREAM"
				ev.Tags = "udp_stream"
			}
			eventChan <- ev
		}

		// ICMP
		if icmpLayer := packet.Layer(layers.LayerTypeICMPv4); icmpLayer != nil {
			icmp, _ := icmpLayer.(*layers.ICMPv4)
			ev := ForensicEventJSON{
				Timestamp:    float64(ts.UnixNano()) / 1e9,
				TimestampISO: ts.Format(time.RFC3339Nano),
				SrcIP:        srcIP,
				DstIP:        dstIP,
				Network:      "IPv4",
				Transport:    "ICMP",
				Protocol:     "ICMP",
				Tags:         "icmp",
			}
			if len(icmp.Payload) > 64 {
				ev.Tags = "icmp_tunnel"
			}
			eventChan <- ev
		}
	}

	assembler.FlushAll()
	wg.Wait()
	close(eventChan)

	return events, nil
}

// parseWiFi 解析 802.11 WiFi 帧
func parseWiFiFrame(wifi *layers.Dot11, ts time.Time) ForensicEventJSON {
	ev := ForensicEventJSON{
		Timestamp:    float64(ts.UnixNano()) / 1e9,
		TimestampISO: ts.Format(time.RFC3339Nano),
		Network:      "802.11",
		Protocol:     "WIFI",
	}

	// 通过类型和子类型字符串判断
	typeStr := wifi.Type.String()
	ev.WiFiFrameType = typeStr
	
	ev.Tags = "wifi"
	if strings.Contains(typeStr, "Beacon") {
		ev.Tags = "wifi_beacon"
		// 从 Payload 提取 SSID
		if len(wifi.Payload) > 2 && wifi.Payload[0] == 0x00 {
			ssidLen := int(wifi.Payload[1])
			if ssidLen > 0 && ssidLen <= 32 && len(wifi.Payload) >= ssidLen+2 {
				ev.WiFiSSID = string(wifi.Payload[2 : 2+ssidLen])
			}
		}
	} else if strings.Contains(typeStr, "Probe") {
		ev.Tags = "wifi_probe"
	} else if strings.Contains(typeStr, "Auth") {
		ev.Tags = "wifi_auth"
	} else if strings.Contains(typeStr, "Deauth") {
		ev.Tags = "wifi_deauth"
	} else if strings.Contains(typeStr, "Data") {
		ev.Tags = "wifi_data"
	}

	// BSSID
	ev.WiFiBSSID = wifi.Address3.String()

	return ev
}

// parseEAPOL 解析 EAPOL 握手包
func parseEAPOL(eapol *layers.EAPOL, ts time.Time) ForensicEventJSON {
	ev := ForensicEventJSON{
		Timestamp:    float64(ts.UnixNano()) / 1e9,
		TimestampISO: ts.Format(time.RFC3339Nano),
		Network:      "802.1X",
		Protocol:     "EAPOL",
	}

	switch eapol.Type {
	case layers.EAPOLTypeEAP:
		ev.EAPOLType = "EAP"
		ev.Tags = "eap"
	case layers.EAPOLTypeStart:
		ev.EAPOLType = "Start"
		ev.Tags = "eap_start"
	case layers.EAPOLTypeLogOff:
		ev.EAPOLType = "Logoff"
		ev.Tags = "eap_logoff"
	case layers.EAPOLTypeKey:
		ev.EAPOLType = "Key"
		ev.Tags = "wifi_handshake"
		ev.EAPOLKey = fmt.Sprintf("EAPOL Key frame, Length: %d", eapol.Length)
	}

	return ev
}

// parseUSB 解析 USB 流量
func parseUSB(usb *layers.USB, ts time.Time, keyMap map[uint8]string) ForensicEventJSON {
	ev := ForensicEventJSON{
		Timestamp:    float64(ts.UnixNano()) / 1e9,
		TimestampISO: ts.Format(time.RFC3339Nano),
		Network:      "USB",
		Protocol:     "USB",
	}

	// 通过 payload 长度判断类型
	ev.USBEventType = "unknown"
	ev.Tags = "usb"

	// HID 键盘数据检测 (payload 长度通常 >= 8)
	if len(usb.Payload) >= 8 {
		modifier := usb.Payload[0]
		keyCode := usb.Payload[2]

		if keyCode > 0 && keyCode != 0x01 && keyCode != 0x03 {
			if char, ok := keyMap[keyCode]; ok {
				ev.USBEventType = "keyboard"
				ev.USBKeyPress = char
				ev.Tags = "usb_keyboard"

				if modifier&0x22 != 0 {
					ev.USBKeyPress = strings.ToUpper(char)
				}
			}
		}

		// 鼠标
		if len(usb.Payload) >= 4 {
			x := int8(usb.Payload[1])
			y := int8(usb.Payload[2])
			if x != 0 || y != 0 {
				ev.USBEventType = "mouse"
				ev.USBMouseMove = fmt.Sprintf("dx:%d, dy:%d", x, y)
				ev.Tags = "usb_mouse"
			}
		}
	}

	// 存储设备
	if len(usb.Payload) >= 12 && (usb.Payload[0] == 0x28 || usb.Payload[0] == 0x2A) {
		ev.USBEventType = "storage"
		ev.Tags = "usb_storage"
	}

	ev.Content = fmt.Sprintf("USB data length: %d", len(usb.Payload))
	return ev
}

// ========== 保留原有 TCP 流处理 ==========

type pcapStreamFactory struct {
	eventChan      chan ForensicEventJSON
	wg             *sync.WaitGroup
	contentMaxLen int
}

func (f *pcapStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	f.wg.Add(1)
	stream := tcpreader.NewReaderStream()
	go f.handleStream(net, transport, &stream)
	return &stream
}

func (f *pcapStreamFactory) handleStream(netFlow, transFlow gopacket.Flow, r io.Reader) {
	defer f.wg.Done()

	peekBuf := make([]byte, 64)
	n, err := r.Read(peekBuf)
	if err != nil && err != io.EOF {
		return
	}
	if n == 0 {
		return
	}

	peek := peekBuf[:n]
	srcPort := transFlow.Src().String()
	dstPort := transFlow.Dst().String()

	switch {
	case f.matchHTTP(srcPort, dstPort, peek):
		f.parseHTTP(netFlow, transFlow, io.MultiReader(bytes.NewReader(peek), r))
	case f.matchFTP(srcPort, dstPort, peek):
		f.parseFTP(netFlow, transFlow, io.MultiReader(bytes.NewReader(peek), r))
	case f.matchSMTP(srcPort, dstPort, peek):
		f.parseSMTP(netFlow, transFlow, io.MultiReader(bytes.NewReader(peek), r))
	case f.matchTelnet(srcPort, dstPort, peek):
		f.parseTelnet(netFlow, transFlow, io.MultiReader(bytes.NewReader(peek), r))
	case f.matchSMB(srcPort, dstPort, peek):
		f.parseSMB(netFlow, transFlow, io.MultiReader(bytes.NewReader(peek), r))
	case f.matchRedis(srcPort, dstPort, peek):
		f.parseRedis(netFlow, transFlow, io.MultiReader(bytes.NewReader(peek), r))
	case f.matchMySQL(srcPort, dstPort, peek):
		f.parseMySQL(netFlow, transFlow, io.MultiReader(bytes.NewReader(peek), r))
	case f.matchMSSQL(srcPort, dstPort, peek):
		f.parseMSSQL(netFlow, transFlow, io.MultiReader(bytes.NewReader(peek), r))
	case f.matchPOP3(srcPort, dstPort, peek):
		f.parsePOP3(netFlow, transFlow, io.MultiReader(bytes.NewReader(peek), r))
	case f.matchIMAP(srcPort, dstPort, peek):
		f.parseIMAP(netFlow, transFlow, io.MultiReader(bytes.NewReader(peek), r))
	default:
		f.parseRawTCP(netFlow, transFlow, io.MultiReader(bytes.NewReader(peek), r))
	}
}

// 协议匹配器 (简化版，保留核心协议)
func (f *pcapStreamFactory) matchHTTP(srcPort, dstPort string, peek []byte) bool {
	ports := map[string]bool{"80": true, "8080": true, "443": true, "8443": true, "7001": true}
	return ports[srcPort] || ports[dstPort] || strings.HasPrefix(string(peek), "HTTP/")
}

func (f *pcapStreamFactory) matchFTP(srcPort, dstPort string, peek []byte) bool {
	return srcPort == "21" || dstPort == "21"
}

func (f *pcapStreamFactory) matchSMTP(srcPort, dstPort string, peek []byte) bool {
	ports := map[string]bool{"25": true, "587": true, "465": true}
	return ports[srcPort] || ports[dstPort]
}

func (f *pcapStreamFactory) matchTelnet(srcPort, dstPort string, peek []byte) bool {
	return srcPort == "23" || dstPort == "23" || (len(peek) >= 1 && peek[0] == 0xFF)
}

func (f *pcapStreamFactory) matchSMB(srcPort, dstPort string, peek []byte) bool {
	if srcPort == "445" || dstPort == "445" {
		return true
	}
	if len(peek) >= 4 && ((peek[0] == 0xFF && peek[1] == 0x53 && peek[2] == 0x4D && peek[3] == 0x42) ||
		(peek[0] == 0xFE && peek[1] == 0x53 && peek[2] == 0x4D && peek[3] == 0x42)) {
		return true
	}
	return false
}

func (f *pcapStreamFactory) matchRedis(srcPort, dstPort string, peek []byte) bool {
	return srcPort == "6379" || dstPort == "6379"
}

func (f *pcapStreamFactory) matchMySQL(srcPort, dstPort string, peek []byte) bool {
	return srcPort == "3306" || dstPort == "3306" || (len(peek) >= 1 && peek[0] == 0x0a)
}

func (f *pcapStreamFactory) matchMSSQL(srcPort, dstPort string, peek []byte) bool {
	return srcPort == "1433" || dstPort == "1433"
}

func (f *pcapStreamFactory) matchPOP3(srcPort, dstPort string, peek []byte) bool {
	return srcPort == "110" || dstPort == "110" || srcPort == "995" || dstPort == "995"
}

func (f *pcapStreamFactory) matchIMAP(srcPort, dstPort string, peek []byte) bool {
	return srcPort == "143" || dstPort == "143" || srcPort == "993" || dstPort == "993"
}

// 协议解析器 (简化版)
func (f *pcapStreamFactory) parseHTTP(netFlow, transFlow gopacket.Flow, reader io.Reader) {
	ts := time.Now()
	allData, _ := io.ReadAll(reader)
	if len(allData) == 0 {
		return
	}

	br := bytes.NewReader(allData)
	for br.Len() > 0 {
		peek := make([]byte, 10)
		n, _ := br.Read(peek)
		if n == 0 {
			break
		}
		br.Seek(-int64(n), 1)

		if strings.HasPrefix(string(peek), "HTTP/") {
			if resp, _ := http.ReadResponse(bufio.NewReader(br), nil); resp != nil {
				dump, _ := httputil.DumpResponse(resp, true)
				ev := f.newEvent(netFlow, transFlow, ts, "HTTP")
				ev.HTTPStatus = resp.StatusCode
				ev.Tags = "http_response"
				ev.Content = f.truncate(string(dump))
				f.eventChan <- ev
				resp.Body.Close()
			}
		} else {
			if req, _ := http.ReadRequest(bufio.NewReader(br)); req != nil {
				dump, _ := httputil.DumpRequest(req, true)
				ev := f.newEvent(netFlow, transFlow, ts, "HTTP")
				ev.HTTPMethod = req.Method
				ev.HTTPURL = req.URL.String()
				ev.HTTPHost = req.Host
				ev.Tags = "http_request"
				ev.Content = f.truncate(string(dump))
				f.eventChan <- ev
				req.Body.Close()
			}
		}
	}
}

func (f *pcapStreamFactory) parseFTP(netFlow, transFlow gopacket.Flow, reader io.Reader) {
	ts := time.Now()
	allData, _ := io.ReadAll(reader)
	if len(allData) == 0 {
		return
	}

	scanner := bufio.NewScanner(bytes.NewReader(allData))
	var user, pass string
	for scanner.Scan() {
		line := strings.ToUpper(scanner.Text())
		if strings.HasPrefix(line, "USER ") {
			user = strings.TrimSpace(scanner.Text()[5:])
		}
		if strings.HasPrefix(line, "PASS ") {
			pass = strings.TrimSpace(scanner.Text()[5:])
		}
	}

	ev := f.newEvent(netFlow, transFlow, ts, "FTP")
	ev.Tags = "ftp"
	if user != "" || pass != "" {
		ev.Tags = "ftp_auth"
		ev.Content = fmt.Sprintf("USER: %s | PASS: %s", user, f.truncate(pass))
	} else {
		ev.Content = f.truncate(string(allData))
	}
	f.eventChan <- ev
}

func (f *pcapStreamFactory) parseSMTP(netFlow, transFlow gopacket.Flow, reader io.Reader) {
	ts := time.Now()
	allData, _ := io.ReadAll(reader)
	if len(allData) == 0 {
		return
	}

	scanner := bufio.NewScanner(bytes.NewReader(allData))
	var from, to []string
	for scanner.Scan() {
		line := strings.ToUpper(scanner.Text())
		if strings.HasPrefix(line, "MAIL FROM:") {
			from = append(from, strings.TrimSpace(scanner.Text()[10:]))
		}
		if strings.HasPrefix(line, "RCPT TO:") {
			to = append(to, strings.TrimSpace(scanner.Text()[8:]))
		}
	}

	ev := f.newEvent(netFlow, transFlow, ts, "SMTP")
	if len(from) > 0 || len(to) > 0 {
		ev.Tags = "smtp_mail"
		ev.Content = fmt.Sprintf("From: %v | To: %v", from, to)
	} else {
		ev.Tags = "smtp"
		ev.Content = f.truncate(string(allData))
	}
	f.eventChan <- ev
}

func (f *pcapStreamFactory) parseTelnet(netFlow, transFlow gopacket.Flow, reader io.Reader) {
	ts := time.Now()
	allData, _ := io.ReadAll(reader)
	if len(allData) == 0 {
		return
	}

	var printable strings.Builder
	for _, b := range allData {
		if b >= 32 && b < 127 {
			printable.WriteByte(b)
		}
	}

	ev := f.newEvent(netFlow, transFlow, ts, "TELNET")
	ev.Content = f.truncate(printable.String())
	ev.Tags = "telnet"
	f.eventChan <- ev
}

func (f *pcapStreamFactory) parseSMB(netFlow, transFlow gopacket.Flow, reader io.Reader) {
	ts := time.Now()
	allData, _ := io.ReadAll(reader)
	if len(allData) == 0 {
		return
	}

	ev := f.newEvent(netFlow, transFlow, ts, "SMB")
	ev.Tags = "smb"
	content := strings.ToUpper(string(allData))
	if strings.Contains(content, "SESSION") && strings.Contains(content, "SETUP") {
		ev.Tags = "smb_auth"
	}
	ev.Content = f.truncate(string(allData))
	f.eventChan <- ev
}

func (f *pcapStreamFactory) parseRedis(netFlow, transFlow gopacket.Flow, reader io.Reader) {
	ts := time.Now()
	allData, _ := io.ReadAll(reader)
	if len(allData) == 0 {
		return
	}

	ev := f.newEvent(netFlow, transFlow, ts, "REDIS")
	ev.Tags = "redis"
	content := strings.ToUpper(string(allData))
	if strings.Contains(content, "AUTH") {
		ev.Tags = "redis_auth"
	}
	ev.Content = f.truncate(string(allData))
	f.eventChan <- ev
}

func (f *pcapStreamFactory) parseMySQL(netFlow, transFlow gopacket.Flow, reader io.Reader) {
	ts := time.Now()
	allData, _ := io.ReadAll(reader)
	if len(allData) == 0 {
		return
	}

	ev := f.newEvent(netFlow, transFlow, ts, "MYSQL")
	ev.Tags = "mysql"
	sqlRe := regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE|FROM|WHERE|USER|PASSWORD)`)
	matches := sqlRe.FindAllString(string(allData), 5)
	if len(matches) > 0 {
		ev.Tags = "mysql_query"
		ev.Content = fmt.Sprintf("SQL: %v", matches)
	} else {
		ev.Content = f.truncate(string(allData))
	}
	f.eventChan <- ev
}

func (f *pcapStreamFactory) parseMSSQL(netFlow, transFlow gopacket.Flow, reader io.Reader) {
	ts := time.Now()
	allData, _ := io.ReadAll(reader)
	if len(allData) == 0 {
		return
	}

	ev := f.newEvent(netFlow, transFlow, ts, "MSSQL")
	ev.Tags = "mssql"
	content := strings.ToUpper(string(allData))
	if strings.Contains(content, "XP_CMDSHELL") || strings.Contains(content, "EXEC(") {
		ev.Tags = "mssql_exec"
	}
	ev.Content = f.truncate(string(allData))
	f.eventChan <- ev
}

func (f *pcapStreamFactory) parsePOP3(netFlow, transFlow gopacket.Flow, reader io.Reader) {
	ts := time.Now()
	allData, _ := io.ReadAll(reader)
	if len(allData) == 0 {
		return
	}

	ev := f.newEvent(netFlow, transFlow, ts, "POP3")
	scanner := bufio.NewScanner(bytes.NewReader(allData))
	var user, pass string
	for scanner.Scan() {
		line := strings.ToUpper(scanner.Text())
		if strings.HasPrefix(line, "USER ") {
			user = strings.TrimSpace(scanner.Text()[5:])
		}
		if strings.HasPrefix(line, "PASS ") {
			pass = strings.TrimSpace(scanner.Text()[5:])
		}
	}
	if user != "" || pass != "" {
		ev.Tags = "pop3_auth"
		ev.Content = fmt.Sprintf("USER: %s | PASS: %s", user, pass)
	} else {
		ev.Tags = "pop3"
		ev.Content = f.truncate(string(allData))
	}
	f.eventChan <- ev
}

func (f *pcapStreamFactory) parseIMAP(netFlow, transFlow gopacket.Flow, reader io.Reader) {
	ts := time.Now()
	allData, _ := io.ReadAll(reader)
	if len(allData) == 0 {
		return
	}

	ev := f.newEvent(netFlow, transFlow, ts, "IMAP")
	content := strings.ToUpper(string(allData))
	if strings.Contains(content, "LOGIN") {
		ev.Tags = "imap_auth"
	} else {
		ev.Tags = "imap"
	}
	ev.Content = f.truncate(string(allData))
	f.eventChan <- ev
}

func (f *pcapStreamFactory) parseRawTCP(netFlow, transFlow gopacket.Flow, reader io.Reader) {
	ts := time.Now()
	allData, _ := io.ReadAll(reader)
	if len(allData) == 0 {
		return
	}

	ev := f.newEvent(netFlow, transFlow, ts, "TCP_STREAM")
	ev.Tags = "tcp_stream"
	ev.Content = f.truncate(string(allData))
	f.eventChan <- ev
}

func (f *pcapStreamFactory) newEvent(netFlow, transFlow gopacket.Flow, ts time.Time, protocol string) ForensicEventJSON {
	return ForensicEventJSON{
		Timestamp:    float64(ts.UnixNano()) / 1e9,
		TimestampISO: ts.Format(time.RFC3339Nano),
		SrcIP:        netFlow.Src().String(),
		DstIP:        netFlow.Dst().String(),
		SrcPort:      transFlow.Src().String(),
		DstPort:      transFlow.Dst().String(),
		Network:      netFlow.EndpointType().String(),
		Transport:    "TCP",
		Protocol:     protocol,
	}
}

func (f *pcapStreamFactory) truncate(s string) string {
	if f.contentMaxLen > 0 && len(s) > f.contentMaxLen {
		return s[:f.contentMaxLen]
	}
	return s
}

func buildTCPFlags(tcp *layers.TCP) string {
	flags := ""
	if tcp.SYN {
		flags = "SYN"
	}
	if tcp.ACK {
		if flags != "" {
			flags += "+"
		}
		flags += "ACK"
	}
	if tcp.FIN {
		if flags != "" {
			flags += "+"
		}
		flags += "FIN"
	}
	if tcp.RST {
		if flags != "" {
			flags += "+"
		}
		flags += "RST"
	}
	if tcp.PSH {
		if flags != "" {
			flags += "+"
		}
		flags += "PSH"
	}
	return flags
}