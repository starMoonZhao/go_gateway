package http_proxy_middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/middleware"
	"github.com/starMoonZhao/go_gateway/reverse_proxy"
)

// 反向代理匹配
func HTTPReverseProxyMiddleware() gin.HandlerFunc {
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

		//根据serviceDetail创建负载均衡器
		loadBalance, err := dao.LoadBalancerHandler.GetLoadBalance(serviceDetail)
		if err != nil {
			middleware.ResponseError(c, 9002, err)
			//中断中间件传递链
			c.Abort()
			return
		}

		//根据serviceDetail创建连接池
		trans, err := dao.TransportorHandler.GetTrans(serviceDetail)
		if err != nil {
			middleware.ResponseError(c, 9003, err)
			//中断中间件传递链
			c.Abort()
			return
		}
		//创建reverseproxy
		proxy := reverse_proxy.NewLoadBalanceReverseProxy(c, loadBalance, trans)
		//使用reverseproxy.ServerHTTP(c.Request,c.Response)
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
		return
	}
}
