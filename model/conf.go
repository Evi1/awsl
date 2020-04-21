package model

// HTTPIn HttpIn
type HTTPIn struct {
	Chan int
}

// AwslIn awsl
type AwslIn struct {
	Key  string
	Cert string
	URI  string
	Auth string
	Chan int
}

// TCPIn TCPIn
type TCPIn struct {
	Auth string
}

// AwslOut awsl
type AwslOut struct {
	Host   string
	Port   string
	URI    string
	Auth   string
	BackUp []string
}

// TCPOut TCPOut
type TCPOut struct {
	Host   string
	Port   string
	Auth   string
	BackUp []string
}

// In in
type In struct {
	Host string
	Port string
	Awsl *AwslIn
	HTTP *HTTPIn
	TCP  *TCPIn
	Type string
	Tag  string
}

// Out out
type Out struct {
	Type string
	Awsl *AwslOut
	TCP  *TCPOut
	Tag  string
}

// Object object
type Object struct {
	Ins        []In
	Outs       []Out
	BufSize    int
	NoVerify   bool
	Data       map[string]DataFile
	RouteRules []RouteRule
}

// DataFile DataFile
type DataFile struct {
	Name string
	Type int
}

// RouteRule route rule
type RouteRule struct {
	InTags   []string
	OutTags  []string
	DataTags []string
}
