package grpc_proxy_middleware

import (
	"fmt"
	"github.com/starMoonZhao/go_gateway/circuit_rate"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/public"
	"google.golang.org/grpc"
	"log"
)

// 流量统计中间件
func GrpcFlowCountMiddleware(service *dao.ServiceDetail) func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		//统计项 1.全站 2.服务 3.租户
		totalFlowCount, err := circuit_rate.FlowCounterHandler.GetFlowCounter(public.FlowTotal)
		if err != nil {
			return err
		}
		totalFlowCount.Increase()

		serviceFlowCount, err := circuit_rate.FlowCounterHandler.GetFlowCounter(fmt.Sprintf("%s_%s", public.FlowService, service.Info.ServiceName))
		if err != nil {
			return err
		}
		serviceFlowCount.Increase()

		if err = handler(srv, stream); err != nil {
			log.Printf("grpc_proxy_middleware: error handling grpc flow count: %v", err)
			return err
		}
		return nil
	}
}
