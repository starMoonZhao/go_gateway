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
	"time"
)

type APPController struct {
}

func APPRegister(group *gin.RouterGroup) {
	appController := &APPController{}
	//注册路由
	group.GET("/app_list", appController.APPList)
	group.DELETE("/app_delete", appController.APPDelete)
	group.GET("/app_detail", appController.APPDetail)
	group.GET("/app_stat", appController.APPStat)
	group.POST("/app_add", appController.APPAdd)
	group.PUT("/app_update", appController.APPUpdate)
}

// APPList godoc
// @Summary 租户信息列表查询
// @Description 租户信息列表查询
// @Tags 租户管理
// @ID /app/app_list
// @Accept  json
// @Produce  json
// @Param info query string false "关键词"
// @Param page_num query int64 true "页码"
// @Param page_size query int64 true "条数"
// @Success 200 {object} middleware.Response{data=dto.APPListOutput} "success"
// @Router /app/app_list [get]
func (appController *APPController) APPList(c *gin.Context) {
	appListInput := &dto.APPListInput{}
	if err := appListInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 4011, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 4012, err)
		return
	}

	//分页查询app基本信息
	app := &dao.APP{}
	appList, total, err := app.PageList(c, tx, appListInput)
	if err != nil {
		middleware.ResponseError(c, 4013, err)
		return
	}

	appListOutput := []dto.APPListItemOutput{}
	for _, appItem := range appList {
		//获取服务流量统计器
		flowCount, err := circuit_rate.FlowCounterHandler.GetFlowCounter(fmt.Sprintf("%s_%s", public.FlowApp, appItem.APPID))
		if err != nil {
			middleware.ResponseError(c, 4014, err)
			return
		}
		appItemOutput := dto.APPListItemOutput{
			ID:       appItem.ID,
			AppID:    appItem.APPID,
			Name:     appItem.Name,
			Secret:   appItem.Secret,
			WhiteIPS: appItem.WhiteIPS,
			Qps:      appItem.Qps,
			Qpd:      appItem.Qpd,
			RealQps:  flowCount.QPS,
			RealQpd:  flowCount.TotalCount,
		}
		appListOutput = append(appListOutput, appItemOutput)
	}

	//封装输出信息
	out := &dto.APPListOutput{
		Total: total,
		List:  appListOutput,
	}

	middleware.ResponseSuccess(c, out)
}

// APPDelete godoc
// @Summary 租户信息删除
// @Description 租户信息删除
// @Tags 租户管理
// @ID /app/app_delete
// @Accept  json
// @Produce  json
// @Param id query dto.APPDeleteInput true "删除租户id"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_delete [delete]
func (appController *APPController) APPDelete(c *gin.Context) {
	appDeleteInput := &dto.APPDeleteInput{}
	if err := appDeleteInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 4021, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 4022, err)
		return
	}

	//读取待删除的服务基本信息
	app := &dao.APP{ID: appDeleteInput.ID}
	if err := app.Find(c, tx); err != nil {
		middleware.ResponseError(c, 4023, err)
		return
	}

	//删除服务
	app.IsDelete = 1
	if err := app.Save(c, tx); err != nil {
		middleware.ResponseError(c, 4024, err)
		return
	}

	middleware.ResponseSuccess(c, "")
}

// APPAdd godoc
// @Summary 租户新增
// @Description 租户新增
// @Tags 租户管理
// @ID /app/app_add
// @Accept  json
// @Produce  json
// @Param body body dto.APPAddInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_add [post]
func (appController *APPController) APPAdd(c *gin.Context) {
	appAddInput := &dto.APPAddInput{}
	if err := appAddInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 4031, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 4032, err)
		return
	}

	//开启事务
	tx = tx.Begin()

	//查看租户标志app_id是否被占用
	app := &dao.APP{
		APPID:    appAddInput.AppID,
		IsDelete: 0,
	}
	app.Find(c, tx)
	if app.ID != 0 {
		tx.Rollback()
		middleware.ResponseError(c, 4033, errors.New("租户已经存在"))
		return
	}

	//如果未输入密钥 则以app_id以md5算法生成密钥
	if appAddInput.Secret == "" {
		appAddInput.Secret = public.MD5(appAddInput.AppID)
	}

	//保存租户基本信息
	app.Name = appAddInput.Name
	app.Secret = appAddInput.Secret
	app.WhiteIPS = appAddInput.WhiteIPS
	app.Qps = appAddInput.Qps
	app.Qpd = appAddInput.Qpd
	if err := app.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 4034, err)
		return
	}

	//提交事务
	tx.Commit()

	middleware.ResponseSuccess(c, app.ID)
}

// APPUpdate godoc
// @Summary 租户更新
// @Description 租户更新
// @Tags 租户管理
// @ID /app/app_update
// @Accept  json
// @Produce  json
// @Param body body dto.APPUpdateInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /app/app_update [put]
func (appController *APPController) APPUpdate(c *gin.Context) {
	appUpdateInput := &dto.APPUpdateInput{}
	if err := appUpdateInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 4041, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 4042, err)
		return
	}

	//开启事务
	tx = tx.Begin()

	//查看租户是否存在
	app := &dao.APP{
		ID:    appUpdateInput.ID,
		APPID: appUpdateInput.AppID,
	}

	if err := app.Find(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 4043, errors.New("租户不存在"))
		return
	}

	//如果未输入密钥 则以app_id以md5算法生成密钥
	if appUpdateInput.Secret == "" {
		appUpdateInput.Secret = public.MD5(appUpdateInput.AppID)
	}

	//更新租户基本信息
	app.Name = appUpdateInput.Name
	app.Secret = appUpdateInput.Secret
	app.WhiteIPS = appUpdateInput.WhiteIPS
	app.Qps = appUpdateInput.Qps
	app.Qpd = appUpdateInput.Qpd
	if err := app.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 4044, err)
		return
	}

	//提交事务
	tx.Commit()

	middleware.ResponseSuccess(c, app.ID)
}

// APPDetail godoc
// @Summary 租户详情查询
// @Description 租户详情查询
// @Tags 租户管理
// @ID /app/app_detail
// @Accept  json
// @Produce  json
// @Param id query dto.APPDetailInput true "租户详情查询id"
// @Success 200 {object} middleware.Response{data=dao.APP} "success"
// @Router /app/app_detail [get]
func (appController *APPController) APPDetail(c *gin.Context) {
	appDetailInput := &dto.APPDetailInput{}
	if err := appDetailInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 4051, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 4052, err)
		return
	}

	//查询租户基本信息
	app := &dao.APP{ID: appDetailInput.ID}
	if err := app.Find(c, tx); err != nil {
		middleware.ResponseError(c, 4053, err)
		return
	}

	middleware.ResponseSuccess(c, app)
}

// APPStat godoc
// @Summary 租户统计信息查询
// @Description 租户统计信息查询
// @Tags 租户管理
// @ID /app/app_stat
// @Accept  json
// @Produce  json
// @Param id query dto.APPStatisticsInput true "服务统计信息查询id"
// @Success 200 {object} middleware.Response{data=dto.APPStatisticsOutput} "success"
// @Router /app/app_stat [get]
func (appController *APPController) APPStat(c *gin.Context) {
	appStatisticsInput := &dto.APPStatisticsInput{}
	if err := appStatisticsInput.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 4061, err)
		return
	}

	//获取数据库连接池
	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 4062, err)
		return
	}

	//查询租户基本信息
	appInfo := &dao.APP{ID: appStatisticsInput.ID}
	if err := appInfo.Find(c, tx); err != nil {
		middleware.ResponseError(c, 4063, err)
		return
	}

	//获取服务流量统计器
	flowCount, err := circuit_rate.FlowCounterHandler.GetFlowCounter(fmt.Sprintf("%s_%s", public.FlowApp, appInfo.APPID))

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

	middleware.ResponseSuccess(c, &dto.APPStatisticsOutput{
		Today:     todayList,
		Yesterday: yesterdayList,
	})
}
