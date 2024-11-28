package grpc_proxy_middleware

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/circuit_rate"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/public"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
)

// jwt限流器中间件
func GrpcJwtFlowLimitMiddleware(service *dao.ServiceDetail) func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		//获取上游租户信息
		md, ok := metadata.FromIncomingContext(stream.Context())
		if !ok {
			return errors.New("failed to get metadata from incoming context")
		}
		appInfos := md.Get("app")
		if len(appInfos) == 0 {
			if err := handler(srv, stream); err != nil {
				log.Printf(fmt.Sprintf("grpc_jwt_flow_count middleware error: %s\n", err.Error()))
				return err
			}
		}

		appInfo := &dao.APP{}
		if err := json.Unmarshal([]byte(appInfos[0]), appInfo); err != nil {
			return err
		}

		//限流项 租户
		if appInfo.Qps > 0 {
			appFlowLimiter, err := circuit_rate.FlowLimiterHandler.GetFlowLimiter(fmt.Sprintf("%s_%s", public.FlowApp, appInfo.APPID), int(appInfo.Qps))
			if err != nil {
				return err
			}
			if !appFlowLimiter.Allow() {
				return errors.New(fmt.Sprintf("app flow limit exceeded: %v", appInfo.Qps))
			}
		}

		if err := handler(srv, stream); err != nil {
			log.Printf(fmt.Sprintf("grpc_jwt_flow_limit middleware error: %v\n", err))
			return err
		}
		return nil
	}
}
