package controller

import (
	"encoding/json"
	"fmt"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/starMoonZhao/go_gateway/dao"
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
	group.PUT("/change_pwd", adminController.AdminChangePwd)
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

// AdminChangePwd godoc
// @Summary 管理员登录密码修改
// @Description 管理员登录密码修改
// @Tags 管理员登录密码修改接口
// @ID /admin/change_pwd
// @Accept  json
// @Produce  json
// @Param body body dto.AdminChangePwdInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin/change_pwd [put]
func (adminController *AdminController) AdminChangePwd(c *gin.Context) {
	//校验登录信息是否合法并解析传参
	adminChangePwdInput := &dto.AdminChangePwdInput{}
	if err := adminChangePwdInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//获取session客户端
	sessions := sessions.Default(c)
	sessiosInfo := sessions.Get(public.AdminSessionInfoKey)
	//取出用户session信息
	adminsessionInfo := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(fmt.Sprint(sessiosInfo)), adminsessionInfo); err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2003, err)
	}
	//查询用户信息 Admin
	admin := &dao.Admin{Id: adminsessionInfo.Id}
	err = admin.Find(c, tx)
	if err != nil {
		middleware.ResponseError(c, 2004, err)
		return
	}

	//获取新的加盐密码
	newEncodeSaltPwd := public.EncodeSaltPassword(admin.Salt, adminChangePwdInput.Password)

	//将新密码保存到数据库
	admin.Password = newEncodeSaltPwd
	err = admin.Save(c, tx)
	if err != nil {
		middleware.ResponseError(c, 2005, err)
		return
	}

	middleware.ResponseSuccess(c, "")
}
