package main

import (
	"flag"
	"github.com/e421083458/golang_common/lib"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/grpc_proxy_router"
	"github.com/starMoonZhao/go_gateway/http_proxy_router"
	"github.com/starMoonZhao/go_gateway/router"
	"github.com/starMoonZhao/go_gateway/tcp_server"
	"os"
	"os/signal"
	"syscall"
)

var (
	endpoint = flag.String("endpoint", "", "input endpoint dashboard or server")
	config   = flag.String("config", "", "input config file like ./conf/dev/")
)

func main() {
	//校验命令行输入的启动参数
	flag.Parse()
	if *endpoint == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *config == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *endpoint == "dashboard" {
		lib.InitModule(*config, []string{"base", "mysql", "redis"})
		defer lib.Destroy()
		router.HttpServerRun()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		router.HttpServerStop()
	} else {
		lib.InitModule(*config, []string{"base", "mysql", "redis"})
		defer lib.Destroy()

		//系统启动 加载服务信息
		dao.ServiceManegerHandler.LoadOnce()

		//系统启动 加载租户信息
		dao.AppManegerHandler.LoadOnce()

		//启动http代理服务器
		go func() {
			http_proxy_router.HttpServerRun()
		}()
		//启动https代理服务器
		go func() {
			http_proxy_router.HttpsServerRun()
		}()
		//启动tcp代理服务器
		go func() {
			tcp_server.TCPServerRun()
		}()
		//启动grpc代理服务器
		go func() {
			grpc_proxy_router.GrpcServerRun()
		}()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		//停止http代理服务器
		http_proxy_router.HttpServerStop()

		//停止https代理服务器
		http_proxy_router.HttpsServerStop()

		//停止tcp代理服务器
		tcp_server.TCPServerStop()

		//停止grpc代理服务器
		grpc_proxy_router.GrpcServerStop()
	}
	/*	lib.InitModule("./conf/dev/", []string{"base", "mysql", "redis"})
		defer lib.Destroy()
		router.HttpServerRun()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		router.HttpServerStop()*/

}
