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
	group.POST("/logout", adminLoginController.AdminLogout)

}

// AdminLogin godoc
// @Summary 管理员登录
// @Description 管理员登录
// @Tags 管理员登录
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
		middleware.ResponseError(c, 1011, err)
		return
	}
	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 1012, err)
		return
	}
	//查询登录人员信息
	admin := &dao.Admin{UserName: adminLoginInput.UserName}
	err = admin.Find(c, tx)
	if err != nil {
		middleware.ResponseError(c, 1013, err)
		return
	}
	//获取加盐密码编码->密码校验
	inputPassword := public.EncodeSaltPassword(admin.Salt, adminLoginInput.Password)
	if inputPassword != admin.Password {
		middleware.ResponseError(c, 1014, errors.New("密码错误"))
		return
	}

	//设置session
	adminSessionInfo := &dto.AdminSessionInfo{Id: admin.ID, UserName: admin.UserName, LoginTime: time.Now()}
	//将session转换为json
	adminSessionInfoJson, err := json.Marshal(adminSessionInfo)
	if err != nil {
		middleware.ResponseError(c, 1015, err)
		return
	}
	//获取session客户端
	sessions := sessions.Default(c)
	sessions.Set(public.AdminSessionInfoKey, string(adminSessionInfoJson))
	err = sessions.Save()
	if err != nil {
		middleware.ResponseError(c, 1016, err)
		return
	}
	outParam := &dto.AdminLoginOutput{Token: adminLoginInput.UserName}
	middleware.ResponseSuccess(c, outParam)
}

// AdminLogout godoc
// @Summary 管理员登出
// @Description 管理员登出
// @Tags 管理员登录
// @ID /admin_login/logout
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/logout [post]
func (adminLoginController *AdminLoginController) AdminLogout(c *gin.Context) {
	//获取session客户端
	sessions := sessions.Default(c)
	sessions.Delete(public.AdminSessionInfoKey)
	if err := sessions.Save(); err != nil {
		middleware.ResponseError(c, 1021, err)
		return
	}
	middleware.ResponseSuccess(c, "")
}
