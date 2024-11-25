package tcp_server

import (
	"context"
	"fmt"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/tcp_proxy_middleware"
	"github.com/starMoonZhao/go_gateway/tcp_proxy_router"
	"log"
	"net/http"
)

// 所有服务的TCPServer列表
var tcpServerList = []*tcp_proxy_router.TCPServer{}

// 启动TCP服务器
// 与启动http不同的是 这个server实际上会启动很多个：服务和TCPServer是一对一的关系，这是因为TCP是使用port接入的方式，因为需要监听锁设置的端口
// 而http代理使用域名接入或者路径接入的方式，器监听的端口是固定的 不要需要启动多个HTTPServer
func TCPServerRun() {
	//step1: 查询所有的tcp服务列表
	tcpServiceList := dao.ServiceManegerHandler.GetTCPServiceList()
	for _, service := range tcpServiceList {
		//step2: 为每个service创建TCPServer
		tempService := service
		go func(service *dao.ServiceDetail) {
			//step3: 获取带服务对应的负载均衡器
			addr := fmt.Sprintf(":%d", service.TCPRule.Port)

			//step4: 构建路由及设置中间件
			tcpSliceGroup := tcp_proxy_router.NewTCPSliceGroup().Use(
				tcp_proxy_middleware.TCPFlowCountMiddleware(),
				tcp_proxy_middleware.TCPFlowLimitMiddleware(),
				tcp_proxy_middleware.TCPWhiteListMiddleware(),
				tcp_proxy_middleware.TCPBlackListMiddleware(),
				tcp_proxy_middleware.TCPReverseProxyMiddleware(),
			)
			tcpSliceRouterHandler := tcp_proxy_router.NewTCPSliceRouterHandler(tcpSliceGroup)

			baseCtx := context.WithValue(context.Background(), "service", service)

			//step5: 构建TCPServer
			tcpServer := &tcp_proxy_router.TCPServer{
				Handler: tcpSliceRouterHandler,
				BaseCtx: baseCtx,
				Addr:    addr,
			}

			//step6: 将TCPServer存入列表中
			tcpServerList = append(tcpServerList, tcpServer)
			log.Printf("tcpServer addr:%s\n", addr)

			//step7: 启动TCPServer
			if err := tcpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("start tcp server %v err:%v\n", addr, err)
			}
		}(tempService)
	}
}

// 停止TCP服务器
func TCPServerStop() {
	for _, tcpServer := range tcpServerList {
		tcpServer.Close()
		log.Printf(" [INFO] TCPServerStop %v stopped\n", tcpServer.Addr)
	}
}
