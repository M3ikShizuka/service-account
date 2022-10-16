package oauth2

import (
	client "github.com/ory/hydra-client-go"
	"golang.org/x/net/context"
	"service-account/internal/domain"
)

func (h *OAuth2Service) GetConsentRequest(context context.Context, challenge string) (*domain.OA2ConsentRequest, error) {
	// Get consent request.
	requestGetConsent := h.hydra.AdminApi.GetConsentRequest(context)
	requestGetConsent = requestGetConsent.ConsentChallenge(challenge)
	consentRequest, _, err := requestGetConsent.Execute()
	if err != nil {
		// Error request to hydra OAuth admin API.
		return nil, err
	}

	clientData, _ := consentRequest.Client.MarshalJSON()

	return &domain.OA2ConsentRequest{
			Skip: consentRequest.GetSkip(),
			// Subject is the user ID of the end-user that authenticated. Now, that end user needs to grant or deny the scope requested by the OAuth 2.0 client.
			Subject:                      consentRequest.GetSubject(),
			ClientData:                   clientData,
			RequestedAccessTokenAudience: consentRequest.GetRequestedAccessTokenAudience(),
			RequestedScope:               consentRequest.GetRequestedScope(),
		},
		nil
}

func (h *OAuth2Service) AcceptConsentRequest(context context.Context, challenge string, grantScope []string, grantAccessTokenAudience []string, remember bool, rememberFor int64) (string, error) {
	//// The session allows us to set session data for id and access tokens
	//let session: ConsentRequestSession = {
	//	// This data will be available when introspecting the token. Try to avoid sensitive information here,
	//	// unless you limit who can introspect tokens.
	//access_token: {
	//	// foo: 'bar'
	//},
	//
	//	// This data will be available in the ID token.
	//id_token: {
	//	// baz: 'bar'
	//}
	//}

	// Here is also the place to add data to the ID or access token. For example,
	// if the scope 'profile' is added, add the family and given name to the ID Token claims:
	// if (grantScope.indexOf('profile')) {
	//   session.id_token.family_name = 'Doe'
	//   session.id_token.given_name = 'John'
	// }

	// Accept consent request.
	/*
	 // We can grant all scopes that have been requested - hydra already checked for us that no additional scopes
	  // are requested accidentally.
	  grant_scope: grantScope,

	  // If the environment variable CONFORMITY_FAKE_CLAIMS is set we are assuming that
	  // the app is built for the automated OpenID Connect Conformity Test Suite. You
	  // can peak inside the code for some ideas, but be aware that all data is fake
	  // and this only exists to fake a login system which works in accordance to OpenID Connect.
	  //
	  // If that variable is not set, the session will be used as-is.
	  session: oidcConformityMaybeFakeSession(grantScope, body, session),

	  // ORY hydra checks if requested audiences are allowed by the client, so we can simply echo this.
	  grant_access_token_audience: body.requested_access_token_audience,

	  // This tells hydra to remember this consent request and allow the same client to request the same
	  // scopes from the same user, without showing the UI, in the future.
	  remember: Boolean(req.body.remember),

	  // When this "remember" sesion expires, in seconds. Set this to 0 so it will never expire.
	  remember_for: 3600
	*/

	/*
		// ConsentRequestSession struct for ConsentRequestSession
		type ConsentRequestSession struct {
			// AccessToken sets session data for the access and refresh token, as well as any future tokens issued by the refresh grant. Keep in mind that this data will be available to anyone performing OAuth 2.0 Challenge Introspection. If only your services can perform OAuth 2.0 Challenge Introspection, this is usually fine. But if third parties can access that endpoint as well, sensitive data from the session might be exposed to them. Use with care!
			AccessToken interface{} `json:"access_token,omitempty"`
			// IDToken sets session data for the OpenID Connect ID token. Keep in mind that the session'id payloads are readable by anyone that has access to the ID Challenge. Use with care!
			IdToken interface{} `json:"id_token,omitempty"`
		}
	*/
	consentSession := client.ConsentRequestSession{}
	// TODO: do better value for acceptConsentRequest.SetSession()
	var acceptConsentRequest client.AcceptConsentRequest
	acceptConsentRequest.SetSession(consentSession)
	acceptConsentRequest.SetGrantScope(grantScope)
	acceptConsentRequest.SetGrantAccessTokenAudience(grantAccessTokenAudience)
	acceptConsentRequest.SetRemember(remember)
	acceptConsentRequest.SetRememberFor(rememberFor)

	requestAcceptConsent := h.hydra.AdminApi.AcceptConsentRequest(context)
	requestAcceptConsent = requestAcceptConsent.ConsentChallenge(challenge)
	requestAcceptConsent = requestAcceptConsent.AcceptConsentRequest(acceptConsentRequest)
	completedRequest, _, err := requestAcceptConsent.Execute()
	if err != nil {
		// Error request to hydra OAuth admin API.
		return "", nil
	}

	return completedRequest.RedirectTo, nil
}

func (h *OAuth2Service) RejectConsentRequest(context context.Context, challenge string, errStr string, errDescStr string) (string, error) {
	var rejectRequest client.RejectRequest
	rejectRequest.SetError(errStr)
	rejectRequest.SetErrorDescription(errDescStr)

	request := h.hydra.AdminApi.RejectConsentRequest(context)
	request = request.ConsentChallenge(challenge)
	request = request.RejectRequest(rejectRequest)
	completedRequest, _, err := request.Execute()
	if err != nil {
		// Error request to hydra OAuth admin API.
		return "", err
	}

	return completedRequest.RedirectTo, nil
}
