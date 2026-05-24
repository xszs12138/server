package controller

import (
	"errors"
	"net/http"

	"blog-server/internal/dto"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

func (ctl *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.auth.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrRegistrationClosed) {
			c.JSON(http.StatusConflict, dto.Error(40900, "注册已关闭"))
			return
		}
		if errors.Is(err, service.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, dto.Error(40900, "用户名或邮箱已存在"))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	ip := c.ClientIP()
	res, err := ctl.auth.Login(c.Request.Context(), req, dto.OperationMeta{
		IP:        ip,
		Region:    service.ResolveIPRegion(ip),
		UserAgent: c.GetHeader("User-Agent"),
	})
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

func (ctl *AuthController) Refresh(c *gin.Context) {
	res, err := ctl.auth.Refresh(c.Request.Context(), c.GetHeader("Authorization"))
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}

func (ctl *AuthController) Logout(c *gin.Context) {
	ip := c.ClientIP()
	if err := ctl.auth.Logout(c.Request.Context(), c.GetHeader("Authorization"), dto.OperationMeta{
		IP:        ip,
		Region:    service.ResolveIPRegion(ip),
		UserAgent: c.GetHeader("User-Agent"),
	}); err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(true))
}

func (ctl *AuthController) Me(c *gin.Context) {
	res, err := ctl.auth.Me(c.Request.Context(), c.GetHeader("Authorization"))
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			c.JSON(http.StatusUnauthorized, dto.Error(40100, "未登录或登录失效"))
			return
		}

		c.JSON(http.StatusInternalServerError, dto.Error(50000, "服务端错误"))
		return
	}

	c.JSON(http.StatusOK, dto.OK(res))
}
