package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Email    EmailConfig
	Logging  LoggingConfig
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Host            string
	Port            int
	BasePath        string
	AllowedOrigins  []string
	ShutdownTimeout int
}

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	Type       string // "postgres" or "sqlite"
	Host       string
	Port       int
	User       string
	Password   string
	Name       string
	SSLMode    string
	TimeZone   string
	SQLitePath string // Path for SQLite database file
}

// AuthConfig holds all authentication-related configuration
type AuthConfig struct {
	SecretKey            string
	AccessTokenExpire    int
	RefreshTokenExpire   int
	PasswordResetExpire  int
	VerificationExpire   int
	PasswordMinLength    int
	PasswordHashCost     int
	FirstSuperuserEmail  string
	FirstSuperuserPasswd string
}

// EmailConfig holds all email-related configuration
type EmailConfig struct {
	Enabled    bool
	SMTPHost   string
	SMTPPort   int
	SMTPUser   string
	SMTPPasswd string
	FromEmail  string
	FromName   string
}

// LoggingConfig holds all logging-related configuration
type LoggingConfig struct {
	Level string
	File  string
}

// Load loads the configuration from files and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default configuration values
	setDefaults(v)

	// Read configuration from file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Look for config in the working directory
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
		v.SetConfigName("config")
	}

	// Read environment variables
	v.SetEnvPrefix("MOJITO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, using defaults and environment variables
	}

	// Create config directory if it doesn't exist
	if configPath == "" {
		if err := os.MkdirAll("config", 0755); err != nil {
			return nil, fmt.Errorf("error creating config directory: %w", err)
		}
		// Write default config file if it doesn't exist
		if _, err := os.Stat(filepath.Join("config", "config.yaml")); os.IsNotExist(err) {
			if err := v.WriteConfigAs(filepath.Join("config", "config.yaml")); err != nil {
				return nil, fmt.Errorf("error writing default config file: %w", err)
			}
		}
	}

	// Unmarshal config
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.basePath", "/api/v1")
	v.SetDefault("server.allowedOrigins", []string{"http://localhost:3000"})
	v.SetDefault("server.shutdownTimeout", 5)

	// Database defaults
	v.SetDefault("database.type", "postgres") // Default to postgres
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "mojito")
	v.SetDefault("database.sslMode", "disable")
	v.SetDefault("database.timeZone", "UTC")
	v.SetDefault("database.sqlitePath", "./mojito.db") // Default SQLite path

	// Auth defaults
	v.SetDefault("auth.secretKey", "supersecretkey")
	v.SetDefault("auth.accessTokenExpire", 30)   // minutes
	v.SetDefault("auth.refreshTokenExpire", 7)   // days
	v.SetDefault("auth.passwordResetExpire", 24) // hours
	v.SetDefault("auth.verificationExpire", 48)  // hours
	v.SetDefault("auth.passwordMinLength", 8)
	v.SetDefault("auth.passwordHashCost", 10)
	v.SetDefault("auth.firstSuperuserEmail", "admin@example.com")
	v.SetDefault("auth.firstSuperuserPasswd", "admin")

	// Email defaults
	v.SetDefault("email.enabled", false)
	v.SetDefault("email.smtpHost", "smtp.example.com")
	v.SetDefault("email.smtpPort", 587)
	v.SetDefault("email.smtpUser", "user")
	v.SetDefault("email.smtpPasswd", "password")
	v.SetDefault("email.fromEmail", "noreply@example.com")
	v.SetDefault("email.fromName", "Mojito App")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.file", "")
}
