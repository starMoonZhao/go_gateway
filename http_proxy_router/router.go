package http_proxy_router

import "github.com/gin-gonic/gin"

func InitRouter(middleWares ...gin.HandlerFunc) *gin.Engine {
	router := gin.Default()
	router.Use(middleWares...)

	//注册ping路由
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "PONG"})
	})

	return router
}
