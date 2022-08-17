package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"service-account/internal/transport/http/coockie"
	"service-account/internal/transport/http/response"
	"service-account/pkg/logger"
)

func (h *Handler) initGeneralRoutes(router *gin.Engine) {
	// Main/Home page
	router.GET(pathRoot, h.rootGet)
	// 404
	router.NoRoute(h.notFound)
}

func (h *Handler) rootGet(context *gin.Context) {
	// Login sessions, prompt, max_age, id_token_hint
	// https://<hydra-public>:4444/oauth2/auth?prompt=login&max_age=60&id_token_hint=...'
	// SRC: https://www.ory.sh/docs/hydra/concepts/login#login-sessions-prompt-max_age-id_token_hint
	// We can get login_challenge by sent id_token_hint for re-auth automaticly.
	var accessToken, _ = coockie.GetValue(context.Request, "access_token")
	var isAuth bool

	if accessToken != "" {
		// Get introspect token request.
		tokenIntrospection, err := h.services.OAuth2.IntrospectOAuth2Token(context, accessToken)
		if err != nil {
			// Error request to hydra OAuth admin API.
			response.AbortError(context, http.StatusInternalServerError, err)
			return
		}

		isAuth = tokenIntrospection.Active
		if isAuth {
			// Get OpenID Token.
			var idToken, err = coockie.GetValue(context.Request, "id_token")
			if err == nil {
				logoutUrl := h.services.OAuth2.GenerateLogoutURL(idToken, "", "")

				// Token is valid.
				// Render home html with auth info.
				context.HTML(http.StatusOK, "index.html", gin.H{
					"isAuth": isAuth,
					"URL":    logoutUrl,
				})
				return
			}

			// Can't get OpenID.
			logger.Error("handlerRootGet() - GetCookieValue(id_token)\n",
				logger.NamedError("error", err),
			)
		}
	}

	// Token is invalid!
	// Render home html with auth url.
	context.HTML(http.StatusOK, "index.html", gin.H{
		"isAuth": isAuth,
		"URL":    h.services.OAuth2.GetAuthCodeUrl(),
	})
}

func (h *Handler) notFound(context *gin.Context) {
	context.Writer.WriteHeader(http.StatusNotFound)
	_, _ = context.Writer.Write([]byte("Page not found! dontknown！"))
}
