package dao

import (
	"errors"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/starMoonZhao/go_gateway/dto"
	"github.com/starMoonZhao/go_gateway/public"
	"gorm.io/gorm"
	"net/http/httptest"
	"strings"
	"sync"
)

// 初始化函数
func init() {
	ServiceManegerHandler = NewServiceManager()
}

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

var ServiceManegerHandler *ServiceManger

type ServiceManger struct {
	ServiceMap   map[string]*ServiceDetail
	ServiceSlice []*ServiceDetail
	Locker       sync.RWMutex
	init         sync.Once
	err          error
}

func NewServiceManager() *ServiceManger {
	return &ServiceManger{
		ServiceMap:   map[string]*ServiceDetail{},
		ServiceSlice: []*ServiceDetail{},
		Locker:       sync.RWMutex{},
		init:         sync.Once{},
	}
}

// 系统初始化时加载服务信息
func (serviceManger *ServiceManger) LoadOnce() error {
	serviceManger.init.Do(func() {
		//查询所有的服务信息
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		tx, err := lib.GetGormPool("default")
		if err != nil {
			serviceManger.err = err
			return
		}
		//分页条件
		params := &dto.ServiceListInput{PageNum: 1, PageSize: 99999}
		serviceInfo := &ServiceInfo{}
		list, _, err := serviceInfo.PageList(c, tx, params)
		if err != nil {
			serviceManger.err = err
			return
		}

		serviceManger.Locker.Lock()
		defer serviceManger.Locker.Unlock()
		//将查询出的所有服务填充到ServiceManger中的ServiceMap、ServiceSlice
		for _, serviceItem := range list {
			tmpServiceItem := serviceItem
			//查询该服务对应的详情
			serviceDetail, err := tmpServiceItem.ServiceDetail(c, tx)
			if err != nil {
				serviceManger.err = err
				return
			}
			serviceManger.ServiceMap[tmpServiceItem.ServiceName] = serviceDetail
			serviceManger.ServiceSlice = append(serviceManger.ServiceSlice, serviceDetail)
		}
	})
	return serviceManger.err
}

// 进行http访问的校验
func (serviceManger *ServiceManger) HTTPAccessMode(c *gin.Context) (*ServiceDetail, error) {
	//HTTP的匹配规则
	//1、前缀匹配 /abc ==> ServiceSlice.rule
	//2、域名匹配 www.test.com ==> ServiceSlice.rule

	//host:c.Request.host path:c.Request.URL.Path
	host := c.Request.Host
	host = host[0:strings.Index(host, ":")]
	path := c.Request.URL.Path

	//循环遍历ServiceSlice是否有符合规则的HTTP服务 如无则抛出异常
	for _, serviceItem := range serviceManger.ServiceSlice {
		//校验服务类型
		if public.LoadTypeHTTP != serviceItem.Info.LoadType {
			continue
		}
		//查看该http服务的规则类型
		if public.HTTPRuleTypePrefixURL == serviceItem.HTTPRule.RuleType {
			//前缀匹配
			if strings.HasPrefix(path, serviceItem.HTTPRule.Rule) {
				return serviceItem, nil
			}
		} else if public.HTTPRuleTypeDomain == serviceItem.HTTPRule.RuleType {
			//域名匹配
			if host == serviceItem.HTTPRule.Rule {
				return serviceItem, nil
			}
		}
	}
	return nil, errors.New("not matched service.")
}

func (s *ServiceManger) GetTCPServiceList() []*ServiceDetail {
	serviceList := []*ServiceDetail{}
	for _, serviceItem := range s.ServiceSlice {
		tempServiceItem := serviceItem
		if tempServiceItem.Info.LoadType == public.LoadTypeTCP {
			serviceList = append(serviceList, tempServiceItem)
		}
	}
	return serviceList
}

func (s *ServiceManger) GetGRPCServiceList() []*ServiceDetail {
	serviceList := []*ServiceDetail{}
	for _, serviceItem := range s.ServiceSlice {
		tempServiceItem := serviceItem
		if tempServiceItem.Info.LoadType == public.LoadTypeGRPC {
			serviceList = append(serviceList, tempServiceItem)
		}
	}
	return serviceList
}
