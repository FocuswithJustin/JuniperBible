package api

// Config holds server configuration.
type Config struct {
	Port              int
	CapsulesDir       string
	PluginsDir        string
	PluginsExternal   bool
	RateLimitRequests int        // Requests per minute (0 = disabled)
	RateLimitBurst    int        // Burst size
	Auth              AuthConfig // Authentication configuration
	TLS               TLSConfig  // TLS configuration
	AllowedOrigins    []string   // CORS allowed origins (empty = allow all)
}

// TLSConfig holds TLS/HTTPS configuration.
type TLSConfig struct {
	Enabled  bool   // Enable HTTPS
	CertFile string // Path to TLS certificate file
	KeyFile  string // Path to TLS private key file
}

// ServerConfig is the active server configuration.
var ServerConfig Config
