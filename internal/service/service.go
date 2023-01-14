package service

import (
	"golang.org/x/net/context"
	"service-account/internal/config"
	"service-account/internal/domain"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type Hasher interface {
	Hash(password string, salt []byte) []byte
}

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserById(ctx context.Context, id uint32) (*domain.User, error)
}

// Dependencies of services.
type Depends struct {
	UserRepo UserRepository
	Hasher   Hasher
}

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

type User interface {
	SignUp(ctx context.Context, inputUserData *UserSignUpInput) error
	SignIn(ctx context.Context, inputUserData *UserSignInInput) (*domain.User, error)
	GetUserById(ctx context.Context, id uint32) (*domain.User, error)
}

type Services struct {
	Config *config.Config
	OAuth2 OAuth2 // AuthZ
	User   User
	// TODO: AuthN  *authn.AuthNHandler   // AuthN
}

func NewService(
	config *config.Config,
	depends *Depends,
	oa2 OAuth2,
	userService User,
) *Services {
	return &Services{
		Config: config,
		OAuth2: oa2,
		User:   userService,
		// TODO: AuthN
	}
}
