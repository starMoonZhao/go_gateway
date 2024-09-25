package dao

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ServiceDetail struct {
	Info          *ServiceInfo   `json:"info" description:"基本信息"`
	HTTPRule      *HttpRule      `json:"http_rule" description:"http_rule"`
	TCPRule       *TcpRule       `json:"tcp_rule" description:"tcp_rule"`
	GRPCRule      *GrpcRule      `json:"grpc_rule" description:"grpc_rule"`
	LoadBalance   *LoadBalance   `json:"load_balance" description:"load_balance"`
	AccessControl *AccessControl `json:"access_control" description:"access_control"`
}

func (serviceInfo *ServiceInfo) ServiceDetail(c *gin.Context, tx *gorm.DB) (*ServiceDetail, error) {
	httpRule := &HttpRule{
		ServiceID: serviceInfo.ID,
	}
	if err := httpRule.Find(c, tx); err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	tcpRule := &TcpRule{
		ServiceID: serviceInfo.ID,
	}
	if err := tcpRule.Find(c, tx); err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	grpcRule := &GrpcRule{
		ServiceID: serviceInfo.ID,
	}
	if err := grpcRule.Find(c, tx); err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	loadBalance := &LoadBalance{
		ServiceID: serviceInfo.ID,
	}
	if err := loadBalance.Find(c, tx); err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	accessControl := &AccessControl{
		ServiceID: serviceInfo.ID,
	}
	if err := accessControl.Find(c, tx); err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	//构造ServiceDetail对象
	serviceDetail := &ServiceDetail{
		Info:          serviceInfo,
		HTTPRule:      httpRule,
		TCPRule:       tcpRule,
		GRPCRule:      grpcRule,
		LoadBalance:   loadBalance,
		AccessControl: accessControl,
	}
	return serviceDetail, nil
}
