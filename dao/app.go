package dao

import (
	"github.com/gin-gonic/gin"
	"github.com/starMoonZhao/go_gateway/dto"
	"gorm.io/gorm"
	"time"
)

type APP struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	APPID     string    `json:"app_id" gorm:"column:app_id" description:"租户id	"`
	Name      string    `json:"name" gorm:"column:name" description:"租户名称	"`
	Secret    string    `json:"secret" gorm:"column:secret" description:"密钥"`
	WhiteIPS  string    `json:"white_ips" gorm:"column:white_ips" description:"ip白名单，支持前缀匹配"`
	Qpd       int64     `json:"qpd" gorm:"column:qpd" description:"日请求量限制"`
	Qps       int64     `json:"qps" gorm:"column:qps" description:"每秒请求量限制"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"添加时间	"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete  int8      `json:"is_delete" gorm:"column:is_delete" description:"是否已删除；0：否；1：是"`
}

func (app *APP) TableName() string {
	return "gateway_app"
}

func (app *APP) Find(c *gin.Context, tx *gorm.DB) error {
	if err := tx.WithContext(c).Where(app).Find(app).Error; err != nil {
		return err
	}
	return nil
}

func (app *APP) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.WithContext(c).Save(app).Error; err != nil {
		return err
	}
	return nil
}

// 租户列表信息分页查询
func (app *APP) PageList(c *gin.Context, tx *gorm.DB, param *dto.APPListInput) ([]APP, int64, error) {
	//总条数
	total := int64(0)
	//结果集
	list := []APP{}

	//分页查询偏移量
	offset := int((param.PageNum - 1) * param.PageSize)

	query := tx.WithContext(c)
	query = query.Table(app.TableName()).Where("is_delete = 0")

	//构造模糊查询条件
	if param.Info != "" {
		query = query.Where("app_id like ? or name like ?", "%"+param.Info+"%", "%"+param.Info+"%")
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
