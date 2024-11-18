package controller

import (
	"encoding/base64"
	"github.com/dgrijalva/jwt-go"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/dto"
	"github.com/starMoonZhao/go_gateway/middleware"
	"github.com/starMoonZhao/go_gateway/public"
	"strings"
	"time"
)

type OAuthController struct {
}

func OAuthRegister(group *gin.RouterGroup) {
	oauthController := &OAuthController{}
	//注册路由
	group.POST("/tokens", oauthController.Tokens)
}

// Tokens godoc
// @Summary 获取token
// @Description 获取token
// @Tags OAUTH
// @ID /oauth/tokens
// @Accept  json
// @Produce  json
// @Param body body dto.TokensInput true "body"
// @Success 200 {object} middleware.Response{data=dto.TokensOutput} "success"
// @Router /oauth/tokens [post]
func (oauthController *OAuthController) Tokens(c *gin.Context) {
	tokensInput := &dto.TokensInput{}
	if err := tokensInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 6001, err)
		return
	}

	//获取请求头中的用户信息 格式：Authorization:Basic base64编码
	headerSpilt := strings.Split(c.GetHeader("Authorization"), " ")
	if len(headerSpilt) != 2 {
		middleware.ResponseError(c, 6002, errors.New("header Authorization format error"))
		return
	}

	//step1 取出app_id secret
	//step2 生成app_list
	//step3 匹配app_id
	//step4 基于jwt生成token
	//step5 生成output
	userInfo, err := base64.StdEncoding.DecodeString(headerSpilt[1])
	if err != nil {
		middleware.ResponseError(c, 6003, err)
		return
	}

	userInfoSplit := strings.Split(string(userInfo), ":")
	if len(userInfoSplit) != 2 {
		middleware.ResponseError(c, 6004, errors.New("UserInfo format error"))
		return
	}
	appList := dao.AppManegerHandler.GetAppList()
	for _, appItem := range appList {
		if appItem.APPID == userInfoSplit[0] && appItem.Secret == userInfoSplit[1] {
			claims := jwt.StandardClaims{
				Issuer:    appItem.APPID,
				ExpiresAt: time.Now().Add(public.JwtExpires * time.Second).In(lib.TimeLocation).Unix(),
			}
			token, err := public.JwtEncode(claims)
			if err != nil {
				middleware.ResponseError(c, 6005, err)
				return
			}
			output := &dto.TokensOutput{
				ExpiresIn:   public.JwtExpires,
				Scope:       "read_write",
				TokenType:   "Bearer",
				AccessToken: token,
			}
			middleware.ResponseSuccess(c, output)
			return
		}
	}

	middleware.ResponseError(c, 6006, errors.New("app info not found"))
}
