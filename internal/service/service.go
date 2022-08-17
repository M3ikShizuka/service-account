package service

import (
	"golang.org/x/net/context"
	"service-account/internal/config"
	"service-account/internal/domain"
	"service-account/internal/service/authz/oauth2"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type OAuth2 interface {
	TokenExchange(ctx context.Context, code string) (*domain.Token, error)
	GetLoginRequest(context context.Context, challenge string) (*domain.OA2LoginRequest, error)
	AcceptLoginRequest(context context.Context, challenge string, subject string, remember bool, rememberFor int64) (string, error)
	RejectLoginRequest(context context.Context, challenge string, errStr string, errDescStr string) (string, error)
	GetConsentRequest(context context.Context, challenge string) (*domain.OA2ConsentRequest, error)
	AcceptConsentRequest(context context.Context, challenge string, grantScope []string, grantAccessTokenAudience []string, remember bool, rememberFor int64) (string, error)
	RejectConsentRequest(context context.Context, challenge string, errStr string, errDescStr string) (string, error)
	RejectLogoutRequest(context context.Context, challenge string) error
	AcceptLogoutRequest(context context.Context, challenge string) (string, error)
	IntrospectOAuth2Token(context context.Context, accessToken string) (*domain.OA2TokenIntrospection, error)
	GenerateLogoutURL(idTokenHint string, state string, postLogoutRedirectUri string) string
	GetAuthCodeUrl() string
}

type Services struct {
	Config *config.Config
	OAuth2 OAuth2 // AuthZ
	// TODO: AuthN  *authn.AuthNHandler   // AuthN
}

func NewService(cfg *config.Config) *Services {
	oa2 := oauth2.NewOAuth2Service(&cfg.OAuth2)
	return &Services{
		Config: cfg,
		OAuth2: oa2,
		// TODO: AuthN
	}
}
