package reverse_proxy

import (
	"bytes"
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"github.com/starMoonZhao/go_gateway/middleware"
	"github.com/starMoonZhao/go_gateway/reverse_proxy/load_balance"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

// 根据http上下文、负载均衡器LoadBalance和连接池构建反向代理
func NewLoadBalanceReverseProxy(c *gin.Context, lb load_balance.LoadBalance, transport *http.Transport) *httputil.ReverseProxy {
	//构建请求协调者：将请求进行参数配置、服务节点选择、请求转发
	director := func(req *http.Request) {
		//根据负载均衡器获取下一可用服务地址
		nextAddr, err := lb.Get(req.URL.String())
		if err != nil || nextAddr == "" {
			panic("get next addr error")
		}
		//解析可用服务地址
		target, err := url.Parse(nextAddr)
		if err != nil {
			panic(err)
		}
		//获取URL 中查询部分（即 ? 后面的部分）
		targetQuery := target.RawQuery
		//参数填充
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		//解析请求路径
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		req.Host = target.Host

		//解析请求参数
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		//User-Agent 是一个 HTTP 请求头（HTTP header），它包含了客户端应用程序的信息，通常用于标识发起请求的浏览器、操作系统或设备类型等
		//Web 服务器和应用程序可以通过 User-Agent 来检测客户端的特性或行为，例如，浏览器类型、设备类型（手机或电脑）、操作系统等
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "user-agent")
		}
	}
	//构建返回体输出：将代理的请求收到的内容写入的原请求中
	modifier := func(res *http.Response) error {
		//兼容websocket
		if strings.Contains(res.Header.Get("Connection"), "Upgrade") {
			return nil
		}

		var payLoad []byte
		var readErr error

		//兼容gzip
		if strings.Contains(res.Header.Get("Content-Encoding"), "gzip") {
			//读取内容
			reader, err := gzip.NewReader(res.Body)
			if err != nil {
				return err
			}
			payLoad, readErr = ioutil.ReadAll(reader)
			//删除请求头参数
			res.Header.Del("Content-Encoding")
		} else {
			//直接读取内容
			payLoad, readErr = ioutil.ReadAll(res.Body)
		}

		if readErr != nil {
			return readErr
		}

		//将预读的数据重新写回
		c.Set("status_code", res.StatusCode)
		c.Set("payload", payLoad)
		res.Body = ioutil.NopCloser(bytes.NewBuffer(payLoad))
		res.ContentLength = int64(len(payLoad))
		res.Header.Set("Content-Length", strconv.FormatInt(res.ContentLength, 10))

		return nil
	}

	//错误回调函数 范围：transport.RoundTrip发生的错误、以及ModifyResponse发生的错误
	errFunc := func(res http.ResponseWriter, req *http.Request, err error) {
		log.Printf("Reverse Proxy Err:%v\n", err)
		middleware.ResponseError(c, 9999, err)
	}

	return &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifier,
		Transport:      transport,
		ErrorHandler:   errFunc,
	}
}

// 使用/连接可用服务路径和原始请求的路径
func singleJoiningSlash(a, b string) string {
	aSlash := strings.HasSuffix(a, "/")
	bSlash := strings.HasPrefix(b, "/")
	if aSlash && bSlash {
		return a + b[1:]
	} else if aSlash || bSlash {
		return a + b
	} else {
		return a + "/" + b
	}
}
