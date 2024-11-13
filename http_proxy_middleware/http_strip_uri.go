package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/middleware"
	"github.com/starMoonZhao/go_gateway/public"
	"strings"
)

// StripUri
func HTTPStripUriMiddleware() gin.HandlerFunc {
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

		//如果需要StripUri并且HTTPRule规则是前缀匹配 才需要StripUri
		if serviceDetail.HTTPRule.NeedStripUri == 1 && serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL {
			//将请求path中的接入规则去除
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path, serviceDetail.HTTPRule.Rule, "", 1)
		}

		//传递到下一中间件
		c.Next()
	}
}
