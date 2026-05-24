package controller

import (
	"errors"
	"net/http"
	"strconv"

	"blog-server/internal/dto"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	auth *service.AuthService
}

func NewAuthController(auth *service.AuthService) *AuthController {
	return &AuthController{auth: auth}
}

func parseID(value string) (uint64, error) {
	return strconv.ParseUint(value, 10, 64)
}

func parsePagination(c *gin.Context) (int, int, error) {
	page := 1
	pageSize := 10

	if value := c.Query("page"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			return 0, 0, errors.New("invalid page")
		}
		page = parsed
	}
	if value := c.Query("pageSize"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			return 0, 0, errors.New("invalid pageSize")
		}
		pageSize = parsed
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize, nil
}

func (ctl *AuthController) ensureAuthenticated(c *gin.Context) bool {
	err := ctl.auth.EnsureAuthenticated(c.Request.Context(), c.GetHeader("Authorization"))
	if err == nil {
		return true
	}
	if errors.Is(err, service.ErrInvalidToken) {
		c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
		return false
	}
	c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
	return false
}
