package http_proxy_middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/middleware"
	"github.com/starMoonZhao/go_gateway/public"
	"strings"
)

// 请求路径重写规则
func HTTPWhiteListMiddleware() gin.HandlerFunc {
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

		//根据serviceDetail获取ip白名单
		whiteIpList := []string{}
		if serviceDetail.AccessControl.WhiteList != "" {
			whiteIpList = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}

		if serviceDetail.AccessControl.OpenAuth == 1 && len(whiteIpList) > 0 {
			//如果ip不在白名单中 直接返回请求
			if !public.InStringSlice(whiteIpList, c.ClientIP()) {
				middleware.ResponseError(c, 9002, errors.New(fmt.Sprintf("%s not in white ip list.", c.ClientIP())))
				//中断中间件传递链
				c.Abort()
				return
			}
		}

		//传递到下一中间件
		c.Next()
	}
}
