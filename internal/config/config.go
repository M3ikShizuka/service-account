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
	DB     Database     `mapstructure:"database"`
}

type HTTPConfig struct {
	Proto      string `mapstructure:"proto" validate:"required"`
	ListenAddr string `mapstructure:"listen_addr" validate:"required"`
	Port       int32  `mapstructure:"port" validate:"required"`
	Host       string
	HostURL    string
}

type OAuth2Config struct {
	ClientID     string `mapstructure:"client_id" validate:"required"`
	ClientSecret string `mapstructure:"client_secret" validate:"required"`
	HydraProto   string `mapstructure:"hydra_proto" validate:"required"`
	RedirectAddr string `mapstructure:"redirect_addr" validate:"required"`
	// Init based on data.
	HydraPublicURL           string `mapstructure:"hydra_public_host" validate:"required"`
	HydraPublicURLPrivateLan string `mapstructure:"hydra_public_host_private_lan" validate:"required"`
	HydraAdminURLPrivateLan  string `mapstructure:"hydra_admin_host_private_lan" validate:"required"`
	ConsentURL               string
	RedirectHost             string // It's for redirect to app from OAuth service.
	RedirectURL              string
	RedirectURLCallback      string
	Backend                  string
	Frontend                 string
}

type Database struct {
	DSN  string `mapstructure:"dsn" validate:"required"`
	Salt string `mapstructure:"salt" validate:"required"`
}

func NewConfig() *Config {
	return &Config{}
}

func (config *Config) Init(configPath string) error {
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

	// Env
	config.getEnv()

	// Init composite fields.
	config.initСompositeFields()

	return nil
}

func (config *Config) setDefault() {
	viper.SetDefault("http.listen_addr", defHttpAddr)
	viper.SetDefault("http.port", defHttpPort)
}

func (config *Config) parseConfig(configPath string) error {
	viper.SetConfigType("yml")
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()
	//viper.SetEnvPrefix("SERVICE_ACCOUNT") // will be uppercased automatically

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

func (config *Config) getEnv() {
	if envar := viper.GetString("SERVICE_ACCOUNT_CLIENT_ID"); envar != "" {
		config.OAuth2.ClientID = envar
	}

	if envar := viper.GetString("SERVICE_ACCOUNT_CLIENT_SECRET"); envar != "" {
		config.OAuth2.ClientSecret = envar
	}

	if envar := viper.GetString("SERVICE_ACCOUNT_HYDRA_PROTO"); envar != "" {
		config.OAuth2.HydraProto = envar
	}

	if envar := viper.GetString("SERVICE_ACCOUNT_HYDRA_PUBLIC_HOST"); envar != "" {
		config.OAuth2.HydraPublicURL = envar
	}

	if envar := viper.GetString("SERVICE_ACCOUNT_HYDRA_PUBLIC_HOST_PRIVATE_LAN"); envar != "" {
		config.OAuth2.HydraPublicURLPrivateLan = envar
	}

	if envar := viper.GetString("SERVICE_ACCOUNT_HYDRA_ADMIN_HOST_PRIVATE_LAN"); envar != "" {
		config.OAuth2.HydraAdminURLPrivateLan = envar
	}

	if envar := viper.GetString("SERVICE_ACCOUNT_REDIRECT_ADDR"); envar != "" {
		config.OAuth2.RedirectAddr = envar
	}

	if envar := viper.GetString("SERVICE_ACCOUNT_DSN"); envar != "" {
		config.DB.DSN = envar
	}

	if envar := viper.GetString("SERVICE_ACCOUNT_SALT"); envar != "" {
		config.DB.Salt = envar
	}
}

func (config *Config) initСompositeFields() {
	config.HTTP.Host = fmt.Sprintf("%s:%d", config.HTTP.ListenAddr, config.HTTP.Port)
	config.HTTP.HostURL = fmt.Sprintf("%s://%s", config.HTTP.Proto, config.HTTP.Host)
	// The consent procedure process by the same service.
	config.OAuth2.ConsentURL = config.HTTP.HostURL
	config.OAuth2.RedirectHost = config.OAuth2.RedirectAddr
	config.OAuth2.RedirectURL = fmt.Sprintf("%s://%s", config.HTTP.Proto, config.OAuth2.RedirectHost)
	config.OAuth2.RedirectURLCallback = fmt.Sprintf("%s/callback", config.OAuth2.RedirectURL)
	config.OAuth2.Backend = fmt.Sprintf("%s%s", config.OAuth2.HydraPublicURLPrivateLan, "/oauth2/token")
	config.OAuth2.Frontend = fmt.Sprintf("%s%s", config.OAuth2.HydraPublicURL, "/oauth2/auth")
}
