package http_proxy_middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/middleware"
	"github.com/starMoonZhao/go_gateway/public"
)

// 基于请求信息 匹配接入方式
func HTTPAccessModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceDetail, err := dao.ServiceManegerHandler.HTTPAccessMode(c)
		if err != nil {
			middleware.ResponseError(c, 9001, err)
			//中断中间件传递链
			c.Abort()
			return
		}
		fmt.Println("serviceInfo:{}", public.Obj2Json(serviceDetail))
		c.Set("service", serviceDetail)
		//此中间件执行结束后传递到下一中间件
		c.Next()
	}
}
