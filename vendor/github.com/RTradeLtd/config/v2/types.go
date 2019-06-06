package config

// TemporalConfig defines Temporal configuration fields
type TemporalConfig struct {
	V3          `json:"v3"`
	API         `json:"api,omitempty"`
	APIKeys     `json:"api_keys,omitempty"`
	AWS         `json:"aws,omitempty"`
	Database    `json:"database,omitempty"`
	Services    `json:"services,omitempty"`
	Ethereum    `json:"ethereum,omitempty"`
	IPFSCluster `json:"ipfs_cluster,omitempty"`
	IPFS        `json:"ipfs,omitempty"`
	Pay         `json:"pay,omitempty"`
	RabbitMQ    `json:"rabbitmq,omitempty"`
	Sendgrid    `json:"sendgrid,omitempty"`
	Stripe      `json:"stripe,omitempty"`
	Wallets     `json:"wallets,omitempty"`
	LogDir      string `json:"log_dir,omitempty"`
}

// API configures the Temporal API
type API struct {
	Connection struct {
		Certificates struct {
			CertPath string `json:"cert_path"`
			KeyPath  string `json:"key_path"`
		} `json:"certificates"`
		ListenAddress string `json:"listen_address"`
		// defines parameters for prometheus metric collector
		Prometheus struct {
			IP   string `json:"ip"`
			Port string `json:"port"`
		} `json:"prometheus"`
		CORS struct {
			AllowedOrigins []string `json:"allowed_origins"`
		} `json:"cors"`
		// define the maximum number of people allowed to connect to the API
		Limit string `json:"limit"`
	} `json:"connection"`
	JWT struct {
		Key   string `json:"key"`
		Realm string `json:"realm"`
	} `json:"jwt"`
	SizeLimitInGigaBytes string `json:"size_limit_in_giga_bytes"`
}

// Pay configures connection to our payment processor
type Pay struct {
	Address  string `json:"address"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
	TLS      struct {
		CertPath string `json:"cert"`
		KeyPath  string `json:"key"`
	} `json:"tls"`
	AuthKey string `json:"auth_key"`
}

// Database configures Temporal's connection to a Postgres database
type Database struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// IPFS configures Temporal's connection to an IPFS node
type IPFS struct {
	APIConnection struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"api_connection"`
	KeystorePath string `json:"keystore_path"`
}

// IPFSCluster configures Temporal's connection to an IPFS cluster
type IPFSCluster struct {
	APIConnection struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"api_connection"`
}

// RabbitMQ configures Temporal's connection to a RabbitMQ instance
type RabbitMQ struct {
	URL       string `json:"url"`
	TLSConfig struct {
		CertFile   string `json:"cert_file"`
		KeyFile    string `json:"key_file"`
		CACertFile string `json:"ca_cert_file"`
	} `json:"tls_config"`
}

// AWS configures Temporal's connection to AWS
type AWS struct {
	KeyID  string `json:"key_id"`
	Secret string `json:"secret"`
}

// Sendgrid configures Temporal's connection to Sendgrid
type Sendgrid struct {
	APIKey       string `json:"api_key"`
	EmailAddress string `json:"email_address"`
	EmailName    string `json:"email_name"`
}

// Ethereum configures Temporal's connection, and interaction with the Ethereum blockchain
type Ethereum struct {
	Account struct {
		Address string `json:"address"`
		KeyFile string `json:"key_file"`
		KeyPass string `json:"key_pass"`
	} `json:"account"`
	Connection struct {
		RPC struct {
			IP   string `json:"ip"`
			Port string `json:"port"`
		} `json:"rpc"`
		IPC struct {
			Path string `json:"path"`
		} `json:"ipc"`
		INFURA struct {
			URL string `json:"url"`
		} `json:"infura"`
	} `json:"connection"`
	Contracts struct {
		RTCAddress             string `json:"rtc_address"`
		PaymentContractAddress string `json:"payment_contract_address"`
	} `json:"contracts"`
}

// Wallets are the addresses of RTrade Ltd's wallets
type Wallets struct {
	ETH  string `json:"eth"`
	RTC  string `json:"rtc"`
	XMR  string `json:"xmr"`
	DASH string `json:"dash"`
	BTC  string `json:"btc"`
	LTC  string `json:"ltc"`
}

// APIKeys are the various API keys we use
type APIKeys struct {
	ChainRider string `json:"chain_rider"`
}

// Services are various endpoints we connect to
type Services struct {
	MoneroRPC string `json:"monero_rpc"`
	Lens      `json:"lens"`
	Nexus     `json:"nexus"`
	MongoDB   struct {
		URL              string `json:"url"`
		DB               string `json:"db"`
		UploadCollection string `json:"uploads"`
	} `json:"mongodb"`
	Raven struct {
		URL  string `json:"url"`
		User string `json:"user"`
		Pass string `json:"pass"`
	} `json:"raven"`
	BchGRPC struct {
		URL      string `json:"url"`
		CertFile string `json:"cert_file"`
		KeyFile  string `json:"key_file"`
		// Wallet defines connection information
		// to the bchwallet gRPC service
		Wallet struct {
			URL      string `json:"url"`
			CertFile string `json:"cert_file"`
			KeyFile  string `json:"key_file"`
		} `json:"wallet"`
	} `json:"bch_grpc"`
	Krab         `json:"krab"`
	KrabFallback Krab `json:"krab_fallback"`
	RTNS         `json:"rtns"`
}

// Krab is used to for key management
type Krab struct {
	URL string `json:"url"`
	TLS struct {
		CertPath string `json:"cert_path"`
		KeyFile  string `json:"key_file"`
	}
	AuthKey          string `json:"auth_key"`
	LogFile          string `json:"log_file"`
	KeystorePassword string `json:"keystore_password"`
}

// KrabFallback is a fallback configuration for
// connecting to a secondary krab server
type KrabFallback Krab

// Lens defines options for the Lens search engine
type Lens struct {
	URL string `json:"url"`
	TLS struct {
		CertPath string `json:"cert_path"`
		KeyFile  string `json:"key_file"`
	} `json:"tls"`
	AuthKey string `json:"auth_key"`
	Options struct {
		Engine struct {
			StorePath string `json:"store_path"`
			Queue     struct {
				Rate  int `json:"rate"`
				Batch int `json:"batch"`
			} `json:"queue"`
		} `json:"engine"`
	} `json:"options"`
}

// Nexus defines options for the Nexus, our private network
// management tool for IPFS.
type Nexus struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Key  string `json:"key"`
	TLS  struct {
		CertPath string `json:"cert"`
		KeyPath  string `json:"key"`
	} `json:"tls"`
	Delegator struct {
		Port string `json:"port"`
	} `json:"delegator"`
}

// Stripe is used to configure our connection with stripe api
type Stripe struct {
	PublishableKey string `json:"publishable_key"`
	SecretKey      string `json:"secret_key"`
}

// RTNS is used to configure our RTNS publishing service
type RTNS struct {
	MultiAddresses []string `json:"multi_addresses,omitempty"`
	// name of the private key stored within krab
	KeyName string `json:"pk_name,omitempty"`
	// path to persistent data store
	DatastorePath string `json:"datastore_path"`
}
