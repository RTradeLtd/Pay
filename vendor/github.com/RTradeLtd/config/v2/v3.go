package config

// V3 contains v3-specific configuration
type V3 struct {
	API     V3API
	Gateway V3Gateway
}

// V3API configures the V3 gRPC API
type V3API struct {
	VerifyDomain string
	Address      string
	JWT          V3JWT
	TLS          V3TLS
}

// V3Gateway configures the V3 API gateway
type V3Gateway struct {
	Address       string
	TargetAddress string
	TLS           V3TLS
}

// V3JWT configures JWT settings
type V3JWT struct {
	Key     string
	Realm   string
	Timeout int // hours
}

// V3TLS configures TLS options
type V3TLS struct {
	CertPath string `json:"cert_path"`
	KeyPath  string `json:"key_path"`
}
