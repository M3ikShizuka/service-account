package oauth2

import (
	"fmt"
	client "github.com/ory/hydra-client-go"
	"github.com/ory/x/randx"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"log"
	"service-account/internal/config"
	"service-account/internal/domain"
	"strconv"
)

type OAuth2Service struct {
	hydra      *client.APIClient
	confOAuth2 oauth2.Config

	AuthCodeUrl       string
	AuthState         []rune // I think it's should be generated on client app side.
	logoutUrlTemplate string
}

func NewOAuth2Service(config *config.OAuth2Config) *OAuth2Service {
	// Init OAuth config.
	scopes := []string{"openid", "offline"}

	configuration := client.NewConfiguration()
	configuration.Servers = []client.ServerConfiguration{
		{
			URL: config.HydraAdminURLPrivateLan, // Admin API URL
		},
	}

	confOAuth2 := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: config.Backend,
			AuthURL:  config.Frontend,
		},
		RedirectURL: config.RedirectURLCallback,
		Scopes:      scopes,
	}

	handlerOA2 := &OAuth2Service{
		hydra:      client.NewAPIClient(configuration),
		confOAuth2: confOAuth2,
	}

	// Init auth code url.
	authCodeUrl, authState := handlerOA2.generateAuthCodeURL()
	// Init logout code url.
	logoutUrlTemplate := generateLogoutURLTemplate(config.HydraPublicURL)

	handlerOA2.AuthCodeUrl = authCodeUrl
	handlerOA2.AuthState = authState
	handlerOA2.logoutUrlTemplate = logoutUrlTemplate

	return handlerOA2
}

func (h *OAuth2Service) TokenExchange(ctx context.Context, code string) (*domain.Token, error) {
	token, err := h.confOAuth2.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	// Retrive OpenID token.
	idt := token.Extra("id_token")
	IdToken := idt.(string)

	return &domain.Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		IdToken:      IdToken,
	}, nil
}

func (h *OAuth2Service) generateAuthCodeURL() (string, []rune) {
	state, err := randx.RuneSequence(24, randx.AlphaLower)
	if err != nil {
		log.Fatal(err)
	}

	nonce, err := randx.RuneSequence(24, randx.AlphaLower)
	if err != nil {
		log.Fatal(err)
	}

	authCodeURL := h.confOAuth2.AuthCodeURL(
		string(state),
		oauth2.SetAuthURLParam("nonce", string(nonce)),
		oauth2.SetAuthURLParam("max_age", strconv.Itoa(0)),
	)

	return authCodeURL, state
}

func generateLogoutURLTemplate(hydraPublicURL string) string {
	return fmt.Sprintf("%s%s?id_token_hint=%%s&state=%%s&post_logout_redirect_uri=%%s", hydraPublicURL, "/oauth2/sessions/logout")
}

func (h *OAuth2Service) GenerateLogoutURL(idTokenHint string, state string, postLogoutRedirectUri string) string {
	return fmt.Sprintf(h.logoutUrlTemplate, idTokenHint, state, postLogoutRedirectUri)
}

func (h *OAuth2Service) GetAuthCodeUrl() string {
	return h.AuthCodeUrl
}
