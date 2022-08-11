package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/ory/viper"
)

const (
	defHttpAddr = "0.0.0.0"
	defHttpPort = 3000
)

type Config struct {
	HTTP   HTTPConfig   `mapstructure:"http"`
	OAuth2 OAuth2Config `mapstructure:"oauth2"`
}

type HTTPConfig struct {
	Proto   string `mapstructure:"proto" validate:"required"`
	Addr    string `mapstructure:"addr" validate:"required"`
	Port    int32  `mapstructure:"port" validate:"required"`
	Host    string
	HostURL string
}

type OAuth2Config struct {
	ClientID        string `mapstructure:"client_id" validate:"required"`
	ClientSecret    string `mapstructure:"client_secret" validate:"required"`
	HydraProto      string `mapstructure:"hydra_proto" validate:"required"`
	HydraHost       string `mapstructure:"hydra_host" validate:"required"`
	HydraPublicPort int32  `mapstructure:"hydra_public_port" validate:"required"`
	HydraAdminPort  int32  `mapstructure:"hydra_admin_port" validate:"required"`
	// Init based on data.
	HydraHostURL   string
	HydraPublicURL string
	HydraAdminURL  string
	ConsentURL     string
	RedirectUrl    string
	Backend        string
	Frontend       string
}

func NewConfig() *Config {
	return &Config{}
}

func (config *Config) Init(configPath string) error {
	// TODO: init config from file.
	// Set default values.
	config.setDefault()

	// Load settings from config file.
	if err := config.parseConfig(configPath); err != nil {
		return err
	}

	// Unmarshal config to struct.
	if err := config.unmarshal(); err != nil {
		return err
	}

	config.initСompositeFields()

	return nil
}

func (config *Config) setDefault() {
	viper.SetDefault("http.addr", defHttpAddr)
	viper.SetDefault("http.port", defHttpPort)
}

func (config *Config) parseConfig(configPath string) error {
	viper.SetConfigType("yml")
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

func (config *Config) unmarshal() error {
	//// Unmarshal all but not check.
	if err := viper.Unmarshal(config); err != nil {
		return err
	}

	// Check initialization of all important fields.
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return fmt.Errorf("Initializing the configuration: Missing required attributes %w\n", err)
	}

	return nil
}

func (config *Config) initСompositeFields() {
	// TODO: init Server struct from config.
	config.HTTP.Host = fmt.Sprintf("%s:%d", config.HTTP.Addr, config.HTTP.Port)
	config.HTTP.HostURL = fmt.Sprintf("%s://%s", config.HTTP.Proto, config.HTTP.Host)

	// The consent procedure process by the same service.
	config.OAuth2.ConsentURL = config.HTTP.HostURL

	config.OAuth2.HydraHostURL = fmt.Sprintf("%s://%s", config.OAuth2.HydraProto, config.OAuth2.HydraHost)
	config.OAuth2.HydraPublicURL = fmt.Sprintf("%s:%d", config.OAuth2.HydraHostURL, config.OAuth2.HydraPublicPort)
	config.OAuth2.HydraAdminURL = fmt.Sprintf("%s:%d", config.OAuth2.HydraHostURL, config.OAuth2.HydraAdminPort)

	config.OAuth2.RedirectUrl = config.HTTP.HostURL + "/callback"
	config.OAuth2.Backend = fmt.Sprintf("%s%s", config.OAuth2.HydraPublicURL, "/oauth2/token")
	config.OAuth2.Frontend = fmt.Sprintf("%s%s", config.OAuth2.HydraPublicURL, "/oauth2/auth")
}
