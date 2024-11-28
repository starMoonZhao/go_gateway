package tcp_proxy_middleware

import (
	"fmt"
	"github.com/starMoonZhao/go_gateway/circuit_rate"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/public"
	"github.com/starMoonZhao/go_gateway/tcp_proxy_router"
	"strings"
)

// 限流器中间件
func TCPFlowLimitMiddleware() func(t *tcp_proxy_router.TCPRouterSliceContext) {
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

		//限流项 1.服务端 2.客户端
		if serviceDetail.AccessControl.ServiceFlowLimit > 0 {
			serviceFlowLimiter, err := circuit_rate.FlowLimiterHandler.GetFlowLimiter(fmt.Sprintf("%s_%s", public.FlowService, serviceDetail.Info.ServiceName), serviceDetail.AccessControl.ServiceFlowLimit)
			if err != nil {
				t.Conn.Write([]byte(err.Error()))
				//中断中间件传递链
				t.Abort()
				return
			}
			if !serviceFlowLimiter.Allow() {
				t.Conn.Write([]byte(fmt.Sprintf("service flow limit exceeded: %v", serviceDetail.AccessControl.ServiceFlowLimit)))
				//中断中间件传递链
				t.Abort()
				return
			}
		}

		//获取clinetIp
		split := strings.Split(t.Conn.RemoteAddr().String(), ":")
		clientIP := split[0]

		if serviceDetail.AccessControl.ClientIPFlowLimit > 0 {
			clientFlowLimiter, err := circuit_rate.FlowLimiterHandler.GetFlowLimiter(fmt.Sprintf("%s_%s_%s", public.FlowService, serviceDetail.Info.ServiceName, clientIP), serviceDetail.AccessControl.ClientIPFlowLimit)
			if err != nil {
				t.Conn.Write([]byte(err.Error()))
				//中断中间件传递链
				t.Abort()
				return
			}
			if !clientFlowLimiter.Allow() {
				t.Conn.Write([]byte(fmt.Sprintf("%v client flow limit exceeded: %v", clientIP, serviceDetail.AccessControl.ClientIPFlowLimit)))
				//中断中间件传递链
				t.Abort()
				return
			}
		}

		//传递到下一中间件
		t.Next()
	}
}
