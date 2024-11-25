package controller

import (
	"fmt"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/starMoonZhao/go_gateway/circuit_rate"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/dto"
	"github.com/starMoonZhao/go_gateway/middleware"
	"github.com/starMoonZhao/go_gateway/public"
	"strings"
	"time"
)

type ServiceController struct {
}

func ServiceRegister(group *gin.RouterGroup) {
	serviceController := &ServiceController{}
	//注册路由
	group.GET("/service_list", serviceController.ServiceList)
	group.DELETE("/service_delete", serviceController.ServiceDelete)
	group.GET("/service_detail", serviceController.ServiceDetail)
	group.GET("/service_stat", serviceController.ServiceStat)

	group.POST("/service_add_http", serviceController.ServiceAddHTTP)
	group.PUT("/service_update_http", serviceController.ServiceUpdateHTTP)

	group.POST("/service_add_tcp", serviceController.ServiceAddTCP)
	group.PUT("/service_update_tcp", serviceController.ServiceUpdateTCP)

	group.POST("/service_add_grpc", serviceController.ServiceAddGRPC)
	group.PUT("/service_update_grpc", serviceController.ServiceUpdateGRPC)
}

// ServiceList godoc
// @Summary 服务信息列表查询
// @Description 服务信息列表查询
// @Tags 服务管理
// @ID /service/service_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_num query int64 true "页码"
// @Param page_size query int64 true "条数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/service_list [get]
func (serviceController *ServiceController) ServiceList(c *gin.Context) {
	serviceListInput := &dto.ServiceListInput{}
	if err := serviceListInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3011, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 3012, err)
		return
	}

	//分页查询ServiceInfo基本信息
	serviceInfo := &dao.ServiceInfo{}
	serviceInfoList, total, err := serviceInfo.PageList(c, tx, serviceListInput)
	if err != nil {
		middleware.ResponseError(c, 3013, err)
		return
	}

	serviceInfoOutList := []dto.ServiceListItemOutput{}
	for _, serviceInfoItem := range serviceInfoList {
		//根据ServiceInfo基本信息查询完整的服务信息
		serviceDetail, err := serviceInfoItem.ServiceDetail(c, tx)
		if err != nil {
			middleware.ResponseError(c, 3014, err)
			return
		}

		//将完整的服务信息转换为前台页面输出输出对象
		//http后缀接入：clusterIP+clusterPort+path
		//http域名接入：domain
		//grpc或tcp接入：clusterIP+servicePort

		//获取代理服务集群配置信息
		clusterIP := lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort := lib.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort := lib.GetStringConf("base.cluster.cluster_ssl_port")

		//构造服务地址
		serviceAddr := "unknow"
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == public.HTTPDontNeedHttps {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterPort, serviceDetail.HTTPRule.Rule)
		}
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypePrefixURL &&
			serviceDetail.HTTPRule.NeedHttps == public.HTTPNeedHttps {
			serviceAddr = fmt.Sprintf("%s:%s%s", clusterIP, clusterSSLPort, serviceDetail.HTTPRule.Rule)
		}
		if serviceDetail.Info.LoadType == public.LoadTypeHTTP &&
			serviceDetail.HTTPRule.RuleType == public.HTTPRuleTypeDomain {
			serviceAddr = serviceDetail.HTTPRule.Rule
		}
		if serviceDetail.Info.LoadType == public.LoadTypeTCP {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.TCPRule.Port)
		}
		if serviceDetail.Info.LoadType == public.LoadTypeGRPC {
			serviceAddr = fmt.Sprintf("%s:%d", clusterIP, serviceDetail.GRPCRule.Port)
		}

		//获取负载均衡ip列表
		ipList := strings.Split(serviceDetail.LoadBalance.IpList, ",")

		//获取服务流量统计器
		flowCount, err := circuit_rate.FlowCounterHandler.GetFlowCounter(fmt.Sprintf("%s_%s", public.FlowService, serviceDetail.Info.ServiceName))
		if err != nil {
			middleware.ResponseError(c, 3015, err)
			return
		}
		serviceInfoOutputItem := dto.ServiceListItemOutput{
			Id:          serviceInfoItem.ID,
			ServiceName: serviceInfoItem.ServiceName,
			ServiceDesc: serviceInfoItem.ServiceDesc,
			LoadType:    serviceInfoItem.LoadType,
			ServiceAddr: serviceAddr,
			Qps:         flowCount.QPS,
			Qpd:         flowCount.TotalCount,
			TotalNode:   len(ipList),
		}
		serviceInfoOutList = append(serviceInfoOutList, serviceInfoOutputItem)
	}

	//封装输出信息
	out := &dto.ServiceListOutput{
		Total: total,
		List:  serviceInfoOutList,
	}

	middleware.ResponseSuccess(c, out)
}

// ServiceDelete godoc
// @Summary 服务信息删除
// @Description 服务信息删除
// @Tags 服务管理
// @ID /service/service_delete
// @Accept  json
// @Produce  json
// @Param id query dto.ServiceDeleteInput true "删除服务id"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_delete [delete]
func (serviceController *ServiceController) ServiceDelete(c *gin.Context) {
	serviceDeleteInput := &dto.ServiceDeleteInput{}
	if err := serviceDeleteInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3021, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 3022, err)
		return
	}

	//读取待删除的服务基本信息
	serviceInfo := &dao.ServiceInfo{ID: serviceDeleteInput.ID}
	if err := serviceInfo.Find(c, tx); err != nil {
		middleware.ResponseError(c, 3023, err)
		return
	}

	//删除服务
	serviceInfo.IsDelete = 1
	if err := serviceInfo.Save(c, tx); err != nil {
		middleware.ResponseError(c, 3024, err)
		return
	}

	middleware.ResponseSuccess(c, "")
}

// ServiceAddHTTP godoc
// @Summary HTTP服务新增
// @Description HTTP服务新增
// @Tags 服务管理
// @ID /service/service_add_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_http [post]
func (serviceController *ServiceController) ServiceAddHTTP(c *gin.Context) {
	serviceAddHTTPInput := &dto.ServiceAddHTTPInput{}
	if err := serviceAddHTTPInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3031, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 3032, err)
		return
	}

	//开启事务
	tx = tx.Begin()

	//查看服务名称是否被占用
	serviceInfo := &dao.ServiceInfo{
		ServiceName: serviceAddHTTPInput.ServiceName,
	}
	serviceInfo.Find(c, tx)
	if serviceInfo.ID != 0 {
		tx.Rollback()
		middleware.ResponseError(c, 3033, errors.New("服务已经存在"))
		return
	}

	//查看服务接入前缀或域名是否存在
	httpRule := &dao.HttpRule{RuleType: serviceAddHTTPInput.RuleType, Rule: serviceAddHTTPInput.Rule}
	httpRule.Find(c, tx)
	if httpRule.ID != 0 {
		tx.Rollback()
		middleware.ResponseError(c, 3034, errors.New("服务接入前缀或域名已存在"))
		return
	}

	//保存服务基本信息
	serviceInfo.ServiceDesc = serviceAddHTTPInput.ServiceDesc
	if err := serviceInfo.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3035, err)
		return
	}

	serviceId := serviceInfo.ID
	//保存http服务的规则信息
	httpRule.ServiceID = serviceId
	httpRule.NeedHttps = serviceAddHTTPInput.NeedHttps
	httpRule.NeedStripUri = serviceAddHTTPInput.NeedStripUri
	httpRule.NeedWebsocket = serviceAddHTTPInput.NeedWebsocket
	httpRule.UrlRewrite = serviceAddHTTPInput.UrlRewrite
	httpRule.HeaderTransfer = serviceAddHTTPInput.HeaderTransfer
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3036, err)
		return
	}

	//保存http服务的权限控制信息
	accessControl := &dao.AccessControl{
		ServiceID:         serviceId,
		OpenAuth:          serviceAddHTTPInput.OpenAuth,
		BlackList:         serviceAddHTTPInput.BlackList,
		WhiteList:         serviceAddHTTPInput.WhiteList,
		ClientIPFlowLimit: serviceAddHTTPInput.ClientIPFlowLimit,
		ServiceFlowLimit:  serviceAddHTTPInput.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3037, err)
		return
	}

	//保存服务负载均衡信息
	loadBalance := &dao.LoadBalance{
		ServiceID:              serviceId,
		RoundType:              serviceAddHTTPInput.RoundType,
		IpList:                 serviceAddHTTPInput.IpList,
		WeightList:             serviceAddHTTPInput.WeightList,
		UpstreamConnectTimeout: serviceAddHTTPInput.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  serviceAddHTTPInput.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    serviceAddHTTPInput.UpstreamIdleTimeout,
		UpstreamMaxIdle:        serviceAddHTTPInput.UpstreamMaxIdle,
	}
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3038, err)
		return
	}

	//提交事务
	tx.Commit()

	middleware.ResponseSuccess(c, serviceInfo.ID)
}

// ServiceUpdateHTTP godoc
// @Summary HTTP服务更新
// @Description HTTP服务更新
// @Tags 服务管理
// @ID /service/service_update_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_http [put]
func (serviceController *ServiceController) ServiceUpdateHTTP(c *gin.Context) {
	serviceUpdateHTTPInput := &dto.ServiceUpdateHTTPInput{}
	if err := serviceUpdateHTTPInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3041, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 3042, err)
		return
	}

	//开启事务
	tx = tx.Begin()

	//查看服务是否存在
	serviceInfo := &dao.ServiceInfo{
		ID:          serviceUpdateHTTPInput.ID,
		ServiceName: serviceUpdateHTTPInput.ServiceName,
	}

	if err := serviceInfo.Find(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3043, errors.New("服务不存在"))
		return
	}

	//查询服务详情
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3044, errors.New("服务不存在"))
		return
	}

	//更新服务基本信息
	serviceInfo.ServiceDesc = serviceUpdateHTTPInput.ServiceDesc
	if err := serviceInfo.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3045, err)
		return
	}

	//更新http服务的规则信息
	httpRule := serviceDetail.HTTPRule
	httpRule.NeedHttps = serviceUpdateHTTPInput.NeedHttps
	httpRule.NeedStripUri = serviceUpdateHTTPInput.NeedStripUri
	httpRule.NeedWebsocket = serviceUpdateHTTPInput.NeedWebsocket
	httpRule.UrlRewrite = serviceUpdateHTTPInput.UrlRewrite
	httpRule.HeaderTransfer = serviceUpdateHTTPInput.HeaderTransfer
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3046, err)
		return
	}

	//更新http服务的权限控制信息
	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = serviceUpdateHTTPInput.OpenAuth
	accessControl.BlackList = serviceUpdateHTTPInput.BlackList
	accessControl.WhiteList = serviceUpdateHTTPInput.WhiteList
	accessControl.ClientIPFlowLimit = serviceUpdateHTTPInput.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = serviceUpdateHTTPInput.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3047, err)
		return
	}

	//更新服务负载均衡信息
	loadBalance := serviceDetail.LoadBalance
	loadBalance.RoundType = serviceUpdateHTTPInput.RoundType
	loadBalance.IpList = serviceUpdateHTTPInput.IpList
	loadBalance.WeightList = serviceUpdateHTTPInput.WeightList
	loadBalance.UpstreamConnectTimeout = serviceUpdateHTTPInput.UpstreamConnectTimeout
	loadBalance.UpstreamHeaderTimeout = serviceUpdateHTTPInput.UpstreamHeaderTimeout
	loadBalance.UpstreamIdleTimeout = serviceUpdateHTTPInput.UpstreamIdleTimeout
	loadBalance.UpstreamMaxIdle = serviceUpdateHTTPInput.UpstreamMaxIdle
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3048, err)
		return
	}

	//提交事务
	tx.Commit()

	middleware.ResponseSuccess(c, serviceInfo.ID)
}

// ServiceDetail godoc
// @Summary 服务详情查询
// @Description 服务详情查询
// @Tags 服务管理
// @ID /service/service_detail
// @Accept  json
// @Produce  json
// @Param id query dto.ServiceDetailInput true "服务详情查询id"
// @Success 200 {object} middleware.Response{data=dao.ServiceDetail} "success"
// @Router /service/service_detail [get]
func (serviceController *ServiceController) ServiceDetail(c *gin.Context) {
	serviceDetailInput := &dto.ServiceDetailInput{}
	if err := serviceDetailInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3051, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 3052, err)
		return
	}

	//查询服务基本信息
	serviceInfo := &dao.ServiceInfo{ID: serviceDetailInput.ID}
	if err := serviceInfo.Find(c, tx); err != nil {
		middleware.ResponseError(c, 3053, err)
		return
	}

	//查询服务详细信息
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx)
	if err != nil {
		middleware.ResponseError(c, 3054, err)
		return
	}

	middleware.ResponseSuccess(c, serviceDetail)
}

// ServiceStat godoc
// @Summary 服务统计信息查询
// @Description 服务统计信息查询
// @Tags 服务管理
// @ID /service/service_stat
// @Accept  json
// @Produce  json
// @Param id query dto.ServiceStatInput true "服务统计信息查询id"
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /service/service_stat [get]
func (serviceController *ServiceController) ServiceStat(c *gin.Context) {
	serviceDetailInput := &dto.ServiceDetailInput{}
	if err := serviceDetailInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3061, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 3062, err)
		return
	}

	//查询服务基本信息
	serviceInfo := &dao.ServiceInfo{ID: serviceDetailInput.ID}
	if err := serviceInfo.Find(c, tx); err != nil {
		middleware.ResponseError(c, 3063, err)
		return
	}

	//获取服务流量统计器
	flowCount, err := circuit_rate.FlowCounterHandler.GetFlowCounter(fmt.Sprintf("%s_%s", public.FlowService, serviceInfo.ServiceName))

	//查询今日数据
	todayList := []int64{}
	currentTime := time.Now()
	for i := 0; i <= currentTime.Hour(); i++ {
		dateTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := flowCount.GetHourData(dateTime)
		todayList = append(todayList, hourData)
	}

	//查询昨日数据
	yesterdayList := []int64{}
	//昨日时间
	yesterTime := currentTime.Add(-1 * time.Duration(time.Hour*24))
	for i := 0; i <= 23; i++ {
		dateTime := time.Date(yesterTime.Year(), yesterTime.Month(), yesterTime.Day(), i, 0, 0, 0, lib.TimeLocation)
		hourData, _ := flowCount.GetHourData(dateTime)
		yesterdayList = append(yesterdayList, hourData)
	}

	middleware.ResponseSuccess(c, &dto.ServiceStatOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})
}

// ServiceAddTCP godoc
// @Summary TCP服务新增
// @Description TCP服务新增
// @Tags 服务管理
// @ID /service/service_add_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddTCPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_tcp [post]
func (serviceController *ServiceController) ServiceAddTCP(c *gin.Context) {
	serviceAddTCPInput := &dto.ServiceAddTCPInput{}
	if err := serviceAddTCPInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3071, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 3072, err)
		return
	}

	//开启事务
	tx = tx.Begin()

	//查看服务名称是否被占用
	serviceInfo := &dao.ServiceInfo{
		ServiceName: serviceAddTCPInput.ServiceName,
	}
	serviceInfo.Find(c, tx)
	if serviceInfo.ID != 0 {
		tx.Rollback()
		middleware.ResponseError(c, 3073, errors.New("服务已经存在"))
		return
	}

	//验证端口是否被tcp占用
	tcpRule := &dao.TcpRule{
		Port: serviceAddTCPInput.Port,
	}
	if tcpRule.Find(c, tx); tcpRule.ID != 0 {
		tx.Rollback()
		middleware.ResponseError(c, 3074, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//验证端口是否被grpc占用
	grpcRule := &dao.GrpcRule{
		Port: serviceAddTCPInput.Port,
	}
	if grpcRule.Find(c, tx); grpcRule.ID != 0 {
		tx.Rollback()
		middleware.ResponseError(c, 3075, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//保存服务基本信息
	serviceInfo.ServiceDesc = serviceAddTCPInput.ServiceDesc
	serviceInfo.LoadType = public.LoadTypeTCP
	if err := serviceInfo.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3076, err)
		return
	}

	serviceId := serviceInfo.ID
	//保存grpc服务的规则信息
	tcpRule.ServiceID = serviceId
	if err := tcpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3077, err)
		return
	}

	//保存grpc服务的权限控制信息
	accessControl := &dao.AccessControl{
		ServiceID:         serviceId,
		OpenAuth:          serviceAddTCPInput.OpenAuth,
		BlackList:         serviceAddTCPInput.BlackList,
		WhiteList:         serviceAddTCPInput.WhiteList,
		WhiteHostName:     serviceAddTCPInput.WhiteHostName,
		ClientIPFlowLimit: serviceAddTCPInput.ClientIPFlowLimit,
		ServiceFlowLimit:  serviceAddTCPInput.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3078, err)
		return
	}

	//保存服务负载均衡信息
	loadBalance := &dao.LoadBalance{
		ServiceID:  serviceId,
		RoundType:  serviceAddTCPInput.RoundType,
		IpList:     serviceAddTCPInput.IpList,
		WeightList: serviceAddTCPInput.WeightList,
		ForbidList: serviceAddTCPInput.ForbidList,
	}
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3079, err)
		return
	}

	//提交事务
	tx.Commit()

	middleware.ResponseSuccess(c, serviceInfo.ID)
}

// ServiceUpdateTCP godoc
// @Summary TCP服务更新
// @Description TCP服务更新
// @Tags 服务管理
// @ID /service/service_update_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateTCPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_tcp [put]
func (serviceController *ServiceController) ServiceUpdateTCP(c *gin.Context) {
	serviceUpdateTCPInput := &dto.ServiceUpdateTCPInput{}
	if err := serviceUpdateTCPInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3081, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 3082, err)
		return
	}

	//开启事务
	tx = tx.Begin()

	//查看服务是否存在
	serviceInfo := &dao.ServiceInfo{
		ID:          serviceUpdateTCPInput.ID,
		ServiceName: serviceUpdateTCPInput.ServiceName,
		IsDelete:    0,
	}

	if err := serviceInfo.Find(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3083, errors.New("服务不存在"))
		return
	}

	//查询服务详情
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3084, errors.New("服务不存在"))
		return
	}

	//更新服务基本信息
	serviceInfo.ServiceDesc = serviceUpdateTCPInput.ServiceDesc
	if err := serviceInfo.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3085, err)
		return
	}

	//更新grpc服务的规则信息
	tcpRule := serviceDetail.TCPRule
	tcpRule.Port = serviceUpdateTCPInput.Port
	if err := tcpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3086, err)
		return
	}

	//更新grpc服务的权限控制信息
	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = serviceUpdateTCPInput.OpenAuth
	accessControl.BlackList = serviceUpdateTCPInput.BlackList
	accessControl.WhiteList = serviceUpdateTCPInput.WhiteList
	accessControl.WhiteHostName = serviceUpdateTCPInput.WhiteHostName
	accessControl.ClientIPFlowLimit = serviceUpdateTCPInput.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = serviceUpdateTCPInput.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3087, err)
		return
	}

	//更新服务负载均衡信息
	loadBalance := serviceDetail.LoadBalance
	loadBalance.RoundType = serviceUpdateTCPInput.RoundType
	loadBalance.IpList = serviceUpdateTCPInput.IpList
	loadBalance.WeightList = serviceUpdateTCPInput.WeightList
	loadBalance.ForbidList = serviceUpdateTCPInput.ForbidList
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3088, err)
		return
	}

	//提交事务
	tx.Commit()

	middleware.ResponseSuccess(c, serviceInfo.ID)
}

// ServiceAddGRPC godoc
// @Summary GRPC服务新增
// @Description GRPC服务新增
// @Tags 服务管理
// @ID /service/service_add_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddGRPCInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_grpc [post]
func (serviceController *ServiceController) ServiceAddGRPC(c *gin.Context) {
	serviceAddGRPCInput := &dto.ServiceAddGRPCInput{}
	if err := serviceAddGRPCInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3091, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 3092, err)
		return
	}

	//开启事务
	tx = tx.Begin()

	//查看服务名称是否被占用
	serviceInfo := &dao.ServiceInfo{
		ServiceName: serviceAddGRPCInput.ServiceName,
	}
	serviceInfo.Find(c, tx)
	if serviceInfo.ID != 0 {
		tx.Rollback()
		middleware.ResponseError(c, 3093, errors.New("服务已经存在"))
		return
	}

	//验证端口是否被tcp占用
	tcpRule := &dao.TcpRule{
		Port: serviceAddGRPCInput.Port,
	}
	if err := tcpRule.Find(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3094, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//验证端口是否被grpc占用
	grpcRule := &dao.GrpcRule{
		Port: serviceAddGRPCInput.Port,
	}
	if err := grpcRule.Find(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3095, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//保存服务基本信息
	serviceInfo.ServiceDesc = serviceAddGRPCInput.ServiceDesc
	serviceInfo.LoadType = public.LoadTypeGRPC
	if err := serviceInfo.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3096, err)
		return
	}

	serviceId := serviceInfo.ID
	//保存grpc服务的规则信息
	grpcRule.ServiceID = serviceId
	grpcRule.HeaderTransfer = serviceAddGRPCInput.HeaderTransfer
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3097, err)
		return
	}

	//保存grpc服务的权限控制信息
	accessControl := &dao.AccessControl{
		ServiceID:         serviceId,
		OpenAuth:          serviceAddGRPCInput.OpenAuth,
		BlackList:         serviceAddGRPCInput.BlackList,
		WhiteList:         serviceAddGRPCInput.WhiteList,
		WhiteHostName:     serviceAddGRPCInput.WhiteHostName,
		ClientIPFlowLimit: serviceAddGRPCInput.ClientIPFlowLimit,
		ServiceFlowLimit:  serviceAddGRPCInput.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3098, err)
		return
	}

	//保存服务负载均衡信息
	loadBalance := &dao.LoadBalance{
		ServiceID:  serviceId,
		RoundType:  serviceAddGRPCInput.RoundType,
		IpList:     serviceAddGRPCInput.IpList,
		WeightList: serviceAddGRPCInput.WeightList,
		ForbidList: serviceAddGRPCInput.ForbidList,
	}
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3099, err)
		return
	}

	//提交事务
	tx.Commit()

	middleware.ResponseSuccess(c, serviceInfo.ID)
}

// ServiceUpdateGRPC godoc
// @Summary GRPC服务更新
// @Description GRPC服务更新
// @Tags 服务管理
// @ID /service/service_update_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateGRPCInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_grpc [put]
func (serviceController *ServiceController) ServiceUpdateGRPC(c *gin.Context) {
	serviceUpdateGRPCInput := &dto.ServiceUpdateGRPCInput{}
	if err := serviceUpdateGRPCInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3101, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 3102, err)
		return
	}

	//开启事务
	tx = tx.Begin()

	//查看服务是否存在
	serviceInfo := &dao.ServiceInfo{
		ID:          serviceUpdateGRPCInput.ID,
		ServiceName: serviceUpdateGRPCInput.ServiceName,
		IsDelete:    0,
	}

	if err := serviceInfo.Find(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3103, errors.New("服务不存在"))
		return
	}

	//查询服务详情
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3104, errors.New("服务不存在"))
		return
	}

	//更新服务基本信息
	serviceInfo.ServiceDesc = serviceUpdateGRPCInput.ServiceDesc
	if err := serviceInfo.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3105, err)
		return
	}

	//更新grpc服务的规则信息
	grpcRule := serviceDetail.GRPCRule
	grpcRule.Port = serviceUpdateGRPCInput.Port
	grpcRule.HeaderTransfer = serviceUpdateGRPCInput.HeaderTransfer
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3106, err)
		return
	}

	//更新grpc服务的权限控制信息
	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = serviceUpdateGRPCInput.OpenAuth
	accessControl.BlackList = serviceUpdateGRPCInput.BlackList
	accessControl.WhiteList = serviceUpdateGRPCInput.WhiteList
	accessControl.WhiteHostName = serviceUpdateGRPCInput.WhiteHostName
	accessControl.ClientIPFlowLimit = serviceUpdateGRPCInput.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = serviceUpdateGRPCInput.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3107, err)
		return
	}

	//更新服务负载均衡信息
	loadBalance := serviceDetail.LoadBalance
	loadBalance.RoundType = serviceUpdateGRPCInput.RoundType
	loadBalance.IpList = serviceUpdateGRPCInput.IpList
	loadBalance.WeightList = serviceUpdateGRPCInput.WeightList
	loadBalance.ForbidList = serviceUpdateGRPCInput.ForbidList
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3108, err)
		return
	}

	//提交事务
	tx.Commit()

	middleware.ResponseSuccess(c, serviceInfo.ID)
}
