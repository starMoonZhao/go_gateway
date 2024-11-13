package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/middleware"
	"strings"
)

// header头转换
func HTTPHeaderTransferMiddleware() gin.HandlerFunc {
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

		//根据serviceDetail获取header转换规则
		for _, item := range strings.Split(serviceDetail.HTTPRule.HeaderTransfer, ",") {
			//规则格式：操作（add edit del） headername headervalue
			items := strings.Split(item, " ")
			if len(items) != 3 {
				continue
			}
			if items[0] == "add" || items[0] == "edit" {
				c.Request.Header.Add(items[1], items[2])
			} else if items[0] == "del" {
				c.Request.Header.Del(items[1])
			}
		}
		//传递到下一中间件
		c.Next()
	}
}
