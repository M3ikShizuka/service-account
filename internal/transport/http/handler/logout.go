package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"service-account/internal/transport/http/coockie"
	"service-account/internal/transport/http/response"
)

func (h *Handler) logoutGet(context *gin.Context) {
	challenge := context.Query("logout_challenge")
	if challenge == "" {
		response.AbortMessage(context, http.StatusBadRequest, "handlerLogoutGet(): Expected a logout challenge to be set but received none.")
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

func (h *Handler) logoutPost(context *gin.Context) {
	challenge := context.PostForm("challenge")
	if challenge == "" {
		response.AbortMessage(context, http.StatusBadRequest, "handlerLogoutPost(): Expected a logout challenge to be set but received none.")
		return
	}

	submit := context.PostForm("submit")
	if submit == submitNo {
		// Reject logout request.
		err := h.services.OAuth2.RejectLogoutRequest(context, challenge)
		if err != nil {
			// Error request to hydra OAuth admin API.
			response.AbortError(context, http.StatusInternalServerError, err)
			return
		}

		// Redirect to main page.
		context.Redirect(http.StatusFound, pathRoot)
		return
	} else if submit != submitYes {
		response.AbortMessage(context, http.StatusBadRequest, "Unexpected submit!")
		return
	}

	// Accept logout request.
	redirectTo, err := h.services.OAuth2.AcceptLogoutRequest(context, challenge)
	if err != nil {
		response.AbortError(context, http.StatusInternalServerError, err)
		return
	}

	context.Redirect(http.StatusFound, redirectTo)
}

func (h *Handler) logoutBackchannel(context *gin.Context) {

}

func (h *Handler) logoutFrontchannel(context *gin.Context) {
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
