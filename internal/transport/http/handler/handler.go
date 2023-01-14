package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	_ "service-account/api"
	"service-account/internal/service"
	"service-account/internal/transport/http/handler/api/v1"
)

const (
	// Paths
	pathRoot = "/"
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
	// Init API
	h.initAPI(router)

	return router
}

func (h *Handler) initHTMLGlob(router *gin.Engine) {
	// template: https://gin-gonic.com/docs/examples/html-rendering/
	router.LoadHTMLGlob("./web/template/*")
}

func (h *Handler) initAPI(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// Init API v1 h.
	handlersV1 := v1.NewHandlerAccountManagementAPI(h.services)
	handlersV1.Init(&router.RouterGroup)
}
