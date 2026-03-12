package ocsf

//   ==============================================================================
// OCSF - (Objects) 定义
//   ==============================================================================

type Metadata struct {
	Version  string `json:"version,omitempty"`
	Product  string `json:"product,omitempty"`
	Profiles string `json:"profiles,omitempty"`
}

type Device struct {
	Hostname string `json:"hostname,omitempty"`
	IP       string `json:"ip,omitempty"`
	MAC      string `json:"mac,omitempty"`
	Vendor   string `json:"vendor,omitempty"`
	OS       *OS    `json:"os,omitempty"`
}

type OS struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	Type    string `json:"type,omitempty"`
}

type Endpoint struct {
	IP       string `json:"ip,omitempty"`
	Port     int    `json:"port,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	MAC      string `json:"mac,omitempty"`
	Domain   string `json:"domain,omitempty"`
}

type User struct {
	Name   string `json:"name,omitempty"`
	UID    string `json:"uid,omitempty"` // 可存放 - Domain string `json:"domain,omitempty"`
	Type   string `json:"type,omitempty"`
	Group  string `json:"group_name,omitempty"`
}

type Process struct {
	Name    string   `json:"name,omitempty"`
	PID     int      `json:"pid,omitempty"`
	CmdLine string   `json:"cmd_line,omitempty"`
	Path    string   `json:"file.path,omitempty"`
	UID     string   `json:"uid,omitempty"`
	Parent  *Process `json:"parent_process,omitempty"`
}

type File struct {
	Name   string `json:"name,omitempty"`
	Path   string `json:"path,omitempty"`
	Size   int64  `json:"size,omitempty"`
	Type   string `json:"type,omitempty"`
	Hashes *Hash  `json:"hashes,omitempty"`
}

type Hash struct {
	MD5    string `json:"md5,omitempty"`
	SHA1   string `json:"sha1,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
}

//   【New增】Windows 注册表对象
type Registry struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
	Data  string `json:"data,omitempty"`
}

//   【New增】Windows Service与PlanTask对象
type Service struct {
	Name      string `json:"name,omitempty"`
	State     string `json:"state,omitempty"`
	StartType string `json:"start_type,omitempty"`
	CmdLine   string `json:"cmd_line,omitempty"`
}

//   【New增】恶意软件与威胁发现对象 (对接 Defender/AV)
type Malware struct {
	Name           string   `json:"name,omitempty"`
	Classification string   `json:"classification,omitempty"`
	Path           string   `json:"path,omitempty"`
	Cves           []string `json:"cves,omitempty"`
}
