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
	"service-account/pkg/logger"
	"strconv"
)

type OAuth2Handler struct {
	Hydra      *client.APIClient
	ConfOAuth2 oauth2.Config

	AuthCodeUrl       string
	AuthState         []rune // i think it's should be generated on client app side.
	LogoutUrlTemplate string
}

func NewOAuth2Handler(config *config.ConfigOAuth2) *OAuth2Handler {
	// Init OAuth config.
	scopes := []string{"openid", "offline"}

	configuration := client.NewConfiguration()
	configuration.Servers = []client.ServerConfiguration{
		{
			URL: config.HydraAdminURL, // Admin API URL
		},
	}

	confOAuth2 := oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: config.Backend,
			AuthURL:  config.Frontend,
		},
		RedirectURL: config.RedirectUrl,
		Scopes:      scopes,
	}

	// Init auth code url.
	authCodeUrl, authState := generateAuthCodeURL()
	// Init logout code url.
	logoutUrlTemplate := generateLogoutURLTemplate()

	return &OAuth2Handler{
		Hydra:             client.NewAPIClient(configuration),
		ConfOAuth2:        confOAuth2,
		AuthCodeUrl:       authCodeUrl,
		AuthState:         authState,
		LogoutUrlTemplate: logoutUrlTemplate,
	}
}

func (h *OAuth2Handler) GetLoginRequest(context context.Context, challenge string) (*domain.OA2LoginRequest, err error) {
	requestGetLogin := h.Hydra.AdminApi.GetLoginRequest(context)
	requestGetLogin = requestGetLogin.LoginChallenge(challenge)
	loginRequestResponseData, _, errGetLogin := requestGetLogin.Execute()
	if errGetLogin != nil {
		return nil, errGetLogin
	}

	return &domain.OA2LoginRequest{}
}

func (h *OAuth2Handler) AcceptLoginRequest(context context.Context, challenge string, subject string, remember bool, rememberFor int64) (string, error) {
	var acceptLoginRequest client.AcceptLoginRequest
	acceptLoginRequest.SetSubject(subject)
	acceptLoginRequest.SetRemember(remember)
	acceptLoginRequest.SetRememberFor(3600)

	// Sets which "level" (e.g. 2-factor authentication) of authentication the user has. The value is really arbitrary
	// and optional. In the context of OpenID Connect, a value of 0 indicates the lowest authorization level.
	// acr: '0',
	//
	// If the environment variable CONFORMITY_FAKE_CLAIMS is set we are assuming that
	// the app is built for the automated OpenID Connect Conformity Test Suite. You
	// can peak inside the code for some ideas, but be aware that all data is fake
	// and this only exists to fake a login system which works in accordance to OpenID Connect.
	//
	// If that variable is not set, the ACR value will be set to the default passed here ('0')

	//acr:
	//	oidcConformityMaybeFakeAcr(loginRequest, '0')
	//	acceptLoginRequest.SetAcr()

	// TODO:  acceptLoginRequest.SetAcr()
	// acr - sets the Authentication AuthorizationContext Class Reference value for this authentication session. You can use it to express that, for example, a user authenticated using two factor authentication.
	// SRC: https://www.ory.sh/docs/hydra/concepts/login

	requestAcceptLogin := Hydra.AdminApi.AcceptLoginRequest(context)
	requestAcceptLogin = requestAcceptLogin.LoginChallenge(challenge)
	requestAcceptLogin = requestAcceptLogin.AcceptLoginRequest(acceptLoginRequest)
	completedRequestAcceptLogin, responseAcceptLogin, errAcceptLogin := requestAcceptLogin.Execute()
	/*
		completedRequestAcceptLogin = {*client.CompletedRequest | 0xc00008c250}
		 RedirectTo = {string} "http://127.0.0.1:4444/oauth2/auth?client_id=auth-code-client&login_verifier=dc6e47d889574c939ec3ace9"

		responseAcceptLogin = {*http.Response | 0xc00041c1b0}
		 Status = {string} "200 OK"
		 StatusCode = {int} 200
		 Proto = {string} "HTTP/1.1"

		Request = {*http.Request | 0xc0003e8200} PUT http://127.0.0.1:4445/oauth2/auth/requests/login/accept
		 Method = {string} "PUT"
		 URL = {*url.URL | 0xc00041c120}
		 Proto = {string} "HTTP/1.1"
	*/
	if errAcceptLogin != nil {
		// Error request to hydra OAuth admin API.
		if responseAcceptLogin != nil {
			logger.Error("handlerLoginPost() - AcceptLoginRequest() result:\n• err: %v\n• response: %v\n", errAcceptLogin, responseAcceptLogin)
		} else {
			logger.Error("handlerLoginPost() - AcceptLoginRequest() result: %v\n", errAcceptLogin)
		}

		//context.AbortWithError(http.StatusInternalServerError, errAcceptLogin)
		return "", errAcceptLogin
	}

	return completedRequestAcceptLogin.RedirectTo, nil
	//context.Redirect(http.StatusFound, completedRequestAcceptLogin.RedirectTo)
}

func generateAuthCodeURL() (string, []rune) {
	state, err := randx.RuneSequence(24, randx.AlphaLower)
	if err != nil {
		log.Fatal(err)
	}

	nonce, err := randx.RuneSequence(24, randx.AlphaLower)
	if err != nil {
		log.Fatal(err)
	}

	authCodeURL := ConfOAuth2.AuthCodeURL(
		string(state),
		oauth2.SetAuthURLParam("nonce", string(nonce)),
		oauth2.SetAuthURLParam("max_age", strconv.Itoa(0)),
	)

	return authCodeURL, state
}

func generateLogoutURLTemplate() string {
	// TODO: get params from config file.
	hydraProto := "http"
	hydraHost := "127.0.0.1"
	hydraPublicPort := 4444

	hydraHostURL := fmt.Sprintf("%s://%s", hydraProto, hydraHost)
	hydraPublicURL := fmt.Sprintf("%s:%d", hydraHostURL, hydraPublicPort)

	return fmt.Sprintf("%s%s?id_token_hint=%%s&state=%%s&post_logout_redirect_uri=%%s", hydraPublicURL, "/oauth2/sessions/logout")
}

func GenerateLogoutURL(logoutUrlTemplate string, idTokenHint string, state string, postLogoutRedirectUri string) string {
	return fmt.Sprintf(logoutUrlTemplate, idTokenHint, state, postLogoutRedirectUri)
}
