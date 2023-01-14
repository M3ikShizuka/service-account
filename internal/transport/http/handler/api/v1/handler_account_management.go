package v1

import (
	"github.com/gin-gonic/gin"
	"service-account/internal/service"
)

const (
	submitDenyAccess  = "Deny access"
	submitLogIn       = "Log in"
	submitSignUp      = "Register"
	submitAllowAccess = "Allow access"
	submitNo          = "No"
	submitYes         = "Yes"
	// Paths
	pathRoot                      = "/"
	PathSignup             string = "/signup"
	pathSignin             string = "/signin"
	pathConsent            string = "/consent"
	pathCallback           string = "/callback"
	pathLogout             string = "/logout"
	pathLogoutBackchannel  string = "/backchannel-logout"
	pathLogoutFrontchannel string = "/frontchannel-logout"
	// Paths v1
	pathUser string = "/users"
)

type HandlerAccountManagementAPI struct {
	services *service.Services
}

func NewHandlerAccountManagementAPI(services *service.Services) *HandlerAccountManagementAPI {
	return &HandlerAccountManagementAPI{
		services: services,
	}
}

func (h *HandlerAccountManagementAPI) Init(router *gin.RouterGroup) {
	h.initHandlersAuthentication(router)
	v1 := router.Group("/api/v1")
	{
		h.initHandlersAccountManagement(v1)
	}
}

func (h *HandlerAccountManagementAPI) initHandlersAccountManagement(router *gin.RouterGroup) {
	user := router.Group(pathUser)
	{
		user.GET(":id", h.userGet)
	}
}

func (h *HandlerAccountManagementAPI) initHandlersAuthentication(router *gin.RouterGroup) {
	// Init router.
	// Sign in
	router.GET(pathSignin, h.signinGet)
	router.POST(pathSignin, h.signinPost)
	// Sign up
	router.GET(PathSignup, h.signupGet)
	router.POST(PathSignup, h.signupPost)
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
