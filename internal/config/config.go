package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Credential holds a username/password pair.
type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Config holds all server configuration.
type Config struct {
	// Address is the TCP address to listen on, e.g. "0.0.0.0:1080".
	Address string `json:"address"`

	// LogLevel controls verbosity: DEBUG, INFO, WARN, ERROR.
	LogLevel string `json:"log_level"`

	// Credentials is the list of allowed username/password pairs.
	// When empty, unauthenticated connections are rejected unless
	// AllowNoAuth is explicitly true (not recommended for production).
	Credentials []Credential `json:"credentials"`

	// AllowNoAuth disables authentication entirely (NOT recommended).
	AllowNoAuth bool `json:"allow_no_auth"`

	// DialTimeout limits how long the proxy waits to connect to the
	// target host.
	DialTimeout time.Duration `json:"dial_timeout"`

	// ReadTimeout is the per-read deadline applied to client connections.
	ReadTimeout time.Duration `json:"read_timeout"`

	// MaxConnections caps simultaneous client connections (0 = unlimited).
	MaxConnections int `json:"max_connections"`

	// AllowedCIDRs restricts which source IPs may connect.
	// An empty slice allows all source addresses.
	AllowedCIDRs []string `json:"allowed_cidrs"`

	// BlockedHosts is a list of destination hostnames/IPs to reject.
	BlockedHosts []string `json:"blocked_hosts"`
}

// Defaults returns a safe, ready-to-use configuration.
func Defaults() *Config {
	return &Config{
		Address:        "0.0.0.0:1080",
		LogLevel:       "INFO",
		DialTimeout:    30 * time.Second,
		ReadTimeout:    60 * time.Second,
		MaxConnections: 1000,
	}
}

func Load(path string) (*Config, error) {
	cfg := Defaults()

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	err = dec.Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	err = cfg.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}

// Validate checks that the configuration is internally consistent.
func (c *Config) Validate() error {
	if c.Address == "" {
		return fmt.Errorf("address must not be empty")
	}
	if !c.AllowNoAuth && len(c.Credentials) == 0 {
		return fmt.Errorf("credentials must not be empty when allow_no_auth is false")
	}
	for i, cr := range c.Credentials {
		if cr.Username == "" {
			return fmt.Errorf("credential[%d]: username must not be empty", i)
		}
		if cr.Password == "" {
			return fmt.Errorf("credential[%d]: password must not be empty", i)
		}
	}
	if c.MaxConnections < 0 {
		return fmt.Errorf("max_connections must be >= 0")
	}
	return nil
}
