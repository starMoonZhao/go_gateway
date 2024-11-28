package grpc_proxy_middleware

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/public"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"strings"
)

// 权限验证中间件
func GrpcJwtAuthTokenMiddleware(service *dao.ServiceDetail) func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		//step1 获取请求头中的授权信息
		//step2 解密授权信息
		//step3 将租户信息存入gin.context
		//获取元数据
		md, ok := metadata.FromIncomingContext(stream.Context())
		if !ok {
			return errors.New("failed to get metadata from incoming context")
		}
		auths := md.Get("authorization")
		authToken := ""
		if len(auths) > 0 {
			authToken = auths[0]
		}
		token := strings.ReplaceAll(authToken, "Bearer ", "")

		appMatched := false
		if token != "" {
			claims, err := public.JwtDecode(token)
			if err != nil {
				return err
			}
			appList := dao.AppManegerHandler.GetAppList()
			for _, app := range appList {
				if app.APPID == claims.Issuer {
					md.Set("app", public.Obj2Json(app))
					appMatched = true
					break
				}
			}
		}

		if service.AccessControl.OpenAuth == 1 && !appMatched {
			return errors.New("access denied")
		}

		if err := handler(srv, stream); err != nil {
			log.Printf(fmt.Sprintf("grpc_jwt_auth_token handler err:%v\n", err))
			return err
		}
		return nil
	}
}
