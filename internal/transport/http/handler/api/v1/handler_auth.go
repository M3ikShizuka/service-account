// Example: https://github.com/ory/hydra-consent-app-go/blob/master/main.go
package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	client "github.com/ory/hydra-client-go"
	"net/http"
	"os"
	oauth22 "service-account/internal/service/authz/oauth2"
	"service-account/internal/transport/http/coockie"
	"service-account/pkg/logger"
	"time"
)

const (
	submitDenyAccess  = "Deny access"
	submitLogIn       = "Log in"
	submitAllowAccess = "Allow access"
	submitNo          = "No"
	submitYes         = "Yes"
	// Paths
	pathRoot                      = "/"
	pathLogin              string = "/login"
	pathConsent            string = "/consent"
	pathCallback           string = "/callback"
	pathLogout             string = "/logout"
	pathLogoutBackchannel  string = "/backchannel-logout"
	pathLogoutFrontchannel string = "/frontchannel-logout"
)

func (handler *HandlerAPIv1) initHandlersAuthentication(router *gin.RouterGroup) {
	// Init router.
	// Log in
	router.GET(pathLogin, handlerLoginGet)
	router.POST(pathLogin, handlerLoginPost)
	// Consent
	router.GET(pathConsent, handlerConsentGet)
	router.POST(pathConsent, handlerConsentPost)
	// Callback
	router.GET(pathCallback, handlerCallback)
	// Logout
	router.GET(pathLogout, handlerLogoutGet)
	router.POST(pathLogout, handlerLogoutPost)
	router.Any(pathLogoutBackchannel, handlerLogoutBackchannel)
	router.GET(pathLogoutFrontchannel, handlerLogoutFrontchannel)
}

func (h *HandlerAPIv1) handlerLoginGet(context *gin.Context) {
	// Login sessions, prompt, max_age, id_token_hint
	// https://<hydra-public>:4444/oauth2/auth?prompt=login&max_age=60&id_token_hint=...'
	// SRC: https://www.ory.sh/docs/hydra/concepts/login#login-sessions-prompt-max_age-id_token_hint
	// return Login page.html or skip and auth user by login_challenge.
	// We can get login_challenge by sent id_token_hint to https://<hydra-public>:4444/oauth2/auth?id_token_hint=... for re-auth automaticly.
	// http://127.0.0.1:3000/login?login_challenge=9d54379b39094ba283ebd5d361b9afe6
	// code 200
	challenge := context.Query("login_challenge")
	if challenge == "" {
		context.String(http.StatusBadRequest, "Expected a login challenge to be set but received none.")
		return
	}

	// Get login request.
	loginRequestData, err := h.services.OAuth2.GetLoginRequest(context, challenge)
	if err != nil {
		//// Error request to hydra OAuth admin API.
		//if responseGetLogin != nil {
		//	logger.Error("GetLoginRequest() result:\n• err: %v\n• response: %v\n", errGetLogin, responseGetLogin)
		//} else {
		logger.Error("GetLoginRequest()",
			logger.NamedError("error", errGetLogin),
		)
		//}

		context.AbortWithError(http.StatusInternalServerError, errGetLogin)
	}

	// If hydra was already able to authenticate the user, skip will be true and we do not need to re-authenticate
	// the user.
	if loginRequestData.GetSkip() {
		// You can apply logic here, for example update the number of times the user logged in.
		// ...

		// Now it's time to grant the login request. You could also deny the request if something went terribly wrong
		// (e.g. your arch-enemy logging in...)

		// Accept login.
		//RedirectTo, err := oauth2.AcceptLoginRequest(context, challenge, loginRequestResponseData.GetSubject(), true, 3600)
		RedirectTo, err := h.AcceptLoginRequest(context, challenge, loginRequestResponseData.GetSubject(), true, 3600)
		if err != nil {
			//context.AbortWithError(http.StatusInternalServerError, err)
			return nil, err
		}

		//context.Redirect(http.StatusFound, RedirectTo)
		return RedirectTo, nil
	}

	// Get hint.
	oidcContext := loginRequestResponseData.GetOidcContext()
	var hint string
	if hintPtr, ok := oidcContext.GetLoginHintOk(); ok {
		hint = *hintPtr
	}

	return hint, http.StatusOK, nil

	// Render login html.
	// TODO: csrfToken for forms.
	context.HTML(http.StatusOK, "login.html",
		gin.H{
			"csrfToken": "",
			"challenge": challenge,
			"action":    "/login",
			"hint":      hint,
		})
}

func handlerLoginPost(context *gin.Context) {
	// Check authN data
	// redirect to hydra public :4444/ouath2/auth
	// Code 302

	challenge := context.PostForm("challenge")
	if challenge == "" {
		context.String(http.StatusBadRequest, "Expected a login challenge to be set but received none.")
		return
	}

	submit := context.PostForm("submit")
	if submit == submitDenyAccess {
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
		rejectRequest.SetError("access_denied")
		rejectRequest.SetErrorDescription("The resource owner denied the request")

		request := oauth22.Hydra.AdminApi.RejectLoginRequest(context)
		request = request.LoginChallenge(challenge)
		request = request.RejectRequest(rejectRequest)
		completedRejectLoginRequest, responseRejectLogin, errRejectLogin := request.Execute()
		if errRejectLogin != nil {
			// Error request to hydra OAuth admin API.
			if responseRejectLogin != nil {
				logger.Error("handlerLoginPost() - RejectLoginRequest() result:\n• err: %v\n• response: %v\n", errRejectLogin, responseRejectLogin)
			} else {
				logger.Error("handlerLoginPost() - RejectLoginRequest() result: %v\n", errRejectLogin)
			}

			context.AbortWithError(http.StatusInternalServerError, errRejectLogin)
			return
		}

		switch responseRejectLogin.StatusCode {
		case http.StatusOK:
			context.Redirect(http.StatusFound, completedRejectLoginRequest.RedirectTo)
		case http.StatusNotFound:
			// Accessing to response details
			// cast err to *client.GenericOpenAPIError object first and then
			// to your desired type
			notFound, ok := errRejectLogin.(*client.GenericOpenAPIError).Model().(client.JsonError)
			fmt.Println(ok)
			fmt.Println(*notFound.ErrorDescription)
		case http.StatusGone:
			responseDetail, ok := errRejectLogin.(*client.GenericOpenAPIError).Model().(client.RequestWasHandledResponse)
			fmt.Println(responseDetail, ok)
			fmt.Println("It's gone")
		default:
			fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.RejectLoginRequest``: %v\n", errRejectLogin)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", responseRejectLogin)
		}
	} else if submit != submitLogIn {
		context.String(http.StatusBadRequest, "Unexpected submit!")
		return
	}

	// Check the user's credentials
	var userEmail = context.PostForm("email")
	var userPassword = context.PostForm("password")

	if userEmail != "foo@bar.com" || userPassword != "foobar" {
		// Render login html with error.
		// TODO: csrfToken for forms.
		context.HTML(http.StatusOK, "login.html",
			gin.H{
				"csrfToken": "",
				"challenge": challenge,
				"action":    "/login",
				"error":     "Provided credentials are wrong, try foo@bar.com:foobar",
			},
		)
		return
	}

	// Get login request.
	/*
		GET http://127.0.0.1:4445/oauth2/auth/requests/login
		Status = {string} "200 OK"
		StatusCode = {int} 200
		Proto = {string} "HTTP/1.1"
		{"challenge":"5abbeb3853264c36993da5b2a1468ad7","requested_scope":["openid","offline"],"requested_access_token_audience":[],"skip":false,"subject":"","oidc_context":{},"client":{"client_id":"auth-code-client","client_name":"","redirect_uris":["http://127.0.0.1:5555/callback"],"grant_types":["authorization_code","refresh_token"],"response_types":["code","id_token"],"scope":"openid offline","audience":[],"owner":"","policy_uri":"","allowed_cors_origins":[],"tos_uri":"","client_uri":"","logo_uri":"","contacts":[],"client_secret_expires_at":0,"subject_type":"public","jwks":{},"token_endpoint_auth_method":"client_secret_basic","userinfo_signed_response_alg":"none","created_at":"2022-07-28T15:43:17Z","updated_at":"2022-07-28T15:43:17.303143Z","metadata":{}},"request_url":"http://127.0.0.1:4444/oauth2/auth?client_id=auth-code-client\u0026max_age=0\u0026nonce=eqemnkccxrwxyripqscbhagw\u0026redirect_uri=http%3A%2F%2F127.0.0.1%3A3000%2Fcallback\u0026response_type=code\u0026scope=openid+offline\u0026state=evgfgsnclrwvoumhuqhbazkq","session_id":"76830b22-bf47-4162-9635-0b211ffb030e"}
	*/
	requestGetLogin := oauth22.Hydra.AdminApi.GetLoginRequest(context)
	requestGetLogin = requestGetLogin.LoginChallenge(challenge)
	_, responseGetLogin, errGetLogin := requestGetLogin.Execute()
	if errGetLogin != nil {
		// Error request to hydra OAuth admin API.
		if responseGetLogin != nil {
			logger.Error("handlerLoginPost() - GetLoginRequest() result:\n• err: %v\n• response: %v\n", errGetLogin, responseGetLogin)
		} else {
			logger.Error("handlerLoginPost() - GetLoginRequest() result:%v\n", errGetLogin)
		}

		context.AbortWithError(http.StatusInternalServerError, errGetLogin)
		return
	}

	// Get accept login request.
	var remember bool = false
	if context.PostForm("remember") != "" {
		remember = true
	}

	// Accept login.
	RedirectTo, err := oauth22.AcceptLoginRequest(context, challenge, userEmail, remember, 3600)
	if err != nil {
		return
	}

	context.Redirect(http.StatusFound, RedirectTo)
}

func handlerConsentGet(context *gin.Context) {
	challenge := context.Query("consent_challenge")
	if challenge == "" {
		context.String(http.StatusBadRequest, "Expected a login challenge to be set but received none.")
		return
	}

	// Get consent request.
	requestGetConsent := oauth22.Hydra.AdminApi.GetConsentRequest(context)
	requestGetConsent = requestGetConsent.ConsentChallenge(challenge)
	consentRequestResponseData, responseGetConsent, errGetConsent := requestGetConsent.Execute()
	if errGetConsent != nil {
		// Error request to hydra OAuth admin API.
		if responseGetConsent != nil {
			logger.Error("handlerConsentGet() - GetConsentRequest() result:\n• err: %v\n• response: %v\n", errGetConsent, responseGetConsent)
		} else {
			logger.Error("handlerConsentGet() - GetConsentRequest() result:%v\n", errGetConsent)
		}

		context.AbortWithError(http.StatusInternalServerError, errGetConsent)
		return
	}

	if consentRequestResponseData.GetSkip() {
		// You can apply logic here, for example grant another scope, or do whatever...
		// ...

		// Now it's time to grant the consent request. You could also deny the request if something went terribly wrong
		/*
			{
			// We can grant all scopes that have been requested - hydra already checked for us that no additional scopes
			// are requested accidentally.
			grant_scope: body.requested_scope,

			// ORY Hydra checks if requested audiences are allowed by the client, so we can simply echo this.
			grant_access_token_audience: body.requested_access_token_audience,

			// The session allows us to set session data for id and access tokens
			session: {
			  // This data will be available when introspecting the token. Try to avoid sensitive information here,
			  // unless you limit who can introspect tokens.
			  // accessToken: { foo: 'bar' },
			  // This data will be available in the ID token.
			  // idToken: { baz: 'bar' },
			}
		*/

		var acceptConsentRequest client.AcceptConsentRequest
		acceptConsentRequest.SetGrantScope(consentRequestResponseData.GetRequestedScope())
		acceptConsentRequest.SetGrantAccessTokenAudience(consentRequestResponseData.GetRequestedAccessTokenAudience())
		/*
			// ConsentRequestSession struct for ConsentRequestSession
			type ConsentRequestSession struct {
				// AccessToken sets session data for the access and refresh token, as well as any future tokens issued by the refresh grant. Keep in mind that this data will be available to anyone performing OAuth 2.0 Challenge Introspection. If only your services can perform OAuth 2.0 Challenge Introspection, this is usually fine. But if third parties can access that endpoint as well, sensitive data from the session might be exposed to them. Use with care!
				AccessToken interface{} `json:"access_token,omitempty"`
				// IDToken sets session data for the OpenID Connect ID token. Keep in mind that the session'id payloads are readable by anyone that has access to the ID Challenge. Use with care!
				IdToken interface{} `json:"id_token,omitempty"`
			}
		*/
		// acceptConsentRequest.SetSession()

		requestAcceptConsent := oauth22.Hydra.AdminApi.AcceptConsentRequest(context)
		requestAcceptConsent = requestAcceptConsent.ConsentChallenge(challenge)
		requestAcceptConsent = requestAcceptConsent.AcceptConsentRequest(acceptConsentRequest)
		completedRequestAcceptConsent, responseAcceptConsent, errAcceptConsent := requestAcceptConsent.Execute()
		if errAcceptConsent != nil {
			// Error request to hydra OAuth admin API.
			if responseAcceptConsent != nil {
				logger.Error("handlerConsentPost() - AcceptConsentRequest() result:\n• err: %v\n• response: %v\n", errAcceptConsent, responseAcceptConsent)
			} else {
				logger.Error("handlerConsentPost() - AcceptConsentRequest() result: %v\n", errAcceptConsent)
			}

			context.AbortWithError(http.StatusInternalServerError, errAcceptConsent)
			return
		}

		context.Redirect(http.StatusFound, completedRequestAcceptConsent.RedirectTo)
	}

	// Render consent html.
	// TODO: csrfToken for forms.
	context.HTML(http.StatusOK, "consent.html",
		gin.H{
			"csrfToken": "",
			"challenge": challenge,
			// We have a bunch of data available from the response, check out the API docs to find what these values mean
			// and what additional data you have available.
			"requested_scope": consentRequestResponseData.GetRequestedScope(),
			"user":            consentRequestResponseData.GetSubject(),
			"client":          consentRequestResponseData.GetClient(),
			"action":          "/consent",
		})
}

func handlerConsentPost(context *gin.Context) {
	challenge := context.PostForm("challenge")
	if challenge == "" {
		context.String(http.StatusBadRequest, "Expected a login challenge to be set but received none.")
		return
	}

	submit := context.PostForm("submit")
	if submit == submitDenyAccess {
		/*
			type AdminApiApiRejectConsentRequestRequest struct {
				ctx context.Context
				ApiService AdminApi
				consentChallenge *string
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
		rejectRequest.SetError("access_denied")
		rejectRequest.SetErrorDescription("The resource owner denied the request")

		request := oauth22.Hydra.AdminApi.RejectConsentRequest(context)
		request = request.ConsentChallenge(challenge)
		request = request.RejectRequest(rejectRequest)
		completedRejectConsentRequest, responseRejectConsent, errRejectConsent := request.Execute()
		if errRejectConsent != nil {
			// Error request to hydra OAuth admin API.
			if responseRejectConsent != nil {
				logger.Error("handlerConsentPost() - RejectConsentRequest() result:\n• err: %v\n• response: %v\n", errRejectConsent, responseRejectConsent)
			} else {
				logger.Error("handlerConsentPost() - RejectConsentRequest() result: %v\n", errRejectConsent)
			}

			context.AbortWithError(http.StatusInternalServerError, errRejectConsent)
			return
		}

		switch responseRejectConsent.StatusCode {
		case http.StatusOK:
			context.Redirect(http.StatusFound, completedRejectConsentRequest.RedirectTo)
		case http.StatusNotFound:
			// Accessing to response details
			// cast err to *client.GenericOpenAPIError object first and then
			// to your desired type
			notFound, ok := errRejectConsent.(*client.GenericOpenAPIError).Model().(client.JsonError)
			fmt.Println(ok)
			fmt.Println(*notFound.ErrorDescription)
		case http.StatusGone:
			responseDetail, ok := errRejectConsent.(*client.GenericOpenAPIError).Model().(client.RequestWasHandledResponse)
			fmt.Println(responseDetail, ok)
			fmt.Println("It's gone")
		default:
			fmt.Fprintf(os.Stderr, "Error when calling `AdminApi.RejectConsentRequest``: %v\n", errRejectConsent)
			fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", responseRejectConsent)
		}
	} else if submit != submitAllowAccess {
		context.String(http.StatusBadRequest, "Unexpected submit!")
		return
	}

	// Get consent request.
	requestGetConsent := oauth22.Hydra.AdminApi.GetConsentRequest(context)
	requestGetConsent = requestGetConsent.ConsentChallenge(challenge)
	consentRequestResponseData, responseGetConsent, errGetConsent := requestGetConsent.Execute()
	if errGetConsent != nil {
		// Error request to hydra OAuth admin API.
		if responseGetConsent != nil {
			logger.Error("handlerConsentPost() - GetConsentRequest() result:\n• err: %v\n• response: %v\n", errGetConsent, responseGetConsent)
		} else {
			logger.Error("handlerConsentPost() - GetConsentRequest() result:%v\n", errGetConsent)
		}

		context.AbortWithError(http.StatusInternalServerError, errGetConsent)
		return
	}

	var remember bool = false
	if context.PostForm("remember") != "" {
		remember = true
	}

	grantScope := context.PostFormArray("grant_scope")
	consentSession := client.ConsentRequestSession{}

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

	  // ORY Hydra checks if requested audiences are allowed by the client, so we can simply echo this.
	  grant_access_token_audience: body.requested_access_token_audience,

	  // This tells hydra to remember this consent request and allow the same client to request the same
	  // scopes from the same user, without showing the UI, in the future.
	  remember: Boolean(req.body.remember),

	  // When this "remember" sesion expires, in seconds. Set this to 0 so it will never expire.
	  remember_for: 3600
	*/
	var acceptConsentRequest client.AcceptConsentRequest
	acceptConsentRequest.SetGrantScope(grantScope)
	// TODO: do better value for acceptConsentRequest.SetSession()
	acceptConsentRequest.SetSession(consentSession)
	acceptConsentRequest.SetGrantAccessTokenAudience(consentRequestResponseData.GetRequestedAccessTokenAudience())
	acceptConsentRequest.SetRemember(remember)
	acceptConsentRequest.SetRememberFor(3600)

	requestAcceptConsent := oauth22.Hydra.AdminApi.AcceptConsentRequest(context)
	requestAcceptConsent = requestAcceptConsent.ConsentChallenge(challenge)
	requestAcceptConsent = requestAcceptConsent.AcceptConsentRequest(acceptConsentRequest)
	completedRequestAcceptConsent, responseAcceptConsent, errAcceptConsent := requestAcceptConsent.Execute()
	if errAcceptConsent != nil {
		// Error request to hydra OAuth admin API.
		if responseAcceptConsent != nil {
			logger.Error("handlerConsentPost() - AcceptConsentRequest() result:\n• err: %v\n• response: %v\n", errAcceptConsent, responseAcceptConsent)
		} else {
			logger.Error("handlerConsentPost() - AcceptConsentRequest() result: %v\n", errAcceptConsent)
		}

		context.AbortWithError(http.StatusInternalServerError, errAcceptConsent)
		return
	}

	context.Redirect(http.StatusFound, completedRequestAcceptConsent.RedirectTo)
}

func handlerCallback(context *gin.Context) {
	error := context.Query("error")
	if error != "" {
		logger.Error("handlerCallback() - query got error: %v\n", error)

		// Render result html.
		context.HTML(http.StatusOK, "error.html", gin.H{
			"Name":        error,
			"Description": context.Query("error_description"),
			"Hint":        context.Query("error_hint"),
			"Debug":       context.Query("error_debug"),
		})
		return
	}

	code := context.Query("code")
	token, err := oauth22.ConfOAuth2.Exchange(context, code)
	if err != nil {
		logger.Error("handlerCallback() - Unable to exchange code for token: %s\n", err)

		// Render result html.
		context.HTML(http.StatusOK, "error.html", gin.H{
			"Name": err.Error(),
		})
		return
	}

	idt := token.Extra("id_token")
	fmt.Printf("Access Token:\n\t%s\n", token.AccessToken)
	fmt.Printf("Refresh Token:\n\t%s\n", token.RefreshToken)
	fmt.Printf("Expires in:\n\t%s\n", token.Expiry.Format(time.RFC3339))
	idToken := fmt.Sprintf("%v", idt)
	fmt.Printf("ID Token:\n\t%s\n\n", idToken)

	// SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	context.SetCookie("access_token", token.AccessToken, 0, "", "", true, true)
	context.SetCookie("refresh_token", token.RefreshToken, 0, "", "", true, true)
	context.SetCookie("access_token_expires_in", token.Expiry.Format(time.RFC3339), 0, "", "", true, true)
	context.SetCookie("id_token", idToken, 0, "", "", true, true)

	// Redirect to main page.
	context.Redirect(http.StatusFound, pathRoot)

	// TODO: Get host data from config.
	//consentProto := "http"
	//consentHost := "127.0.0.1"
	//consentPort := 3000
	//consentURL := fmt.Sprintf("%s://%s:%d", consentProto, consentHost, consentPort)

	//// Render result html.
	//context.HTML(http.StatusOK, "callback.html", gin.H{
	//	"AccessToken":  token.AccessToken,
	//	"RefreshToken": token.RefreshToken,
	//	"Expiry":       token.Expiry.Format(time.RFC1123),
	//	"IDToken":      idToken,
	//	"BackURL":      consentURL,
	//})
}

func handlerLogoutGet(context *gin.Context) {
	challenge := context.Query("logout_challenge")
	if challenge == "" {
		context.String(http.StatusBadRequest, "Expected a login challenge to be set but received none.")
		return
	}

	// TODO: csrfToken for forms.
	// Render home html with auth url.
	context.HTML(http.StatusOK, "logout.html", gin.H{
		"csrfToken": "",
		"challenge": challenge,
		"action":    pathLogout,
	})
}

func handlerLogoutPost(context *gin.Context) {
	challenge := context.PostForm("challenge")
	if challenge == "" {
		context.String(http.StatusBadRequest, "Expected a login challenge to be set but received none.")
		return
	}

	submit := context.PostForm("submit")
	if submit == submitNo {
		// Reject logout request.
		request := oauth22.Hydra.AdminApi.RejectLogoutRequest(context)
		request = request.LogoutChallenge(challenge)
		responseRejectLogout, errRejectLogout := request.Execute()
		if errRejectLogout != nil {
			// Error request to hydra OAuth admin API.
			if responseRejectLogout != nil {
				logger.Error("handlerLogoutPost() - RejectLogoutRequest() result:\n• err: %v\n• response: %v\n", errRejectLogout, responseRejectLogout)
			} else {
				logger.Error("handlerLogoutPost() - RejectLogoutRequest() result: %v\n", errRejectLogout)
			}

			context.AbortWithError(http.StatusInternalServerError, errRejectLogout)
			return
		}

		// TODO: get redirect home page from config.
		// Redirect to main page.
		context.Redirect(http.StatusOK, pathRoot)
	} else if submit != submitYes {
		context.String(http.StatusBadRequest, "Unexpected submit!")
		return
	}

	// Accept logout request.
	requestAcceptLogout := oauth22.Hydra.AdminApi.AcceptLogoutRequest(context)
	requestAcceptLogout = requestAcceptLogout.LogoutChallenge(challenge)
	completedRequestAcceptLogout, responseAcceptLogout, errAcceptLogout := requestAcceptLogout.Execute()
	if errAcceptLogout != nil {
		// Error request to hydra OAuth admin API.
		if responseAcceptLogout != nil {
			logger.Error("handlerLogoutPost() - AcceptLogoutRequest() result:\n• err: %v\n• response: %v\n", errAcceptLogout, responseAcceptLogout)
		} else {
			logger.Error("handlerLogoutPost() - AcceptLogoutRequest() result: %v\n", errAcceptLogout)
		}

		context.AbortWithError(http.StatusInternalServerError, errAcceptLogout)
		return
	}

	context.Redirect(http.StatusFound, completedRequestAcceptLogout.RedirectTo)
}

func handlerLogoutBackchannel(context *gin.Context) {

}

func handlerLogoutFrontchannel(context *gin.Context) {
	//var accessToken, _ = GetCookieValue(context.Request, "access_token")
	//var accessTokenExpiresIn, _ = GetCookieValue(context.Request, "access_token_expires_in")
	//var refreshToken, _ = GetCookieValue(context.Request, "refresh_token")
	//var idToken, _ = GetCookieValue(context.Request, "id_token")

	// Delete tokens from storage.
	coockie.Remove(context.Writer, "access_token")
	coockie.Remove(context.Writer, "access_token_expires_in")
	coockie.Remove(context.Writer, "refresh_token")
	coockie.Remove(context.Writer, "id_token")
}
