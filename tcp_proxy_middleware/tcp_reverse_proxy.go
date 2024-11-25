package tcp_proxy_middleware

import (
	"fmt"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/reverse_proxy"
	"github.com/starMoonZhao/go_gateway/tcp_proxy_router"
)

// 反向代理匹配
func TCPReverseProxyMiddleware() func(t *tcp_proxy_router.TCPRouterSliceContext) {
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

		//根据serviceDetail创建负载均衡器
		loadBalance, err := dao.LoadBalancerHandler.GetLoadBalance(serviceDetail)
		if err != nil {
			t.Conn.Write([]byte(fmt.Sprintf("create LoadBalance fail: %v", err)))
			//中断中间件传递链
			t.Abort()
			return
		}

		//创建reverseproxy
		proxy := reverse_proxy.NewTCPLoadBalanceReverseProxy(t, loadBalance)
		//使用reverseproxy.ServerHTTP(c.Request,c.Response)
		proxy.ServeTCP(t.Ctx, t.Conn)

		t.Abort()
		return
	}
}
