package tcp_proxy_middleware

import (
	"fmt"
	"github.com/starMoonZhao/go_gateway/circuit_rate"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/public"
	"github.com/starMoonZhao/go_gateway/tcp_proxy_router"
)

// 流量统计中间件
func TCPFlowCountMiddleware() func(t *tcp_proxy_router.TCPRouterSliceContext) {
	return func(t *tcp_proxy_router.TCPRouterSliceContext) {
		//获取上游服务信息
		serviceInterface := t.Get("service")
		if serviceInterface == nil {
			t.Conn.Write([]byte("service not found"))
			//中断中间件传递链
			t.Abort()
			return
		}
		//类型转换
		serviceDetail := serviceInterface.(*dao.ServiceDetail)

		//统计项 1.全站 2.服务 3.租户
		totalFlowCount, err := circuit_rate.FlowCounterHandler.GetFlowCounter(public.FlowTotal)
		if err != nil {
			t.Conn.Write([]byte(err.Error()))
			//中断中间件传递链
			t.Abort()
			return
		}
		totalFlowCount.Increase()

		serviceFlowCount, err := circuit_rate.FlowCounterHandler.GetFlowCounter(fmt.Sprintf("%s_%s", public.FlowService, serviceDetail.Info.ServiceName))
		if err != nil {
			t.Conn.Write([]byte(err.Error()))
			//中断中间件传递链
			t.Abort()
			return
		}
		serviceFlowCount.Increase()

		//传递到下一中间件
		t.Next()
	}
}
