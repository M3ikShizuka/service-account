package service

import "service-account/internal/service/authz/oauth2"

type Services struct {
	OAuth2 *oauth2.OAuth2Handler
}

func NewService(config *Config) *Services {
	return &Services{
		OAuth2: oauth2.NewOAuth2Handler(config),
	}
}
