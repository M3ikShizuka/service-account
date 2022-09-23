package oauth2

import (
	client "github.com/ory/hydra-client-go"
	"golang.org/x/net/context"
	"service-account/internal/domain"
)

func (h *OAuth2Service) GetLoginRequest(context context.Context, challenge string) (*domain.OA2LoginRequest, error) {
	/*
		GET http://127.0.0.1:4445/oauth2/auth/requests/login
		Status = {string} "200 OK"
		StatusCode = {int} 200
		Proto = {string} "HTTP/1.1"
		{"challenge":"5abbeb3853264c36993da5b2a1468ad7","requested_scope":["openid","offline"],"requested_access_token_audience":[],"skip":false,"subject":"","oidc_context":{},"client":{"client_id":"client-auth-code-service-account","client_name":"","redirect_uris":["http://127.0.0.1:5555/callback"],"grant_types":["authorization_code","refresh_token"],"response_types":["code","id_token"],"scope":"openid offline","audience":[],"owner":"","policy_uri":"","allowed_cors_origins":[],"tos_uri":"","client_uri":"","logo_uri":"","contacts":[],"client_secret_expires_at":0,"subject_type":"public","jwks":{},"token_endpoint_auth_method":"client_secret_basic","userinfo_signed_response_alg":"none","created_at":"2022-07-28T15:43:17Z","updated_at":"2022-07-28T15:43:17.303143Z","metadata":{}},"request_url":"http://127.0.0.1:4444/oauth2/auth?client_id=client-auth-code-service-account\u0026max_age=0\u0026nonce=eqemnkccxrwxyripqscbhagw\u0026redirect_uri=http%3A%2F%2F127.0.0.1%3A3000%2Fcallback\u0026response_type=code\u0026scope=openid+offline\u0026state=evgfgsnclrwvoumhuqhbazkq","session_id":"76830b22-bf47-4162-9635-0b211ffb030e"}
	*/
	requestGetLogin := h.hydra.AdminApi.GetLoginRequest(context)
	requestGetLogin = requestGetLogin.LoginChallenge(challenge)
	loginRequestResponseData, _, errGetLogin := requestGetLogin.Execute()
	if errGetLogin != nil {
		// Error request to hydra OAuth admin API.
		return nil, errGetLogin
	}

	// Get hint.
	oidcContext := loginRequestResponseData.GetOidcContext()
	var hint string
	if hintPtr, ok := oidcContext.GetLoginHintOk(); ok {
		hint = *hintPtr
	}

	return &domain.OA2LoginRequest{
			Skip:    loginRequestResponseData.GetSkip(),
			Subject: loginRequestResponseData.GetSubject(),
			Hint:    hint,
		},
		nil
}

func (h *OAuth2Service) AcceptLoginRequest(context context.Context, challenge string, subject string, remember bool, rememberFor int64) (string, error) {
	var acceptLoginRequest client.AcceptLoginRequest
	acceptLoginRequest.SetSubject(subject)
	acceptLoginRequest.SetRemember(remember)
	acceptLoginRequest.SetRememberFor(rememberFor)

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

	requestAcceptLogin := h.hydra.AdminApi.AcceptLoginRequest(context)
	requestAcceptLogin = requestAcceptLogin.LoginChallenge(challenge)
	requestAcceptLogin = requestAcceptLogin.AcceptLoginRequest(acceptLoginRequest)
	completedRequest, _, err := requestAcceptLogin.Execute()
	if err != nil {
		// Error request to hydra OAuth admin API.
		return "", err
	}

	/*
		completedRequestAcceptLogin = {*client.CompletedRequest | 0xc00008c250}
		 RedirectTo = {string} "http://127.0.0.1:4444/oauth2/auth?client_id=client-auth-code-service-account&login_verifier=dc6e47d889574c939ec3ace9"

		responseAcceptLogin = {*http.Response | 0xc00041c1b0}
		 Status = {string} "200 OK"
		 StatusCode = {int} 200
		 Proto = {string} "HTTP/1.1"

		Request = {*http.Request | 0xc0003e8200} PUT http://127.0.0.1:4445/oauth2/auth/requests/login/accept
		 Method = {string} "PUT"
		 URL = {*url.URL | 0xc00041c120}
		 Proto = {string} "HTTP/1.1"
	*/

	return completedRequest.RedirectTo, nil
}

func (h *OAuth2Service) RejectLoginRequest(context context.Context, challenge string, errStr string, errDescStr string) (string, error) {
	/*
		type AdminApiApiRejectLoginRequestRequest struct {
			ctx context.Context
			ApiService AdminApi
			loginChallenge *string
			rejectRequest *RejectRequest
		}

		// RejectRequest struct for RejectRequest
		type RejectRequest struct {
			// The error should follow the OAuth2 error format (e.g. `invalid_request`, `login_required`).  Defaults to `request_denied`.
			Error *string `json:"error,omitempty"`
			// Debug contains information to help resolve the problem as a developer. Usually not exposed to the public but only in the server logs.
			ErrorDebug *string `json:"error_debug,omitempty"`
			// Description of the error in a human readable format.
			ErrorDescription *string `json:"error_description,omitempty"`
			// Hint to help resolve the error.
			ErrorHint *string `json:"error_hint,omitempty"`
			// Represents the HTTP status code of the error (e.g. 401 or 403)  Defaults to 400
			StatusCode *int64 `json:"status_code,omitempty"`
		}
	*/
	var rejectRequest client.RejectRequest
	rejectRequest.SetError(errStr)
	rejectRequest.SetErrorDescription(errDescStr)

	request := h.hydra.AdminApi.RejectLoginRequest(context)
	request = request.LoginChallenge(challenge)
	request = request.RejectRequest(rejectRequest)

	completedRequest, _, err := request.Execute()
	if err != nil {
		// Error request to hydra OAuth admin API.
		return "", err
	}

	return completedRequest.RedirectTo, nil
}
