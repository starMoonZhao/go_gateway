package grpc_proxy_middleware

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/circuit_rate"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/public"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"log"
	"strings"
)

// 限流器中间件
func GrpcFlowLimitMiddleware(service *dao.ServiceDetail) func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		//限流项 1.服务端 2.客户端
		if service.AccessControl.ServiceFlowLimit > 0 {
			serviceFlowLimiter, err := circuit_rate.FlowLimiterHandler.GetFlowLimiter(fmt.Sprintf("%s_%s", public.FlowService, service.Info.ServiceName), service.AccessControl.ServiceFlowLimit)
			if err != nil {
				return err
			}
			if !serviceFlowLimiter.Allow() {
				return errors.New(fmt.Sprintf("service flow limit: %v\n", service.AccessControl.ServiceFlowLimit))
			}
		}

		//解析请求来源ip
		peerCtx, ok := peer.FromContext(stream.Context())
		if !ok {
			return errors.New("peer context not found")
		}
		peerAddr := peerCtx.Addr.String()
		lastIndex := strings.LastIndex(peerAddr, ":")
		clientIp := peerAddr[0:lastIndex]

		if service.AccessControl.ClientIPFlowLimit > 0 {
			clientFlowLimiter, err := circuit_rate.FlowLimiterHandler.GetFlowLimiter(fmt.Sprintf("%s_%s_client", public.FlowService, service.Info.ServiceName), service.AccessControl.ClientIPFlowLimit)
			if err != nil {
				return err
			}
			if !clientFlowLimiter.Allow() {
				return errors.New(fmt.Sprintf("client %v flow limit: %v\n", clientIp, service.AccessControl.ClientIPFlowLimit))
			}
		}

		if err := handler(srv, stream); err != nil {
			log.Printf(fmt.Sprintf("grpc flow limit handler error: %v\n", err))
			return err
		}
		return nil
	}
}
