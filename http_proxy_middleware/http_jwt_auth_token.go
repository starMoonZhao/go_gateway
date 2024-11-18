package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/middleware"
	"github.com/starMoonZhao/go_gateway/public"
	"strings"
)

// 限流器中间件
func HTTPJwtAuthTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//获取上游服务信息
		serviceInterface, ok := c.Get("service")
		if !ok {
			middleware.ResponseError(c, 9001, errors.New("service not found"))
			//中断中间件传递链
			c.Abort()
			return
		}
		//类型转换
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		//step1 获取请求头中的授权信息
		//step2 解密授权信息
		//step3 将租户信息存入gin.context
		token := strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")

		appMatched := false
		if token != "" {
			claims, err := public.JwtDecode(token)
			if err != nil {
				middleware.ResponseError(c, 9002, err)
				//中断中间件传递链
				c.Abort()
				return
			}
			appList := dao.AppManegerHandler.GetAppList()
			for _, app := range appList {
				if app.APPID == claims.Issuer {
					c.Set("app", app)
					appMatched = true
					break
				}
			}
		}

		if serviceDetail.AccessControl.OpenAuth == 1 && !appMatched {
			middleware.ResponseError(c, 9003, errors.New("access denied"))
			//中断中间件传递链
			c.Abort()
			return
		}
		//传递到下一中间件
		c.Next()
	}
}
