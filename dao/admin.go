package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type Admin struct {
	Id        int       `json:"id" gorm:"primary_key" description:"自增主键"`
	UserName  string    `json:"username" gorm:"column:user_name" description:"用户名称"`
	Salt      string    `json:"salt" gorm:"column:salt" description:"盐"`
	Password  string    `json:"password" gorm:"column:password" description:"密码"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"创建时间"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete  int       `json:"is_delete" gorm:"column:is_delete" description:"是否删除"`
}

func (admin *Admin) TableName() string {
	return "gateway_admin"
}

func (admin *Admin) Find(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Where(admin).Find(admin).Error
}

func (admin *Admin) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Save(admin).Error
}
