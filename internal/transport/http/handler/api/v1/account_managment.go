package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"service-account/internal/repository"
	"service-account/internal/transport/http/coockie"
	"strconv"
)

func (HandlerAccountManagementAPI) convertStringToUserId(idStr string) (uint32, error) {
	userId64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}

	// Convert userId from string to uint32
	userId := uint32(userId64)

	return userId, nil
}

// userGet godoc
// @Summary     Get user info
// @Security 	ApiKeyAuth
// @Description get user by ID
// @Tags        user
// @Produce     json
// @Success     200 {object} object{user=object{id=uint32,username=string,email=string,date_registration=time.Time,date_last_online=time.Time}}
// @Failure     400 {object} object{error=string}
// @Failure     401 {object} object{error=string}
// @Failure     403 {object} object{error=string}
// @Failure     500 {object} object{error=string}
// @Param id   path int true "UserRepositoryGorm ID"
// @Router      /api/v1/users/{id} [get]
func (h *HandlerAccountManagementAPI) userGet(context *gin.Context) {
	// Get user id.
	userIdStr := context.Param("id")
	if userIdStr == "" {
		context.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": "Expected a user id to be set but received none.",
		})
		return
	}

	// Convert string to id.
	userId, err := h.convertStringToUserId(userIdStr)
	if err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": "UserRepositoryGorm id bad format.",
		})
	}

	// Get access_token from cookie.
	accessToken, _ := coockie.GetValue(context.Request, "access_token")
	if accessToken == "" {
		context.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": "Access Token is not present.",
		})
		return
	}

	// Get introspect token request.
	tokenIntrospection, err := h.services.OAuth2.IntrospectOAuth2Token(context, accessToken)
	if err != nil {
		// Error request to hydra OAuth admin API.
		context.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check authorization.
	isAuth := tokenIntrospection.Active
	if !isAuth {
		context.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": "Access Token is not active.",
		})
		return
	}

	if tokenIntrospection.Sub == nil {
		context.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": "The token's subject user id is in the wrong format.",
		})
		return
	}

	tokenUserId, err := h.convertStringToUserId(*tokenIntrospection.Sub)
	if err != nil {
		context.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": "The token's subject user id is in the bad format.",
		})
	}

	// Check if the user is the same user for whom we want to get information.
	// TODO: Or the user has administrator privileges.
	if userId != tokenUserId {
		context.IndentedJSON(http.StatusForbidden, gin.H{
			"error": "No permission.",
		})
		return
	}

	// Get user data.
	user, err := h.services.User.GetUserById(context, userId)
	if err != nil {
		var errorMessage string
		if errors.Is(err, repository.ErrRecordNotFound) {
			errorMessage = "UserRepositoryGorm not found."
		} else {
			errorMessage = err.Error()
		}

		context.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": errorMessage,
		})
		return
	}

	// Send success response.
	context.IndentedJSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":                user.Id,
			"username":          user.Username,
			"email":             user.Email,
			"date_registration": user.DateRegistration,
			"date_last_online":  user.DateLastOnline,
		},
	})
}
