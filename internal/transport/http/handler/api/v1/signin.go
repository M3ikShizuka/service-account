package v1

// Example: https://github.com/ory/hydra-consent-app-go/blob/master/main.go

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"service-account/internal/service"
	"service-account/internal/transport/http/response"
	"service-account/pkg/convert_to"
)

// signinGet godoc
// @Summary     Signin user
// @Description Get signin page
// @Tags        auth
// @Produce     html
// @Success     200 {object} object{error=string}
// @Success     302 {object} object{error=string}
// @Failure     400 {object} object{error=string}
// @Failure     500 {object} object{error=string}
// @Router      /signin [get]
func (h *HandlerAccountManagementAPI) signinGet(context *gin.Context) {
	// Signin sessions, prompt, max_age, id_token_hint
	// https://<hydra-public>:4444/oauth2/auth?prompt=signin&max_age=60&id_token_hint=...'
	// SRC: https://www.ory.sh/docs/hydra/concepts/signin#signin-sessions-prompt-max_age-id_token_hint
	// return Signin page.html or skip and auth user by login_challenge.
	// We can get login_challenge by sent id_token_hint to https://<hydra-public>:4444/oauth2/auth?id_token_hint=... for re-auth automaticly.
	// http://127.0.0.1:3000/signin?login_challenge=9d54379b39094ba283ebd5d361b9afe6
	// code 200
	challenge := context.Query("login_challenge")
	if challenge == "" {
		response.AbortMessage(context, http.StatusBadRequest, "signinGet(): Expected a signin challenge to be set but received none.")
		return
	}

	// Get signin request.
	signinRequestData, err := h.services.OAuth2.GetLoginRequest(context, challenge)
	if err != nil {
		response.AbortError(context, http.StatusInternalServerError, err)
		return
	}

	// If hydra was already able to authenticate the user, skip will be true, and we do not need to re-authenticate
	// the user.
	if signinRequestData.Skip {
		// You can apply logic here, for example update the number of times the user logged in.
		// ...

		// Now it's time to grant the signin request. You could also deny the request if something went terribly wrong
		// (e.g. your arch-enemy logging in...)

		// Accept signin.
		RedirectTo, err := h.services.OAuth2.AcceptLoginRequest(context, challenge, signinRequestData.Subject, true, 3600)
		if err != nil {
			response.AbortError(context, http.StatusInternalServerError, err)
			return
		}

		context.Redirect(http.StatusFound, RedirectTo)
		return
	}

	// Render signin html.
	// TODO: csrfToken for forms.
	context.HTML(http.StatusOK, "signin.html",
		gin.H{
			"csrfToken": "",
			"challenge": challenge,
			"action":    pathSignin,
			"hint":      signinRequestData.Hint,
		})
}

// signinPost godoc
// @Summary     Signin user
// @Description Signin user
// @Tags        auth
// @Produce     html
// @Success     302 {object} object{error=string}
// @Failure     400 {object} object{error=string}
// @Failure     500 {object} object{error=string}
// @Router      /signin [post]
func (h *HandlerAccountManagementAPI) signinPost(context *gin.Context) {
	// Check authN data
	// redirect to hydra public :4444/ouath2/auth
	// Code 302
	challenge := context.PostForm("challenge")
	if challenge == "" {
		response.AbortMessage(context, http.StatusBadRequest, "signinPost(): Expected a signin challenge to be set but received none.")
		return
	}

	submit := context.PostForm("submit")
	if submit == submitDenyAccess {
		// Reject signin request.
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
	inputUserData := &service.UserSignInInput{
		Email:    userEmail,
		Password: userPassword,
	}

	user, err := h.services.User.SignIn(context, inputUserData)
	if err != nil {
		var statusCode int
		switch {
		case errors.Is(err, service.ErrUserNotFound),
			errors.Is(err, service.ErrPasswordIncorrect):
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusInternalServerError
		}

		// Render signin html with error.
		// TODO: csrfToken for forms.
		context.HTML(statusCode, "signin.html",
			gin.H{
				"csrfToken": "",
				"challenge": challenge,
				"action":    pathSignin,
				"error":     err.Error(),
			},
		)
		return
	}

	// Get signin request.
	if _, err = h.services.OAuth2.GetLoginRequest(context, challenge); err != nil {
		response.AbortError(context, http.StatusInternalServerError, err)
		return
	}

	// Remember auth signin session?
	var remember bool
	if context.PostForm("remember") != "" {
		remember = true
	}

	// Accept signin request.
	redirectTo, err := h.services.OAuth2.AcceptLoginRequest(context, challenge, convert_to.ToString(user.Id), remember, 3600)
	if err != nil {
		response.AbortError(context, http.StatusInternalServerError, err)
		return
	}

	context.Redirect(http.StatusFound, redirectTo)
}
