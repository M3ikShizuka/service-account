package v1

import (
	"github.com/gin-gonic/gin"
	"service-account/internal/service"
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

type HandlerAPIv1 struct {
	services *service.Services
}

func NewHandlerAPIv1(services *service.Services) *HandlerAPIv1 {
	return &HandlerAPIv1{
		services: services,
	}
}

func (h *HandlerAPIv1) Init(router *gin.RouterGroup) {
	v1 := router.Group("/v1")
	{
		h.initHandlersAuthentication(v1)
		h.initHandlersAccountManagment(v1)
	}
}

func (h *HandlerAPIv1) initHandlersAuthentication(router *gin.RouterGroup) {
	// Init router.
	// Log in
	router.GET(pathLogin, h.handlerLoginGet)
	router.POST(pathLogin, h.handlerLoginPost)
	// Consent
	router.GET(pathConsent, h.handlerConsentGet)
	router.POST(pathConsent, h.handlerConsentPost)
	// Callback
	router.GET(pathCallback, h.handlerCallback)
	// Logout
	router.GET(pathLogout, h.handlerLogoutGet)
	router.POST(pathLogout, h.handlerLogoutPost)
	router.Any(pathLogoutBackchannel, h.handlerLogoutBackchannel)
	router.GET(pathLogoutFrontchannel, h.handlerLogoutFrontchannel)
}

func (h *HandlerAPIv1) initHandlersAccountManagment(router *gin.RouterGroup) {
	router.GET(pathUser, handlerUserGet)
	router.POST(pathUser, handlerUserPost)
}
