package handler

// Example: https://github.com/ory/hydra-consent-app-go/blob/master/main.go

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"service-account/internal/transport/http/response"
)

func (h *Handler) loginGet(context *gin.Context) {
	// Login sessions, prompt, max_age, id_token_hint
	// https://<hydra-public>:4444/oauth2/auth?prompt=login&max_age=60&id_token_hint=...'
	// SRC: https://www.ory.sh/docs/hydra/concepts/login#login-sessions-prompt-max_age-id_token_hint
	// return Login page.html or skip and auth user by login_challenge.
	// We can get login_challenge by sent id_token_hint to https://<hydra-public>:4444/oauth2/auth?id_token_hint=... for re-auth automaticly.
	// http://127.0.0.1:3000/login?login_challenge=9d54379b39094ba283ebd5d361b9afe6
	// code 200
	challenge := context.Query("login_challenge")
	if challenge == "" {
		response.AbortMessage(context, http.StatusBadRequest, "loginGet(): Expected a login challenge to be set but received none.")
		return
	}

	// Get login request.
	loginRequestData, err := h.services.OAuth2.GetLoginRequest(context, challenge)
	if err != nil {
		response.AbortError(context, http.StatusInternalServerError, err)
		return
	}

	// If hydra was already able to authenticate the user, skip will be true, and we do not need to re-authenticate
	// the user.
	if loginRequestData.Skip {
		// You can apply logic here, for example update the number of times the user logged in.
		// ...

		// Now it's time to grant the login request. You could also deny the request if something went terribly wrong
		// (e.g. your arch-enemy logging in...)

		// Accept login.
		RedirectTo, err := h.services.OAuth2.AcceptLoginRequest(context, challenge, loginRequestData.Subject, true, 3600)
		if err != nil {
			response.AbortError(context, http.StatusInternalServerError, err)
			return
		}

		context.Redirect(http.StatusFound, RedirectTo)
		return
	}

	// Render login html.
	// TODO: csrfToken for forms.
	context.HTML(http.StatusOK, "login.html",
		gin.H{
			"csrfToken": "",
			"challenge": challenge,
			"action":    "/login",
			"hint":      loginRequestData.Hint,
		})
}

func (h *Handler) loginPost(context *gin.Context) {
	// Check authN data
	// redirect to hydra public :4444/ouath2/auth
	// Code 302
	challenge := context.PostForm("challenge")
	if challenge == "" {
		response.AbortMessage(context, http.StatusBadRequest, "loginPost(): Expected a login challenge to be set but received none.")
		return
	}

	submit := context.PostForm("submit")
	if submit == submitDenyAccess {
		// Reject login request.
		redirectTo, err := h.services.OAuth2.RejectLoginRequest(context, challenge, "access_denied", "The resource owner denied the request")
		if err != nil {
			// Error request to hydra OAuth admin API.
			response.AbortError(context, http.StatusInternalServerError, err)
			return
		}

		context.Redirect(http.StatusFound, redirectTo)
		return
	} else if submit != submitLogIn {
		response.AbortMessage(context, http.StatusBadRequest, "Unexpected submit!")
		return
	}

	// Check the user's credentials.
	var userEmail = context.PostForm("email")
	var userPassword = context.PostForm("password")

	// Check authentication.
	// TODO: call func from authN.
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
	_, err := h.services.OAuth2.GetLoginRequest(context, challenge)
	if err != nil {
		response.AbortError(context, http.StatusInternalServerError, err)
		return
	}

	// Remember auth login session?
	var remember bool
	if context.PostForm("remember") != "" {
		remember = true
	}

	// Accept login request.
	redirectTo, err := h.services.OAuth2.AcceptLoginRequest(context, challenge, userEmail, remember, 3600)
	if err != nil {
		response.AbortError(context, http.StatusInternalServerError, err)
		return
	}

	context.Redirect(http.StatusFound, redirectTo)
}
