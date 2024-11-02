package dao

import (
	"github.com/gin-gonic/gin"
	"github.com/starMoonZhao/go_gateway/dto"
	"gorm.io/gorm"
	"time"
)

type ServiceInfo struct {
	ID          int64     `json:"id" gorm:"primary_key" description:"自增主键"`
	LoadType    int       `json:"load_type" gorm:"column:load_type" description:"负载类型 0=http 1=tcp 2=grpc"`
	ServiceName string    `json:"service_name" gorm:"column:service_name" description:"服务名称"`
	ServiceDesc string    `json:"service_desc" gorm:"column:service_desc" description:"服务描述"`
	CreatedAt   time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	UpdatedAt   time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete    int       `json:"is_delete" gorm:"column:is_delete" description:"是否删除 0=否 1=是"`
}

func (serviceInfo *ServiceInfo) TableName() string {
	return "gateway_service_info"
}

func (serviceInfo *ServiceInfo) Find(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Where(serviceInfo).Find(serviceInfo).Error
}

func (serviceInfo *ServiceInfo) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Save(serviceInfo).Error
}

// 服务列表信息分页查询
func (serviceInfo *ServiceInfo) PageList(c *gin.Context, tx *gorm.DB, param *dto.ServiceListInput) ([]ServiceInfo, int64, error) {
	//总条数
	total := int64(0)
	//结果集
	list := []ServiceInfo{}

	//分页查询偏移量
	offset := int((param.PageNum - 1) * param.PageSize)

	query := tx.WithContext(c)
	query = query.Table(serviceInfo.TableName()).Where("is_delete = 0")

	//构造模糊查询条件
	if param.Info != "" {
		query = query.Where("service_name like ? or service_desc like ?", "%"+param.Info+"%", "%"+param.Info+"%")
	}

	//构造分页查询条件
	query = query.Order("id desc").Offset(offset).Limit(int(param.PageSize))

	if err := query.Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		//不存在符合条件的数据条目
		return nil, 0, err
	}

	//查询总条数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (serviceInfo *ServiceInfo) Delete(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Where(serviceInfo).Delete(serviceInfo).Error
}

// 服务分类信息查询
func (serviceInfo *ServiceInfo) GroupByLoadType(c *gin.Context, tx *gorm.DB) ([]dto.DashServiceStatItemOutput, error) {
	//结果集
	dashServiceStatItemOutputList := []dto.DashServiceStatItemOutput{}

	if err := tx.WithContext(c).Table(serviceInfo.TableName()).Where("is_delete = 0").Select("load_type", "count(*) as value").Group("load_type").Scan(&dashServiceStatItemOutputList).Error; err != nil {
		return nil, err
	}

	return dashServiceStatItemOutputList, nil
}
