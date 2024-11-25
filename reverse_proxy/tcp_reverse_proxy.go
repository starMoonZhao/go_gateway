package reverse_proxy

import (
	"context"
	"github.com/starMoonZhao/go_gateway/reverse_proxy/load_balance"
	"github.com/starMoonZhao/go_gateway/tcp_proxy_router"
	"io"
	"log"
	"net"
	"time"
)

// TCP反向代理
type TCPReverseProxy struct {
	ctx             context.Context                                                   //单次请求单次设置
	SrcConn         net.Conn                                                          //源请求连接
	Addr            string                                                            //代理服务地址（真实服务地址）
	KeepAlivePeriod time.Duration                                                     //连接的检测时长
	DialTimeout     time.Duration                                                     //超时时长
	DialContext     func(ctx context.Context, network, addr string) (net.Conn, error) //拨号函数,向下游服务发起通信获取TCP连接 以过去服务
	OnDialError     func(src net.Conn, dstDialErr error)
}

// 返回TCPReverseProxy
func NewTCPLoadBalanceReverseProxy(c *tcp_proxy_router.TCPRouterSliceContext, lb load_balance.LoadBalance) *TCPReverseProxy {
	//todo:这里为什么要使用闭包的形式返回？
	//获取由负载均衡器产生的下游地址
	nextAddr, err := lb.Get("")
	if err != nil {
		log.Fatalf("get next addr err: %v\n", err)
	}
	return &TCPReverseProxy{
		Addr:            nextAddr,
		SrcConn:         c.Conn,
		ctx:             c.Ctx,
		KeepAlivePeriod: time.Second,
		DialTimeout:     time.Second,
	}
}

// 获取拨号上下文
func (p *TCPReverseProxy) dialContext() func(ctx context.Context, network, addr string) (net.Conn, error) {
	if p.DialContext != nil {
		return p.DialContext
	}
	return (&net.Dialer{
		Timeout:   p.DialTimeout,
		KeepAlive: p.KeepAlivePeriod,
	}).DialContext
}

// 传入上游conn 在这里完成下游连接及上下游数据的交换
func (p *TCPReverseProxy) ServeTCP(ctx context.Context, src net.Conn) {
	//设置连接超时
	var cancel context.CancelFunc
	//拨号获取下游连接
	dst, err := p.dialContext()(ctx, "tcp", p.Addr)
	if err != nil {
		cancel()
	}
	if err != nil {
		p.onDialErr()(src, err)
		return
	}

	//退出下游连接
	defer func() { go dst.Close() }()

	//在数据请求前设置下游连接的keepalive参数
	if p.KeepAlivePeriod > 0 {
		if conn, ok := src.(*net.TCPConn); ok {
			conn.SetKeepAlive(true)
			conn.SetKeepAlivePeriod(p.KeepAlivePeriod)
		}
	}

	//开始数据传递
	errc := make(chan error, 1)
	//上游拷贝到下游
	go p.copy(errc, src, dst)
	//下游拷贝到上游
	go p.copy(errc, dst, src)
	<-errc
}

// 拨号异常处理
func (p *TCPReverseProxy) onDialErr() func(src net.Conn, dstDialErr error) {
	if p.OnDialError != nil {
		return p.OnDialError
	}
	return func(src net.Conn, dstDialErr error) {
		log.Printf("tcpproxy err:for incoming conn %v, error dialing %q: %v", src.RemoteAddr(), p.Addr, dstDialErr)
		src.Close()
	}
}

// 数据拷贝
func (p *TCPReverseProxy) copy(errc chan<- error, dst, src net.Conn) {
	_, err := io.Copy(dst, src)
	if err != nil {
		errc <- err
	}
}
