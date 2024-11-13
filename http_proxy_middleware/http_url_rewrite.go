package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/middleware"
	"regexp"
	"strings"
)

// 请求路径重写规则
func HTTPURLRewriteMiddleware() gin.HandlerFunc {
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

		//根据serviceDetail获取url rewrite转换规则 格式: ^/gatekeeper/test_service(.*) $1
		for _, item := range strings.Split(serviceDetail.HTTPRule.UrlRewrite, ",") {
			//规则格式：匹配原url正则 生成目标url的表达式
			items := strings.Split(item, " ")
			if len(items) != 2 {
				continue
			}

			//根据参数1生成正则
			regexp, err := regexp.Compile(items[0])
			if err != nil {
				continue
			}
			//根据参数2生成目标url
			replacePath := regexp.ReplaceAll([]byte(c.Request.URL.Path), []byte(items[1]))
			c.Request.URL.Path = string(replacePath)
		}

		//传递到下一中间件
		c.Next()
	}
}
