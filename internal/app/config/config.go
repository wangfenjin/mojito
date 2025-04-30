// Package config provides configuration loading and management for the application
package config

import (
	"fmt"
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
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
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

	// Read configuration from file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Look for config in the working directory
		v.AddConfigPath("./config")
		v.SetConfigName("config")
	}

	// Read environment variables
	v.SetEnvPrefix("MOJITO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal config
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}
