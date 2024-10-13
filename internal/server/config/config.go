// Package config provides the server config.
package config

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	ServerAddr string
	LogLevel   string
	JWTSecret  string
	CryptoKey  string
}

func NewConfig() (*Config, error) {
	if err := initConfig(); err != nil {
		return nil, fmt.Errorf("initConfig: %w", err)
	}

	cfg := &Config{
		ServerAddr: viper.GetString("address"),
		LogLevel:   viper.GetString("log-level"),
		JWTSecret:  viper.GetString("jwt-secret"),
		CryptoKey:  viper.GetString("crypto-key"),
	}

	return cfg, nil
}

func initConfig() error {
	pflag.StringP("config", "c", "", "path to config file")
	pflag.StringP("address", "a", "", "server address")
	pflag.StringP("log-level", "l", "", "log level")
	pflag.String("jwt-secret", "", "JWT secret used for token generation and verification")
	pflag.String("crypto-key", "", "crypto key used for data encryption and decryption")
	pflag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		return fmt.Errorf("viper.BindPFlags: %w", err)
	}

	if err := bindEnvs(); err != nil {
		return fmt.Errorf("failed to bind environment variables: %w", err)
	}

	setDefaults()

	if cfgFile := viper.GetString("config"); cfgFile != "" {
		if err := readConfigFile(cfgFile); err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return nil
}

func readConfigFile(cfgFile string) error {
	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("viper.ReadInConfig: %w", err)
	}

	return nil
}

func setDefaults() {
	viper.SetDefault("address", ":8080")
	viper.SetDefault("log-level", "info")
	viper.SetDefault("jwt-secret", "topsecretkey")
	viper.SetDefault("crypto-key", "123456789abcdefg")
}

func bindEnvs() error {
	if err := viper.BindEnv("config", "KEEPER_CONFIG"); err != nil {
		return fmt.Errorf("viper.BindEnv: %w", err)
	}

	if err := viper.BindEnv("address", "KEEPER_ADDR"); err != nil {
		return fmt.Errorf("viper.BindEnv: %w", err)
	}

	if err := viper.BindEnv("log-level", "KEEPER_LOG_LEVEL"); err != nil {
		return fmt.Errorf("viper.BindEnv: %w", err)
	}

	if err := viper.BindEnv("jwt-secret", "KEEPER_JWT_SECRET"); err != nil {
		return fmt.Errorf("viper.BindEnv: %w", err)
	}

	if err := viper.BindEnv("crypto-key", "KEEPER_CRYPTO_KEY"); err != nil {
		return fmt.Errorf("viper.BindEnv: %w", err)
	}

	return nil
}
