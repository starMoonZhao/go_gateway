package reverse_proxy

import (
	"context"
	"github.com/e421083458/grpc-proxy/proxy"
	"github.com/starMoonZhao/go_gateway/reverse_proxy/load_balance"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
)

func NewGrpcLoadBalanceHandler(lb load_balance.LoadBalance) grpc.StreamHandler {
	//闭包
	return func() grpc.StreamHandler {
		//请求协调者
		director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
			nextAddr, err := lb.Get("")
			if err != nil {
				log.Fatalf("get next addr err:", err)
			}
			//拨号
			conn, err := grpc.DialContext(ctx, nextAddr, grpc.WithCodec(proxy.Codec()), grpc.WithInsecure())
			//加载输入内容
			md, _ := metadata.FromIncomingContext(ctx)
			//加载输出上下文
			outCtx, _ := context.WithCancel(ctx)
			outCtx = metadata.NewOutgoingContext(ctx, md.Copy())
			return outCtx, conn, err
		}
		return proxy.TransparentHandler(director)
	}()
}
