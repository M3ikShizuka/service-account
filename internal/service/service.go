package service

import (
	"service-account/internal/config"
	"service-account/internal/service/authz/oauth2"
)

type Services struct {
	Config *config.Config
	OAuth2 *oauth2.HandlerOAuth2 // AuthZ
	// TODO: AuthN  *authn.AuthNHandler   // AuthN
}

func NewService(cfg *config.Config) *Services {
	return &Services{
		Config: cfg,
		OAuth2: oauth2.NewOAuth2Handler(&cfg.OAuth2),
		// TODO: AuthN
	}
}
