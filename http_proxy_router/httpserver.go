package http_proxy_router

import (
	"context"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

var (
	HttpSrvHandler  *http.Server
	HttpsSrvHandler *http.Server
)

// 启动http服务器
func HttpServerRun() {
	gin.SetMode(lib.GetStringConf("proxy.base.debug_mode"))
	//创建路由
	router := InitRouter()
	//根据配置创建http代理服务器
	HttpSrvHandler := &http.Server{
		Addr:           lib.GetStringConf("proxy.http.addr"),
		Handler:        router,
		ReadTimeout:    time.Duration(lib.GetIntConf("proxy.http.read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(lib.GetIntConf("proxy.http.write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << uint(lib.GetIntConf("proxy.http.max_header_bytes")),
	}
	//启动服务器
	log.Println(" [INFO] HttpServerRun:%s\n", lib.GetStringConf("proxy.http.addr"))
	if err := HttpSrvHandler.ListenAndServe(); err != nil {
		log.Fatalf(" [ERROR] HttpServerRun:%s err:%v\n", lib.GetStringConf("proxy.http.addr"), err)
	}
}

// 停止http服务器
func HttpServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpSrvHandler.Shutdown(ctx); err != nil {
		log.Fatalf(" [ERROR] HttpServerStop err:%v\n", err)
	}
	log.Printf(" [INFO] HttpServerStop stopped\n")
}

// 启动https服务器
func HttpsServerRun() {
	gin.SetMode(lib.GetStringConf("proxy.base.debug_mode"))
	//创建路由
	router := InitRouter()
	//根据配置创建http代理服务器
	HttpsSrvHandler := &http.Server{
		Addr:           lib.GetStringConf("proxy.https.addr"),
		Handler:        router,
		ReadTimeout:    time.Duration(lib.GetIntConf("proxy.https.read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(lib.GetIntConf("proxy.https.write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << uint(lib.GetIntConf("proxy.https.max_header_bytes")),
	}
	//启动https服务器
	log.Println(" [INFO] HttpsServerRun:%s\n", lib.GetStringConf("proxy.https.addr"))
	if err := HttpsSrvHandler.ListenAndServeTLS("./cert_file/server.crt", "./cert_file/server.key"); err != nil {
		log.Fatalf(" [ERROR] HttpsServerRun:%s err:%v\n", lib.GetStringConf("proxy.https.addr"), err)
	}
}

// 停止http服务器
func HttpsServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpsSrvHandler.Shutdown(ctx); err != nil {
		log.Fatalf(" [ERROR] HttpsServerStop err:%v\n", err)
	}
	log.Printf(" [INFO] HttpsServerStop stopped\n")
}
