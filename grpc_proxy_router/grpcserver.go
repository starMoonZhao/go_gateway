package grpc_proxy_router

import (
	"fmt"
	"github.com/e421083458/grpc-proxy/proxy"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/grpc_proxy_middleware"
	"github.com/starMoonZhao/go_gateway/reverse_proxy"
	"google.golang.org/grpc"
	"log"
	"net"
)

// grpc服务列表
var grpcServerList = []*WarpGrpcServer{}

// 映射代理地址与服务器的关系
type WarpGrpcServer struct {
	Addr string
	*grpc.Server
}

func GrpcServerRun() {
	//遍历启动grpc服务器
	grpcServiceList := dao.ServiceManegerHandler.GetGRPCServiceList()
	for _, grpcService := range grpcServiceList {
		tempItem := grpcService
		go func(serviceDetail *dao.ServiceDetail) {
			//获取监听地址
			addr := fmt.Sprintf(":%d", serviceDetail.GRPCRule.Port)
			//获取负载均衡器
			loadBalance, err := dao.LoadBalancerHandler.GetLoadBalance(serviceDetail)
			if err != nil {
				log.Fatalf("grpc server load balance error:%v\n", err)
				return
			}
			//监听代理地址
			listen, err := net.Listen("tcp", addr)
			if err != nil {
				log.Fatalf("grpc server listen port error:%v\n", err)
				return
			}
			//获取负载均衡handler
			streamHandler := reverse_proxy.NewGrpcLoadBalanceHandler(loadBalance)
			//创建grpc服务器
			server := grpc.NewServer(grpc.ChainStreamInterceptor(
				grpc_proxy_middleware.GrpcFlowCountMiddleware(serviceDetail),
				grpc_proxy_middleware.GrpcFlowLimitMiddleware(serviceDetail),
				grpc_proxy_middleware.GrpcJwtAuthTokenMiddleware(serviceDetail),
				grpc_proxy_middleware.GrpcJwtFlowCountMiddleware(serviceDetail),
				grpc_proxy_middleware.GrpcJwtFlowLimitMiddleware(serviceDetail),
				grpc_proxy_middleware.GrpcWhiteListMiddleware(serviceDetail),
				grpc_proxy_middleware.GrpcBlackListMiddleware(serviceDetail),
				grpc_proxy_middleware.GrpcHeaderTransferMiddleware(serviceDetail),
			),
				grpc.CustomCodec(proxy.Codec()),
				grpc.UnknownServiceHandler(streamHandler))

			grpcServerList = append(grpcServerList, &WarpGrpcServer{addr, server})
			log.Printf("grpc server run:%d\n", addr)

			//启动grpc服务器
			if err := server.Serve(listen); err != nil {
				log.Fatalf("grpc server start error:%v\n", err)
			}
		}(tempItem)
	}
}

func GrpcServerStop() {
	for _, grpcServer := range grpcServerList {
		grpcServer.GracefulStop()
		log.Printf("grpc server stop:%d\n", grpcServer.Addr)
	}
}
