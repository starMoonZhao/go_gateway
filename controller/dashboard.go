package controller

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/starMoonZhao/go_gateway/dao"
	"github.com/starMoonZhao/go_gateway/dto"
	"github.com/starMoonZhao/go_gateway/middleware"
	"github.com/starMoonZhao/go_gateway/public"
	"time"
)

type DashboardController struct {
}

func DashboardRegister(group *gin.RouterGroup) {
	dashboardController := &DashboardController{}
	//注册路由
	group.GET("/panel_group_data", dashboardController.PanelGroupData)
	group.GET("/flow_stat", dashboardController.FlowStat)
	group.GET("/service_stat", dashboardController.ServiceStat)
}

// PanelGroupData godoc
// @Summary 指标统计
// @Description 指标统计
// @Tags 首页大盘
// @ID /dashboard/panel_group_data
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.PanelGroupDataOutput} "success"
// @Router /dashboard/panel_group_data [get]
func (dashboardController *DashboardController) PanelGroupData(c *gin.Context) {
	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 5011, err)
		return
	}

	//查询服务总数
	serviceInfo := &dao.ServiceInfo{}
	_, serviceTotal, err := serviceInfo.PageList(c, tx, &dto.ServiceListInput{PageSize: 1, PageNum: 1})
	if err != nil {
		middleware.ResponseError(c, 5012, err)
		return
	}

	//查询租户总数
	app := &dao.APP{}
	_, appTotal, err := app.PageList(c, tx, &dto.APPListInput{PageSize: 1, PageNum: 1})
	if err != nil {
		middleware.ResponseError(c, 5013, err)
		return
	}

	//封装输出信息
	out := &dto.PanelGroupDataOutput{
		ServiceNum:      serviceTotal,
		AppNum:          appTotal,
		CurrentQPS:      0,
		TodayRequestNum: 0,
	}

	middleware.ResponseSuccess(c, out)
}

// FlowStat godoc
// @Summary 流量统计
// @Description 流量统计
// @Tags 首页大盘
// @ID /dashboard/flow_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /dashboard/flow_stat [get]
func (dashboardController *DashboardController) FlowStat(c *gin.Context) {
	//先构造虚假的统计信息数据

	//查询今日数据
	todayList := []int64{}
	currentTime := time.Now()
	for i := 0; i <= currentTime.Hour(); i++ {
		todayList = append(todayList, 0)
	}

	//查询昨日数据
	yesterdayList := []int64{}
	for i := 0; i <= 23; i++ {
		yesterdayList = append(yesterdayList, 0)
	}

	middleware.ResponseSuccess(c, &dto.ServiceStatOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})
}

// ServiceStat godoc
// @Summary 服务统计
// @Description 服务统计
// @Tags 首页大盘
// @ID /dashboard/service_stat
// @Accept  json
// @Produce  json
// @Success 200 {object} middleware.Response{data=dto.DashServiceStatOutput} "success"
// @Router /dashboard/service_stat [get]
func (dashboardController *DashboardController) ServiceStat(c *gin.Context) {
	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 5031, err)
		return
	}

	//查询服务类型信息
	serviceInfo := &dao.ServiceInfo{}
	dashServiceStatItemOutputList, err := serviceInfo.GroupByLoadType(c, tx)
	if err != nil {
		middleware.ResponseError(c, 5032, err)
		return
	}

	//组装legend
	legend := []string{}
	for index, item := range dashServiceStatItemOutputList {
		name := public.LoadTypeMap[item.LoadType]
		dashServiceStatItemOutputList[index].Name = name
		legend = append(legend, name)
	}

	//封装输出信息
	out := &dto.DashServiceStatOutput{
		Legend: legend,
		Data:   dashServiceStatItemOutputList,
	}

	middleware.ResponseSuccess(c, out)
}
