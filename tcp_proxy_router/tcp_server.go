package tcp_proxy_router

import (
	"context"
	"github.com/pkg/errors"
	"log"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// 全局变量定义
var (
	ErrServerClosed     = errors.New("tcp:server closed")
	ErrAbortHandler     = errors.New("tcp:abort handler")
	ServerContextKey    = &contextKey{"tcp-server"}
	LocalAddrContextKey = &contextKey{"local-addr"}
)

type contextKey struct {
	name string
}

type TCPServer struct {
	Addr    string     //服务地址
	Handler TCPHandler //服务处理器
	err     error
	BaseCtx context.Context //服务上下文
	Ctx     context.Context //服务上下文

	//连接参数来源于TCP服务配置
	WriteTimeout     time.Duration
	ReadTimeout      time.Duration
	KeepAliveTimeout time.Duration

	mutex      sync.Mutex
	inShutDown int32         //服务器关闭标志
	doneChan   chan struct{} //服务器关闭通道

	listener *OnecCloseListener //服务监听器
}

// 服务监听器包装对象
type OnecCloseListener struct {
	net.Listener //原生TCP服务监听器
	once         sync.Once
	err          error
}

// 关闭监听器
func (l *OnecCloseListener) Close() error {
	l.once.Do(func() {
		l.err = l.Listener.Close()
	})
	return l.err
}

// 检测服务器是否处于关闭状态
func (srv *TCPServer) IsShuttingDown() bool {
	return atomic.LoadInt32(&srv.inShutDown) == 1
}

// 创建TCPServer并启动
func (srv *TCPServer) ListenAndServe() error {
	//检查服务是否处于关闭状态
	if srv.IsShuttingDown() {
		return ErrServerClosed
	}
	//初始化服务关闭通道
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}

	//根据服务地址初始化服务器
	addr := srv.Addr
	if addr == "" {
		return errors.New("tcp:server addr is empty")
	}
	//启动TCP服务监听
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	//使用获取到的监听器去处理接收到的请求
	return srv.Serve(listen)
}

// 使用listener监听请求 并处理请求
func (srv *TCPServer) Serve(ln net.Listener) error {
	//包装原生监听器
	srv.listener = &OnecCloseListener{Listener: ln}
	defer srv.listener.Close()

	if srv.BaseCtx == nil {
		srv.BaseCtx = context.Background()
	}

	baseCtx := srv.BaseCtx
	srv.Ctx = context.WithValue(baseCtx, ServerContextKey, srv)

	//死循环监听请求
	for {
		//接收请求 并获取通信连接
		conn, err := ln.Accept()
		if err != nil {
			//请求失败：这个时候要判断是因为TCPServer服务器关闭引起的还是其他原因
			//如果是服务器关闭引起的错误 那么不再接收请求 返回异常并退出
			select {
			case <-srv.doneChan:
				return ErrServerClosed
			default:
				//没有关闭服务器 继续接收请求
			}
			log.Printf("Listener accept error: %v\n", err)
			continue
		}

		//设置连接参数
		srv.SetConnParam(conn)

		//连接成功后 继续使用获取到的conn异步执行后续逻辑
		go srv.ConnServe(conn)
	}
}

// 设置请求连接的参数
func (srv *TCPServer) SetConnParam(conn net.Conn) {
	if param := srv.WriteTimeout; param != 0 {
		conn.SetWriteDeadline(time.Now().Add(param))
	}

	if param := srv.ReadTimeout; param != 0 {
		conn.SetReadDeadline(time.Now().Add(param))
	}

	if param := srv.KeepAliveTimeout; param != 0 {
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(time.Second * 100)
		}
	}
}

// 异步出去执行请求后的逻辑
func (srv *TCPServer) ConnServe(conn net.Conn) {
	//异常恢复与连接关闭
	defer func() {
		if err := recover(); err != nil && err != ErrAbortHandler {
			//输出连接中断异常信息
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("tcp: %v server panic: %v\n%s", conn.RemoteAddr(), err, buf)
		}
		conn.Close()
	}()
	//获取TCPServer中的handler
	if srv.Handler == nil {
		panic("handler is nil")
	}
	//执行TCPHandler中的处理逻辑
	srv.Handler.ServeTCP(srv.Ctx, conn)
}

// TCPServer关闭逻辑：设置inShutdown管道、inShutDown标志、listener监听器
func (srv *TCPServer) Close() error {
	atomic.StoreInt32(&srv.inShutDown, 1)
	close(srv.doneChan)
	srv.listener.Close()
	return nil
}
