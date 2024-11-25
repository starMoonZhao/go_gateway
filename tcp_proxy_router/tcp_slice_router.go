package tcp_proxy_router

import (
	"context"
	"math"
	"net"
)

const abortIndex int8 = math.MaxInt8

// group结构体：记录TCP服务中的一个router下的中间件（handler配置）
type TCPSliceGroup struct {
	Handlers []TCPHanlderFunc
}

// 创建group
func NewTCPSliceGroup() *TCPSliceGroup {
	return &TCPSliceGroup{}
}

// 设置TCPSliceGroup的中间件列表
func (t *TCPSliceGroup) Use(middlewares ...TCPHanlderFunc) *TCPSliceGroup {
	t.Handlers = append(t.Handlers, middlewares...)
	return t
}

// router请求上下文:链式调用中间件时需要使用到的上下文数据
// 这是与每个请求相关联的
type TCPRouterSliceContext struct {
	Conn  net.Conn        //源请求连接
	Ctx   context.Context //上下文
	index int8            //中间件调用链的下标
	*TCPSliceGroup
}

func (t *TCPRouterSliceContext) Get(key interface{}) interface{} {
	return t.Ctx.Value(key)
}

func (t *TCPRouterSliceContext) Set(key, val interface{}) {
	t.Ctx = context.WithValue(t.Ctx, key, val)
}

// 核心方法：中间件回调入口
func (t *TCPRouterSliceContext) Next() {
	t.index++
	if t.index < int8(len(t.Handlers)) {
		t.Handlers[t.index](t)
	}
}

// 跳出中间件回调
func (t *TCPRouterSliceContext) Abort() {
	t.index = abortIndex
}

// 重置回调
func (t *TCPRouterSliceContext) Reset() {
	t.index = -1
}

// TCP调用链函数（本质上的中间件）
type TCPHanlderFunc func(*TCPRouterSliceContext)

// router中中间件执行的入口 根handler 在这里开始使用TCPRouterSliceContext进行中间件的链式调用
type TCPSliceRouterHandler struct {
	group *TCPSliceGroup
}

func NewTCPSliceRouterHandler(group *TCPSliceGroup) *TCPSliceRouterHandler {
	return &TCPSliceRouterHandler{
		group: group,
	}
}

func (t *TCPSliceRouterHandler) ServeTCP(ctx context.Context, conn net.Conn) {
	//创建调用上下文
	tcpSliceGroup := &TCPSliceGroup{}
	*tcpSliceGroup = *t.group
	tcpRouterSliceContext := &TCPRouterSliceContext{
		Ctx:           ctx,
		Conn:          conn,
		TCPSliceGroup: tcpSliceGroup,
	}
	tcpRouterSliceContext.Reset()
	tcpRouterSliceContext.Next()
}
