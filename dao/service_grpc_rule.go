package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GrpcRule struct {
	ID             int64  `json:"id" gorm:"primary_key"`
	ServiceID      int64  `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	Port           int    `json:"port" gorm:"column:port" description:"端口	"`
	HeaderTransfer string `json:"header_transfer" gorm:"column:header_transfer" description:"header转换支持增加(add)、删除(del)、修改(edit) 格式: add headname headvalue"`
}

func (grpcRule *GrpcRule) TableName() string {
	return "gateway_service_grpc_rule"
}

func (grpcRule *GrpcRule) Find(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Where(grpcRule).Find(grpcRule).Error
}

func (grpcRule *GrpcRule) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Save(grpcRule).Error
}
