package controller

import (
	"errors"
	"net/http"

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

func (ctl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.auth.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, dto.Error(40100, "用户名或密码错误"))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}
