package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	rejectedAuthConfigValues = map[string]struct{}{
		"your-google-client-id":             {},
		"your-google-client-secret":         {},
		"replace-with-a-long-random-secret": {},
	}
)

type AppConfig struct {
	AppPort         int      `yaml:"app_port"`
	EnableTelemetry bool     `yaml:"enable_telemetry"`
	MigratesFolder  string   `yaml:"migrates_folder"`
	ConfigDB        DBConf   `yaml:"conf_db"`
	Auth            AuthConf `yaml:"auth"`
}

type AuthConf struct {
	UIBaseURL     string           `yaml:"ui_base_url"`
	SecureCookies bool             `yaml:"secure_cookies"`
	Google        GoogleOAuth2Conf `yaml:"google"`
	Token         TokenSigningConf `yaml:"token"`
}

type GoogleOAuth2Conf struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"` //nolint:gosec // ok
	RedirectURL  string `yaml:"redirect_url"`
}

type TokenSigningConf struct {
	SigningKey        string        `yaml:"signing_key"`
	ExpirationMinutes time.Duration `yaml:"expiration_minutes"`
}

type DBConf struct {
	Address        string `yaml:"address"`
	Port           string `yaml:"port"`
	User           string `yaml:"user"`
	Pass           string `yaml:"pass"`
	DBName         string `yaml:"db_name"`
	MaxConnections int    `yaml:"max_connections"`
}

func InitConf(confFile string) (*AppConfig, error) {
	file, err := os.Open(filepath.Clean(confFile))
	if err != nil {
		return nil, fmt.Errorf("error open config file: %w", err)
	}
	defer func() {
		if e := file.Close(); e != nil {
			log.Fatal("Error close config file", e)
		}
	}()

	var cfg AppConfig
	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("error decode config file: %w", err)
	}
	if err = cfg.Auth.Validate(); err != nil {
		return nil, fmt.Errorf("error validate config file: %w", err)
	}
	return &cfg, nil
}

func (cfg *AuthConf) Validate() error {
	verifyItems := map[string]string{
		cfg.Google.ClientID:     "google-client-id",
		cfg.Google.ClientSecret: "google-client-secret",
		cfg.Google.RedirectURL:  "google-redirect-url",
		cfg.Token.SigningKey:    "token-signing-key",
	}
	for value, key := range verifyItems {
		if !hasRealConfigValue(value) {
			return fmt.Errorf("config value for %s is missing", key)
		}
	}
	if !hasValidAbsoluteURL(cfg.UIBaseURL) {
		return fmt.Errorf("config value for ui_base_url is missing")
	}
	if cfg.Token.ExpirationMinutes <= 0 {
		return fmt.Errorf("config value for token_expiration_minutes is missing")
	}
	return nil
}

func hasRealConfigValue(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return false
	}
	_, rejected := rejectedAuthConfigValues[normalized]
	return !rejected
}

func hasValidAbsoluteURL(value string) bool {
	parsedURL, err := url.Parse(strings.TrimSpace(value))
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}
