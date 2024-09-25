package router

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/starMoonZhao/go_gateway/controller"
	"github.com/starMoonZhao/go_gateway/docs"
	"github.com/starMoonZhao/go_gateway/middleware"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"log"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server celler server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @query.collection.format multi

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @securitydefinitions.oauth2.application OAuth2Application
// @tokenUrl https://example.com/oauth/token
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.implicit OAuth2Implicit
// @authorizationurl https://example.com/oauth/authorize
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.password OAuth2Password
// @tokenUrl https://example.com/oauth/token
// @scope.read Grants read access
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.accessCode OAuth2AccessCode
// @tokenUrl https://example.com/oauth/token
// @authorizationurl https://example.com/oauth/authorize
// @scope.admin Grants read and write access to administrative information

// @x-extension-openapi {"example": "value on a json format"}

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	// programatically set swagger info
	docs.SwaggerInfo.Title = lib.GetStringConf("base.swagger.title")
	docs.SwaggerInfo.Description = lib.GetStringConf("base.swagger.desc")
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = lib.GetStringConf("base.swagger.host")
	docs.SwaggerInfo.BasePath = lib.GetStringConf("base.swagger.base_path")
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	router := gin.Default()
	router.Use(middlewares...)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//注册登录模块路由
	adminLoginRouter := router.Group("/admin_login")
	//创建redis存储session
	redisAddress := lib.GetConf("redis_map.list.default.proxy_list").([]interface{})[0].(string)
	password := lib.GetStringConf("redis_map.list.default.password")
	maxIdle := lib.GetIntConf("redis_map.list.default.max_idle")
	//初始化redis类型的缓存存储
	redisStore, err := sessions.NewRedisStore(maxIdle, "tcp", redisAddress, password, []byte("secret"))
	if err != nil {
		log.Fatalf("sessions.NewRedisStore err: %v", err)
	}
	//向该路由注册所需的中间件
	adminLoginRouter.Use(sessions.Sessions("mysession", redisStore),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.TranslationMiddleware())
	{
		controller.AdminLoginRegister(adminLoginRouter)
	}

	//注册登录后信息模块路由
	adminRouter := router.Group("/admin")
	//向该路由注册所需的中间件
	adminRouter.Use(sessions.Sessions("mysession", redisStore),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.SessionAuthMiddleware(),
		middleware.TranslationMiddleware())
	{
		controller.AdminRegister(adminRouter)
	}

	//注册服务管理模块路由
	serviceRouter := router.Group("/service")
	//向该路由注册所需的中间件
	serviceRouter.Use(sessions.Sessions("mysession", redisStore),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.SessionAuthMiddleware(),
		middleware.TranslationMiddleware())
	{
		controller.ServiceRegister(serviceRouter)
	}

	//注册租户管理模块路由
	appRouter := router.Group("/app")
	//向该路由注册所需的中间件
	appRouter.Use(sessions.Sessions("mysession", redisStore),
		middleware.RecoveryMiddleware(),
		middleware.RequestLog(),
		middleware.SessionAuthMiddleware(),
		middleware.TranslationMiddleware())
	{
		controller.APPRegister(appRouter)
	}
	return router
}
