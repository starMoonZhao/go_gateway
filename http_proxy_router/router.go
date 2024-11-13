package http_proxy_router

import (
	"github.com/gin-gonic/gin"
	"github.com/starMoonZhao/go_gateway/http_proxy_middleware"
)

func InitRouter(middleWares ...gin.HandlerFunc) *gin.Engine {
	router := gin.Default()
	router.Use(middleWares...)

	//注册ping路由
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "PONG"})
	})

	//注册该路由使用的中间件
	router.Use(http_proxy_middleware.HTTPAccessModeMiddleware())
	router.Use(http_proxy_middleware.HTTPHeaderTransferMiddleware())
	router.Use(http_proxy_middleware.HTTPStripUriMiddleware())
	router.Use(http_proxy_middleware.HTTPURLRewriteMiddleware())
	router.Use(http_proxy_middleware.HTTPReverseProxyMiddleware())

	return router
}
