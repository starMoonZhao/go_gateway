package http_proxy_middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/circuit_rate"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/middleware"
	"github.com/starMoonZhao/go_gateway/public"
)

// jwt限流器中间件
func HTTPJwtFlowLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//获取上游z租户信息
		appInfoInterface, ok := c.Get("app")
		if !ok {
			middleware.ResponseError(c, 9001, errors.New("app not found"))
			//中断中间件传递链
			c.Abort()
			return
		}
		//类型转换
		appDetail := appInfoInterface.(*dao.APP)

		//限流项 租户
		if appDetail.Qps > 0 {
			appFlowLimiter, err := circuit_rate.FlowLimiterHandler.GetFlowLimiter(fmt.Sprintf("%s_%s", public.FlowApp, appDetail.APPID), int(appDetail.Qps))
			if err != nil {
				middleware.ResponseError(c, 9002, err)
				//中断中间件传递链
				c.Abort()
				return
			}
			if !appFlowLimiter.Allow() {
				middleware.ResponseError(c, 9003, errors.New(fmt.Sprintf("app flow limit exceeded: %v", appDetail.Qps)))
				//中断中间件传递链
				c.Abort()
				return
			}
		}

		//传递到下一中间件
		c.Next()
	}
}
