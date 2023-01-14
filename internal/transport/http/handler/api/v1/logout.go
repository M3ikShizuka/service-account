package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"service-account/internal/transport/http/coockie"
	"service-account/internal/transport/http/response"
)

// logoutGet godoc
// @Summary     Logout user
// @Description Get logout page
// @Tags        auth
// @Produce     html
// @Success     200 {object} object{error=string}
// @Failure     400 {object} object{error=string}
// @Router      /logout [get]
func (h *HandlerAccountManagementAPI) logoutGet(context *gin.Context) {
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

// logoutPost godoc
// @Summary     Logout user
// @Description Logout user
// @Tags        auth
// @Produce     html
// @Success     302 {object} object{error=string}
// @Failure     400 {object} object{error=string}
// @Failure     500 {object} object{error=string}
// @Router      /logout [post]
func (h *HandlerAccountManagementAPI) logoutPost(context *gin.Context) {
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

// logoutBackchannel godoc
// @Summary     Logout user back channel
// @Description Logout user back channel delete tokens from storage.
// @Tags        auth
// @Router      /backchannel-logout [get]
func (h *HandlerAccountManagementAPI) logoutBackchannel(context *gin.Context) {

}

// logoutFrontchannel godoc
// @Summary     Logout user front channel
// @Description Logout user front channel delete tokens from storage.
// @Tags        auth
// @Router      /frontchannel-logout [get]
func (h *HandlerAccountManagementAPI) logoutFrontchannel(context *gin.Context) {
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
