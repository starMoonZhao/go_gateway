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

// 流量统计中间件
func HTTPFlowCountMiddleware() gin.HandlerFunc {
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

		//统计项 1.全站 2.服务 3.租户
		totalFlowCount, err := circuit_rate.FlowCounterHandler.GetFlowCounter(public.FlowTotal)
		if err != nil {
			middleware.ResponseError(c, 9002, err)
			//中断中间件传递链
			c.Abort()
			return
		}
		totalFlowCount.Increase()

		serviceFlowCount, err := circuit_rate.FlowCounterHandler.GetFlowCounter(fmt.Sprintf("%s_%s", public.FlowService, serviceDetail.Info.ServiceName))
		if err != nil {
			middleware.ResponseError(c, 9003, err)
			//中断中间件传递链
			c.Abort()
			return
		}
		serviceFlowCount.Increase()

		//传递到下一中间件
		c.Next()
	}
}
