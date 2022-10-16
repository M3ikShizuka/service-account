package handler

import (
	"github.com/gin-gonic/gin"
	"service-account/internal/service"
	"service-account/internal/transport/http/handler/api/v1"
)

const (
	submitDenyAccess  = "Deny access"
	submitLogIn       = "Log in"
	submitAllowAccess = "Allow access"
	submitNo          = "No"
	submitYes         = "Yes"
	// Paths
	pathRoot                      = "/"
	pathSignup             string = "/signup"
	pathLogin              string = "/login"
	pathConsent            string = "/consent"
	pathCallback           string = "/callback"
	pathLogout             string = "/logout"
	pathLogoutBackchannel  string = "/backchannel-logout"
	pathLogoutFrontchannel string = "/frontchannel-logout"
)

type Handler struct {
	services *service.Services
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) Init() *gin.Engine {
	router := gin.Default()
	// Init HTML Glob
	h.initHTMLGlob(router)
	// Init general routes.
	h.initGeneralRoutes(router)
	h.initHandlersAuthentication(router)
	// Init API
	h.initAPI(router)

	return router
}

func (h *Handler) initHTMLGlob(router *gin.Engine) {
	// template: https://gin-gonic.com/docs/examples/html-rendering/
	router.LoadHTMLGlob("./web/template/*")
}

func (h *Handler) initAPI(router *gin.Engine) {
	// Init API v1 h.
	handlersV1 := v1.NewHandlerAPIv1(h.services)
	api := router.Group("/api")
	{
		handlersV1.Init(api)
	}
}

func (h *Handler) initHandlersAuthentication(router *gin.Engine) {
	// Init router.
	// Log in
	router.GET(pathLogin, h.loginGet)
	router.POST(pathLogin, h.loginPost)
	// Consent
	router.GET(pathConsent, h.consentGet)
	router.POST(pathConsent, h.consentPost)
	// Callback
	router.GET(pathCallback, h.callback)
	// Logout
	router.GET(pathLogout, h.logoutGet)
	router.POST(pathLogout, h.logoutPost)
	router.Any(pathLogoutBackchannel, h.logoutBackchannel)
	router.GET(pathLogoutFrontchannel, h.logoutFrontchannel)
}
