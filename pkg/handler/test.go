package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) test(c *gin.Context) {
	c.HTML(http.StatusOK, "ex.html", gin.H{
		"req": "Alex",
	})
}
