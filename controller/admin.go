package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/starMoonZhao/go_gateway/dto"
	"github.com/starMoonZhao/go_gateway/middleware"
	"github.com/starMoonZhao/go_gateway/public"
)

type AdminController struct {
}

func AdminRegister(group *gin.RouterGroup) {
	adminController := &AdminController{}
	//注册路由
	group.GET("/admin_info", adminController.AdminInfo)
}

// AdminInfo godoc
// @Summary 管理员登录信息查询
// @Description 管理员登录信息查询
// @Tags 管理员登录信息查询接口
// @ID /admin/admin_info
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.AdminInfoOutput} "success"
// @Router /admin/admin_info [get]
func (adminController *AdminController) AdminInfo(c *gin.Context) {
	//获取session客户端
	sessions := sessions.Default(c)
	sessiosInfo := sessions.Get(public.AdminSessionInfoKey)
	//取出用户session信息
	adminsessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessiosInfo)), adminsessionInfo); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	//封装输出信息
	out := &dto.AdminInfoOutput{
		Id:           adminsessionInfo.Id,
		UserName:     adminsessionInfo.UserName,
		LoginTime:    adminsessionInfo.LoginTime,
		Avatar:       "",
		Introduction: "",
		Roles:        []string{},
	}
	middleware.ResponseSuccess(c, out)
}
