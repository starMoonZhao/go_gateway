package tcp_proxy_middleware

import (
	"fmt"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/public"
	"github.com/starMoonZhao/go_gateway/tcp_proxy_router"
	"strings"
)

// 请求路径重写规则
func TCPBlackListMiddleware() func(t *tcp_proxy_router.TCPRouterSliceContext) {
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

		//根据serviceDetail获取ip白名单
		whiteIpList := []string{}
		if serviceDetail.AccessControl.WhiteList != "" {
			whiteIpList = strings.Split(serviceDetail.AccessControl.WhiteList, ",")
		}

		//根据serviceDetail获取ip黑名单
		blackIpList := []string{}
		if serviceDetail.AccessControl.BlackList != "" {
			blackIpList = strings.Split(serviceDetail.AccessControl.BlackList, ",")
		}

		//获取clinetIp
		split := strings.Split(t.Conn.RemoteAddr().String(), ":")
		clientIP := split[0]

		//如果已有白名单 白名单规则优先于黑名单 直接通过即可
		if serviceDetail.AccessControl.OpenAuth == 1 && len(whiteIpList) == 0 && len(blackIpList) > 0 {
			//如果ip在黑名单中 直接返回请求
			if public.InStringSlice(blackIpList, clientIP) {
				t.Conn.Write([]byte(fmt.Sprintf("%s in black ip list.", clientIP)))
				//中断中间件传递链
				t.Abort()
				return
			}
		}

		//传递到下一中间件
		t.Next()
	}
}
