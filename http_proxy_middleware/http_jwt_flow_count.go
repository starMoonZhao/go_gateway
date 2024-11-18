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

// jwt流量统计中间件
func HTTPJwtFlowCountMiddleware() gin.HandlerFunc {
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

		//统计项 租户
		appFlowCount, err := circuit_rate.FlowCounterHandler.GetFlowCounter(fmt.Sprintf("%s_%s", public.FlowApp, appDetail.APPID))
		if err != nil {
			middleware.ResponseError(c, 9002, err)
			//中断中间件传递链
			c.Abort()
			return
		}
		appFlowCount.Increase()
		if appDetail.Qpd > 0 && appDetail.Qpd > appFlowCount.TotalCount {
			middleware.ResponseError(c, 9003, errors.New(fmt.Sprintf("APP QPD limit:%v current:%v", appDetail.Qpd, appFlowCount.TotalCount)))
			//中断中间件传递链
			c.Abort()
			return
		}

		//传递到下一中间件
		c.Next()
	}
}
