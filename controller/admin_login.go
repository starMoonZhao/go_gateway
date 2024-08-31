package controller

import (
	"encoding/json"
	"errors"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/dto"
	"github.com/starMoonZhao/go_gateway/middleware"
	"github.com/starMoonZhao/go_gateway/public"
	"time"
)

type AdminLoginController struct {
}

func AdminLoginRegister(group *gin.RouterGroup) {
	adminLoginController := &AdminLoginController{}
	//注册路由
	group.POST("/login", adminLoginController.AdminLogin)
}

// AdminLogin godoc
// @Summary 管理员登录
// @Description 管理员登录
// @Tags 管理员登录接口
// @ID /admin_login/login
// @Accept  json
// @Produce  json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [post]
func (adminLoginController *AdminLoginController) AdminLogin(c *gin.Context) {
	//校验登录信息是否合法
	adminLoginInput := &dto.AdminLoginInput{}
	if err := adminLoginInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 1001, err)
		return
	}
	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 1002, err)
		return
	}
	//查询登录人员信息
	admin := &dao.Admin{}
	err = admin.Find(c, tx, &dao.Admin{UserName: adminLoginInput.UserName})
	if err != nil {
		middleware.ResponseError(c, 1003, err)
		return
	}
	//获取加盐密码编码->密码校验
	inputPassword := public.EncodeSaltPassword(admin.Salt, adminLoginInput.Password)
	if inputPassword != admin.Password {
		middleware.ResponseError(c, 1004, errors.New("密码错误"))
		return
	}

	//设置session
	adminSessionInfo := &dto.AdminSessionInfo{Id: admin.Id, UserName: admin.UserName, LoginTime: time.Now()}
	//将session转换为json
	adminSessionInfoJson, err := json.Marshal(adminSessionInfo)
	if err != nil {
		middleware.ResponseError(c, 1005, err)
		return
	}
	//获取session客户端
	sessions := sessions.Default(c)
	sessions.Set(public.AdminSessionInfoKey, string(adminSessionInfoJson))
	err = sessions.Save()
	if err != nil {
		middleware.ResponseError(c, 1006, err)
		return
	}
	outParam := &dto.AdminLoginOutput{Token: adminLoginInput.UserName}
	middleware.ResponseSuccess(c, outParam)
}
