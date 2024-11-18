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

// 限流器中间件
func HTTPFlowLimitMiddleware() gin.HandlerFunc {
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

		//限流项 1.服务端 2.客户端
		if serviceDetail.AccessControl.ServiceFlowLimit > 0 {
			serviceFlowLimiter, err := circuit_rate.FlowLimiterHandler.GetFlowLimiter(fmt.Sprintf("%s_%s", public.FlowService, serviceDetail.Info.ServiceName), serviceDetail.AccessControl.ServiceFlowLimit)
			if err != nil {
				middleware.ResponseError(c, 9002, err)
				//中断中间件传递链
				c.Abort()
				return
			}
			if !serviceFlowLimiter.Allow() {
				middleware.ResponseError(c, 9003, errors.New(fmt.Sprintf("service flow limit exceeded: %v", serviceDetail.AccessControl.ServiceFlowLimit)))
				//中断中间件传递链
				c.Abort()
				return
			}
		}

		if serviceDetail.AccessControl.ClientIPFlowLimit > 0 {
			clientFlowLimiter, err := circuit_rate.FlowLimiterHandler.GetFlowLimiter(fmt.Sprintf("%s_%s_client", public.FlowService, serviceDetail.Info.ServiceName), serviceDetail.AccessControl.ClientIPFlowLimit)
			if err != nil {
				middleware.ResponseError(c, 9004, err)
				//中断中间件传递链
				c.Abort()
				return
			}
			if !clientFlowLimiter.Allow() {
				middleware.ResponseError(c, 9005, errors.New(fmt.Sprintf("%v client flow limit exceeded: %v", c.ClientIP(), serviceDetail.AccessControl.ClientIPFlowLimit)))
				//中断中间件传递链
				c.Abort()
				return
			}
		}

		//传递到下一中间件
		c.Next()
	}
}
