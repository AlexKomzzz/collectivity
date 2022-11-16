package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader = "Authorization"
	// userCtx             = "userId"
)

func (h *Handler) userIdentity(c *gin.Context) (int, error) {

	// выделение из заголовка поля "Authorization"
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		// newErrorResponse(c, http.StatusUnauthorized, "empty auth header")
		return -1, errors.New("empty auth header")
	}

	return h.service.ValidToken(header)
}

/*
func getUserId(c *gin.Context) (int, error) {
	id, ok := c.Get(userCtx)
	if !ok {
		//newErrorResponse(c, http.StatusInternalServerError, "user id not found")
		return 0, errors.New("user id not found")
	}

	idInt, ok := id.(int)
	if !ok {
		//newErrorResponse(c, http.StatusInternalServerError, "user id is of invalid type")
		return 0, errors.New("user id is of invalid type")
	}

	return idInt, nil
}*/
