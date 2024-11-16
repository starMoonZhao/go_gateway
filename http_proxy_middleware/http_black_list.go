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
func HTTPBlackListMiddleware() gin.HandlerFunc {
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

		//根据serviceDetail获取ip黑名单
		blackIpList := []string{}
		if serviceDetail.AccessControl.BlackList != "" {
			blackIpList = strings.Split(serviceDetail.AccessControl.BlackList, ",")
		}

		//如果已有白名单 白名单规则优先于黑名单 直接通过即可
		if serviceDetail.AccessControl.OpenAuth == 1 && len(whiteIpList) == 0 && len(blackIpList) > 0 {
			//如果ip在黑名单中 直接返回请求
			if public.InStringSlice(blackIpList, c.ClientIP()) {
				middleware.ResponseError(c, 9002, errors.New(fmt.Sprintf("%s in black ip list.", c.ClientIP())))
				//中断中间件传递链
				c.Abort()
				return
			}
		}

		//传递到下一中间件
		c.Next()
	}
}
