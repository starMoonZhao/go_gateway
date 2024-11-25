package tcp_proxy_router

import (
	"context"
	"net"
)

// TCP处理器
type TCPHandler interface {
	ServeTCP(ctx context.Context, conn net.Conn)
}
