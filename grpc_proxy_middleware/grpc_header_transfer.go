package grpc_proxy_middleware

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/dao"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"strings"
)

// header头转换
func GrpcHeaderTransferMiddleware(service *dao.ServiceDetail) func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		//获取元数据
		md, ok := metadata.FromIncomingContext(stream.Context())
		if !ok {
			return errors.New("failed to get metadata from incoming context")
		}
		//根据serviceDetail获取header转换规则
		for _, item := range strings.Split(service.GRPCRule.HeaderTransfer, ",") {
			//规则格式：操作（add edit del） headername headervalue
			items := strings.Split(item, " ")
			if len(items) != 3 {
				continue
			}
			if items[0] == "add" || items[0] == "edit" {
				md.Set(items[1], items[2])
			} else if items[0] == "del" {
				delete(md, items[1])
			}
		}

		if err := handler(srv, stream); err != nil {
			log.Printf(fmt.Sprintf("error handling grpc header transfer request: %v\n", err))
			return err
		}
		return nil
	}
}
