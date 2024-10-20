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
	ObjStorage *objectStorage
}

type objectStorage struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
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
		ObjStorage: &objectStorage{
			Endpoint:  viper.GetString("s3-endpoint"),
			AccessKey: viper.GetString("s3-access-key"),
			SecretKey: viper.GetString("s3-secret-key"),
			Bucket:    viper.GetString("s3-bucket"),
			UseSSL:    viper.GetBool("s3-use-ssl"),
		},
	}

	return cfg, nil
}

func initConfig() error {
	pflag.StringP("config", "c", "", "path to config file")
	pflag.StringP("address", "a", "", "server address")
	pflag.StringP("log-level", "l", "", "log level")
	pflag.String("jwt-secret", "", "JWT secret used for token generation and verification")
	pflag.String("crypto-key", "", "crypto key used for data encryption and decryption")
	pflag.String("s3-endpoint", "", "S3 object storage endpoint")
	pflag.String("s3-access-key", "", "S3 object storage access key")
	pflag.String("s3-secret-key", "", "S3 object storage secret key")
	pflag.String("s3-bucket", "", "S3 object storage bucket")
	pflag.Bool("s3-use-ssl", false, "S3 object storage use SSL")
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
	viper.SetDefault("log-level", "debug")
	viper.SetDefault("jwt-secret", "topsecretkey")
	viper.SetDefault("crypto-key", "123456789abcdefg")
	viper.SetDefault("s3-endpoint", "localhost:9000")
	viper.SetDefault("s3-bucket", "vault")
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

	if err := viper.BindEnv("s3-endpoint", "KEEPER_S3_ENDPOINT"); err != nil {
		return fmt.Errorf("viper.BindEnv: %w", err)
	}

	if err := viper.BindEnv("s3-access-key", "KEEPER_S3_ACCESS_KEY"); err != nil {
		return fmt.Errorf("viper.BindEnv: %w", err)
	}

	if err := viper.BindEnv("s3-secret-key", "KEEPER_S3_SECRET_KEY"); err != nil {
		return fmt.Errorf("viper.BindEnv: %w", err)
	}

	if err := viper.BindEnv("s3-bucket", "KEEPER_S3_BUCKET"); err != nil {
		return fmt.Errorf("viper.BindEnv: %w", err)
	}

	if err := viper.BindEnv("s3-use-ssl", "KEEPER_S3_USE_SSL"); err != nil {
		return fmt.Errorf("viper.BindEnv: %w", err)
	}

	return nil
}
