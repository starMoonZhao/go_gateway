package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TcpRule struct {
	ID        int64 `json:"id" gorm:"primary_key"`
	ServiceID int64 `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	Port      int   `json:"port" gorm:"column:port" description:"端口	"`
}

func (tcpRule *TcpRule) TableName() string {
	return "gateway_service_tcp_rule"
}

func (tcpRule *TcpRule) Find(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Where(tcpRule).Find(tcpRule).Error
}

func (tcpRule *TcpRule) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Save(tcpRule).Error
}
