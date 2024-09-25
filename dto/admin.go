package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/starMoonZhao/go_gateway/public"
	"time"
)

type AdminInfoOutput struct {
	Id           int64     `json:"id"`
	UserName     string    `json:"username"`
	LoginTime    time.Time `json:"login_time"`
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json:"roles"`
}

type AdminChangePwdInput struct {
	Password string `json:"password" form:"password" comment:"密码" example:"123456" validate:"required"`
}

// 校验用户输入的新密码
func (param *AdminChangePwdInput) BindValidParam(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}
