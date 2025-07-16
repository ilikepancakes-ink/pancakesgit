package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Security   SecurityConfig   `mapstructure:"security"`
	Git        GitConfig        `mapstructure:"git"`
	Privacy    PrivacyConfig    `mapstructure:"privacy"`
	Storage    StorageConfig    `mapstructure:"storage"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	Domain       string `mapstructure:"domain"`
	RootURL      string `mapstructure:"root_url"`
	Debug        bool   `mapstructure:"debug"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Type     string `mapstructure:"type"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
	Path     string `mapstructure:"path"` // For SQLite
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	SecretKey            string `mapstructure:"secret_key"`
	SessionTimeout       int    `mapstructure:"session_timeout"`
	PasswordMinLength    int    `mapstructure:"password_min_length"`
	RequireStrongPasswd  bool   `mapstructure:"require_strong_password"`
	Enable2FA            bool   `mapstructure:"enable_2fa"`
	MaxLoginAttempts     int    `mapstructure:"max_login_attempts"`
	LockoutDuration      int    `mapstructure:"lockout_duration"`
	CSRFProtection       bool   `mapstructure:"csrf_protection"`
	SecureCookies        bool   `mapstructure:"secure_cookies"`
	HTTPSOnly            bool   `mapstructure:"https_only"`
}

// GitConfig contains Git-related configuration
type GitConfig struct {
	EnableHTTP     bool   `mapstructure:"enable_http"`
	EnableSSH      bool   `mapstructure:"enable_ssh"`
	SSHPort        int    `mapstructure:"ssh_port"`
	SSHHostKey     string `mapstructure:"ssh_host_key"`
	MaxRepoSize    int64  `mapstructure:"max_repo_size"`
	AllowedFormats []string `mapstructure:"allowed_formats"`
}

// PrivacyConfig contains privacy-focused settings
type PrivacyConfig struct {
	AnonymousAccess      bool   `mapstructure:"anonymous_access"`
	HideUserEmails       bool   `mapstructure:"hide_user_emails"`
	DisableGravatar      bool   `mapstructure:"disable_gravatar"`
	LogRetentionDays     int    `mapstructure:"log_retention_days"`
	DataRetentionDays    int    `mapstructure:"data_retention_days"`
	EnableTorAccess      bool   `mapstructure:"enable_tor_access"`
	DisableAnalytics     bool   `mapstructure:"disable_analytics"`
	MinimizeMetadata     bool   `mapstructure:"minimize_metadata"`
}

// StorageConfig contains storage configuration
type StorageConfig struct {
	Type           string `mapstructure:"type"` // local, s3, etc.
	Path           string `mapstructure:"path"`
	EncryptAtRest  bool   `mapstructure:"encrypt_at_rest"`
	CompressionLevel int  `mapstructure:"compression_level"`
}

// EncryptionConfig contains encryption settings
type EncryptionConfig struct {
	Key                string `mapstructure:"key"`
	Algorithm          string `mapstructure:"algorithm"`
	EncryptRepositories bool   `mapstructure:"encrypt_repositories"`
	EncryptDatabase     bool   `mapstructure:"encrypt_database"`
	EncryptLogs         bool   `mapstructure:"encrypt_logs"`
}

// Load loads configuration from various sources
func Load(cmd *cobra.Command) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Set config name and paths
	v.SetConfigName("app")
	v.SetConfigType("ini")
	
	// Add config paths
	v.AddConfigPath("/config")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")
	
	// Check for custom config file
	if configFile, _ := cmd.Flags().GetString("config"); configFile != "" {
		v.SetConfigFile(configFile)
	}

	// Environment variables
	v.SetEnvPrefix("PANCAKESGIT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Override with command line flags
	if host, _ := cmd.Flags().GetString("host"); host != "" {
		v.Set("server.host", host)
	}
	if port, _ := cmd.Flags().GetString("port"); port != "" {
		v.Set("server.port", port)
	}
	if debug, _ := cmd.Flags().GetBool("debug"); debug {
		v.Set("server.debug", debug)
	}

	// Unmarshal config
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate and set derived values
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", "3000")
	v.SetDefault("server.domain", "localhost")
	v.SetDefault("server.root_url", "http://localhost:3000")
	v.SetDefault("server.debug", false)
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)

	// Database defaults
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "pancakesgit")
	v.SetDefault("database.user", "pancakesgit")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.path", "/data/pancakesgit.db")

	// Security defaults
	v.SetDefault("security.session_timeout", 3600)
	v.SetDefault("security.password_min_length", 8)
	v.SetDefault("security.require_strong_password", true)
	v.SetDefault("security.enable_2fa", true)
	v.SetDefault("security.max_login_attempts", 5)
	v.SetDefault("security.lockout_duration", 900)
	v.SetDefault("security.csrf_protection", true)
	v.SetDefault("security.secure_cookies", true)
	v.SetDefault("security.https_only", false)

	// Git defaults
	v.SetDefault("git.enable_http", true)
	v.SetDefault("git.enable_ssh", true)
	v.SetDefault("git.ssh_port", 2222)
	v.SetDefault("git.max_repo_size", 1073741824) // 1GB
	v.SetDefault("git.allowed_formats", []string{"git"})

	// Privacy defaults
	v.SetDefault("privacy.anonymous_access", false)
	v.SetDefault("privacy.hide_user_emails", true)
	v.SetDefault("privacy.disable_gravatar", true)
	v.SetDefault("privacy.log_retention_days", 30)
	v.SetDefault("privacy.data_retention_days", 365)
	v.SetDefault("privacy.enable_tor_access", true)
	v.SetDefault("privacy.disable_analytics", true)
	v.SetDefault("privacy.minimize_metadata", true)

	// Storage defaults
	v.SetDefault("storage.type", "local")
	v.SetDefault("storage.path", "/data/repositories")
	v.SetDefault("storage.encrypt_at_rest", true)
	v.SetDefault("storage.compression_level", 6)

	// Encryption defaults
	v.SetDefault("encryption.algorithm", "AES-256-GCM")
	v.SetDefault("encryption.encrypt_repositories", true)
	v.SetDefault("encryption.encrypt_database", true)
	v.SetDefault("encryption.encrypt_logs", true)
}

func validateConfig(config *Config) error {
	// Validate required fields
	if config.Security.SecretKey == "" {
		return fmt.Errorf("security.secret_key is required")
	}

	if config.Encryption.Key == "" {
		return fmt.Errorf("encryption.key is required")
	}

	// Set derived values
	if config.Server.RootURL == "" {
		if config.Server.Domain != "" {
			config.Server.RootURL = fmt.Sprintf("https://%s", config.Server.Domain)
		} else {
			config.Server.RootURL = fmt.Sprintf("http://%s:%s", config.Server.Host, config.Server.Port)
		}
	}

	// Create necessary directories
	if config.Storage.Type == "local" {
		if err := os.MkdirAll(config.Storage.Path, 0755); err != nil {
			return fmt.Errorf("failed to create storage directory: %w", err)
		}
	}

	if config.Database.Type == "sqlite" {
		dir := filepath.Dir(config.Database.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	return nil
}
