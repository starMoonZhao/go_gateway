package grpc_proxy_middleware

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/public"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"log"
	"strings"
)

// 请求路径重写规则
func GrpcBlackListMiddleware(service *dao.ServiceDetail) func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		//根据serviceDetail获取ip白名单
		whiteIpList := []string{}
		if service.AccessControl.WhiteList != "" {
			whiteIpList = strings.Split(service.AccessControl.WhiteList, ",")
		}

		//根据serviceDetail获取ip黑名单
		blackIpList := []string{}
		if service.AccessControl.BlackList != "" {
			blackIpList = strings.Split(service.AccessControl.BlackList, ",")
		}

		//解析请求来源ip
		peerCtx, ok := peer.FromContext(stream.Context())
		if !ok {
			return errors.New("peer context not found")
		}
		peerAddr := peerCtx.Addr.String()
		lastIndex := strings.LastIndex(peerAddr, ":")
		clientIp := peerAddr[0:lastIndex]

		//如果已有白名单 白名单规则优先于黑名单 直接通过即可
		if service.AccessControl.OpenAuth == 1 && len(whiteIpList) == 0 && len(blackIpList) > 0 {
			//如果ip在黑名单中 直接返回请求
			if public.InStringSlice(blackIpList, clientIp) {
				return errors.New(fmt.Sprintf("%s in black ip list.", clientIp))
			}
		}

		if err := handler(srv, stream); err != nil {
			log.Printf(fmt.Sprintf("grpc black list error: %v", err))
			return err
		}
		return nil
	}
}
