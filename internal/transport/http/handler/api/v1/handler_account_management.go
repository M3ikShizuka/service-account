package v1

import (
	"github.com/gin-gonic/gin"
	"service-account/internal/service"
)

type HandlerAccountManagementAPI struct {
	services *service.Services
}

func NewHandlerAPIv1(services *service.Services) *HandlerAccountManagementAPI {
	return &HandlerAccountManagementAPI{
		services: services,
	}
}

func (h *HandlerAccountManagementAPI) Init(router *gin.RouterGroup) {
	v1 := router.Group("/v1")
	{
		h.initHandlersAccountManagement(v1)
	}
}

func (h *HandlerAccountManagementAPI) initHandlersAccountManagement(router *gin.RouterGroup) {
	router.GET(pathUser, userGet)
	router.POST(pathUser, userPost)
}
