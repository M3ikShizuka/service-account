package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"service-account/internal/transport/http/response"
	"service-account/pkg/logger"
	"time"
)

func (h *Handler) consentGet(context *gin.Context) {
	challenge := context.Query("consent_challenge")
	if challenge == "" {
		response.AbortMessage(context, http.StatusBadRequest, "handlerConsentGet(): Expected a consent challenge to be set but received none.")
		return
	}

	// Get consent request.
	getConsentData, err := h.services.OAuth2.GetConsentRequest(context, challenge)
	if err != nil {
		response.AbortError(context, http.StatusInternalServerError, err)
		return
	}

	if getConsentData.Skip {
		// You can apply logic here, for example grant another scope, or do whatever...
		// ...

		// Now it's time to grant the consent request. You could also deny the request if something went terribly wrong
		/*
			{
			// We can grant all scopes that have been requested - hydra already checked for us that no additional scopes
			// are requested accidentally.
			grant_scope: body.requested_scope,

			// ORY hydra checks if requested audiences are allowed by the client, so we can simply echo this.
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

		// Accept consent request.
		redirectTo, err := h.services.OAuth2.AcceptConsentRequest(context, challenge, getConsentData.RequestedScope, getConsentData.RequestedAccessTokenAudience, true, 3600)
		if err != nil {
			response.AbortError(context, http.StatusInternalServerError, err)
			return
		}

		context.Redirect(http.StatusFound, redirectTo)
		return
	}

	// Declared an empty map interface
	var clientData map[string]interface{}
	// Unmarshal or Decode the JSON to the interface.
	json.Unmarshal(getConsentData.ClientData, &clientData)

	// Render consent html.
	// TODO: csrfToken for forms.
	context.HTML(http.StatusOK, "consent.html",
		gin.H{
			"csrfToken": "",
			"challenge": challenge,
			// We have a bunch of data available from the response, check out the API docs to find what these values mean
			// and what additional data you have available.
			"requested_scope": getConsentData.RequestedScope,
			"user":            getConsentData.Subject,
			"client":          clientData,
			"action":          "/consent",
		})
}

func (h *Handler) consentPost(context *gin.Context) {
	challenge := context.PostForm("challenge")
	if challenge == "" {
		response.AbortMessage(context, http.StatusBadRequest, "handlerConsentPost(): Expected a consent challenge to be set but received none.")
		return
	}

	submit := context.PostForm("submit")
	if submit == submitDenyAccess {
		// Reject consent request.
		redirectTo, err := h.services.OAuth2.RejectConsentRequest(context, challenge, "access_denied", "The resource owner denied the request")
		if err != nil {
			response.AbortError(context, http.StatusInternalServerError, err)
			return
		}

		context.Redirect(http.StatusFound, redirectTo)
		return
	} else if submit != submitAllowAccess {
		response.AbortMessage(context, http.StatusBadRequest, "Unexpected submit!")
		return
	}

	// Get consent request.
	getConsentData, err := h.services.OAuth2.GetConsentRequest(context, challenge)
	if err != nil {
		response.AbortError(context, http.StatusInternalServerError, err)
		return
	}

	// Remember auth consent session?
	var remember bool
	if context.PostForm("remember") != "" {
		remember = true
	}

	grantScope := context.PostFormArray("grant_scope")

	// Accept consent request.
	redirectTo, err := h.services.OAuth2.AcceptConsentRequest(context, challenge, grantScope, getConsentData.RequestedAccessTokenAudience, remember, 3600)
	if err != nil {
		response.AbortError(context, http.StatusInternalServerError, err)
		return
	}

	context.Redirect(http.StatusFound, redirectTo)
}

func (h *Handler) callback(context *gin.Context) {
	errorParam := context.Query("error")
	if errorParam != "" {
		logger.Error("handlerCallback() - query got error",
			logger.String("error", errorParam),
		)

		// Render result html.
		context.HTML(http.StatusOK, "error.html", gin.H{
			"Name":        errorParam,
			"Description": context.Query("error_description"),
			"Hint":        context.Query("error_hint"),
			"Debug":       context.Query("error_debug"),
		})
		return
	}

	// Retrieve tokens.
	code := context.Query("code")
	token, err := h.services.OAuth2.TokenExchange(context, code)
	if err != nil {
		logger.Error("handlerCallback() - Unable to exchange code for token",
			logger.NamedError("error", err),
		)

		// Render result html.
		context.HTML(http.StatusOK, "error.html", gin.H{
			"Name": err.Error(),
		})
		return
	}

	logger.Debug("handlerCallback() got tokens",
		logger.String("Access Token", token.AccessToken),
		logger.String("Refresh Token", token.RefreshToken),
		logger.String("Expires in", token.Expiry.Format(time.RFC3339)),
		logger.String("ID Token", token.IdToken),
	)

	// Save tokens in cookies.
	// SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	context.SetCookie("access_token", token.AccessToken, 0, "", "", true, true)
	context.SetCookie("refresh_token", token.RefreshToken, 0, "", "", true, true)
	context.SetCookie("access_token_expires_in", token.Expiry.Format(time.RFC3339), 0, "", "", true, true)
	context.SetCookie("id_token", token.IdToken, 0, "", "", true, true)

	// Redirect to main page.
	context.Redirect(http.StatusFound, pathRoot)

	// Render result html.
	//context.HTML(http.StatusOK, "callback.html", gin.H{
	//	"AccessToken":  token.AccessToken,
	//	"RefreshToken": token.RefreshToken,
	//	"Expiry":       token.Expiry.Format(time.RFC1123),
	//	"IDToken":      token.IdToken,
	//	"BackURL":      h.services.Config.OAuth2.ConsentURL,
	//})

	// Don't forget to check retrieved ID Token from client-device.
	/*
		JSON Web Token validation
		You can validate JSON Web Tokens issued by Ory Hydra by pointing your jwt library
		(for example node-jwks-rsa) to http://ory-hydra-public-api/.well-known/jwks.json.
		All necessary keys are available there.
	*/
}
