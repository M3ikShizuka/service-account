package response

import (
	"github.com/gin-gonic/gin"
	"service-account/pkg/logger"
)

func AbortError(context *gin.Context, statusCode int, err error) {
	logger.Error(err.Error())
	_ = context.AbortWithError(statusCode, err)
}

func AbortMessage(context *gin.Context, statusCode int, msg string) {
	logger.Error(msg)
	context.String(statusCode, msg)
}
