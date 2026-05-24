package controller

import (
	"errors"
	"net/http"

	"blog-server/internal/dto"
	"blog-server/internal/service"

	"github.com/gin-gonic/gin"
)

func (ctl *AuthController) ListOperationLogs(c *gin.Context) {
	page, pageSize, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(40000, "请求参数错误"))
		return
	}

	res, err := ctl.auth.ListOperationLogs(c.Request.Context(), c.GetHeader("Authorization"), page, pageSize)
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
