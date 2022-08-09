package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"service-account/internal/service/authz/oauth2"
	"service-account/internal/transport/http/coockie"
)

const pathRoot = "/"

func (handler *Handler) initGeneralRoutes(router *gin.Engine) {
	// Main/Home page
	router.GET(pathRoot, handler.rootGet)
	// 404
	router.NoRoute(handler.notFound)
}

func (handler *Handler) rootGet(context *gin.Context) {
	// Login sessions, prompt, max_age, id_token_hint
	// https://<hydra-public>:4444/oauth2/auth?prompt=login&max_age=60&id_token_hint=...'
	// SRC: https://www.ory.sh/docs/hydra/concepts/login#login-sessions-prompt-max_age-id_token_hint
	// We can get login_challenge by sent id_token_hint for re-auth automaticly.
	var accessToken, _ = coockie.GetValue(context.Request, "access_token")

	var isAuth bool
	if accessToken != "" {
		// Get introspect token request.
		requestIntrospectToken := oauth2.Hydra.AdminApi.IntrospectOAuth2Token(context)
		requestIntrospectToken = requestIntrospectToken.Token(accessToken)
		loginRequestResponseData, responseIntrospectToken, errIntrospectToken := requestIntrospectToken.Execute()
		if errIntrospectToken != nil {
			// Error request to hydra OAuth admin API.
			if responseIntrospectToken != nil {
				logger.Error("handlerRootGet() - IntrospectTokenRequest() result:\n• err: %v\n• response: %v\n", errIntrospectToken, responseIntrospectToken)
			} else {
				logger.Error("handlerRootGet() - IntrospectTokenRequest() result:%v\n", errIntrospectToken)
			}

			context.AbortWithError(http.StatusInternalServerError, errIntrospectToken)
			return
		}

		/*
			{
			  "active": true,
			  "aud": [
			    "string"
			  ],
			  "client_id": "string",
			  "exp": 0,
			  "ext": {},
			  "iat": 0,
			  "iss": "string",
			  "nbf": 0,
			  "obfuscated_subject": "string",
			  "scope": "string",
			  "sub": "string",
			  "token_type": "string",
			  "token_use": "string",
			  "username": "string"
			}
		*/

		isAuth = loginRequestResponseData.GetActive()
		if isAuth == true {
			var idToken, err = coockie.GetValue(context.Request, "id_token")
			if err == nil {
				logoutUrl := oauth2.GenerateLogoutURL(oauth2.LogoutUrlTemplate, idToken, "", "")

				// Token is valid.
				// Render home html with auth info.
				context.HTML(http.StatusOK, "index.html", gin.H{
					"isAuth": isAuth,
					"URL":    logoutUrl,
				})
				return
			}

			// Can't get idToken.
			logger.Error("handlerRootGet() - GetCookieValue(id_token) result:%v\n", err)
		}
	}

	// Token is invalid!
	// Render home html with auth url.
	context.HTML(http.StatusOK, "index.html", gin.H{
		"isAuth": isAuth,
		"URL":    oauth2.AuthCodeUrl,
	})
}

func (handler *Handler) notFound(context *gin.Context) {
	context.Writer.WriteHeader(http.StatusNotFound)
	context.Writer.Write([]byte("Page not found! dontknown！"))
}
