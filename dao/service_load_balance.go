package dao

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/starMoonZhao/go_gateway/public"
	"github.com/starMoonZhao/go_gateway/reverse_proxy/load_balance"
	"gorm.io/gorm"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type LoadBalance struct {
	ID                     int64  `json:"id" gorm:"primary_key"`
	ServiceID              int64  `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	CheckMethod            int    `json:"check_method" gorm:"column:check_method" description:"检查方法 tcpchk=检测端口是否握手成功	"`
	CheckTimeout           int    `json:"check_timeout" gorm:"column:check_timeout" description:"check超时时间	"`
	CheckInterval          int    `json:"check_interval" gorm:"column:check_interval" description:"检查间隔, 单位s		"`
	RoundType              int    `json:"round_type" gorm:"column:round_type" description:"轮询方式 round/weight_round/random/ip_hash"`
	IpList                 string `json:"ip_list" gorm:"column:ip_list" description:"ip列表"`
	WeightList             string `json:"weight_list" gorm:"column:weight_list" description:"权重列表"`
	ForbidList             string `json:"forbid_list" gorm:"column:forbid_list" description:"禁用ip列表"`
	UpstreamConnectTimeout int    `json:"upstream_connect_timeout" gorm:"column:upstream_connect_timeout" description:"下游建立连接超时, 单位s"`
	UpstreamHeaderTimeout  int    `json:"upstream_header_timeout" gorm:"column:upstream_header_timeout" description:"下游获取header超时, 单位s	"`
	UpstreamIdleTimeout    int    `json:"upstream_idle_timeout" gorm:"column:upstream_idle_timeout" description:"下游链接最大空闲时间, 单位s	"`
	UpstreamMaxIdle        int    `json:"upstream_max_idle" gorm:"column:upstream_max_idle" description:"下游最大空闲链接数"`
}

func (loadBalance *LoadBalance) TableName() string {
	return "gateway_service_load_balance"
}

func (loadBalance *LoadBalance) Find(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Where(loadBalance).Find(loadBalance).Error
}

func (loadBalance *LoadBalance) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c).Save(loadBalance).Error
}

func (loadBalance *LoadBalance) GetIPListByModel() []string {
	return strings.Split(loadBalance.IpList, ",")
}

func (loadBalance *LoadBalance) GetWeightListByModel() []string {
	return strings.Split(loadBalance.WeightList, ",")
}

var LoadBalancerHandler *LoadBalancer

// 存储slice中的服务负载均衡器对象serviceName->LoadBalance
type LoadBalancerItem struct {
	ServiceName string
	LoadBalance load_balance.LoadBalance
}

// 存储所有服务的负载均衡器 一个服务对应使用一个负载均衡器serviceName->LoadBalance
type LoadBalancer struct {
	LoadBalanceMap   map[string]*LoadBalancerItem
	LoadBalanceSlice []*LoadBalancerItem
	Locker           sync.RWMutex
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		LoadBalanceMap:   map[string]*LoadBalancerItem{},
		LoadBalanceSlice: []*LoadBalancerItem{},
		Locker:           sync.RWMutex{},
	}
}

func init() {
	//启动时初始化LoadBalancerHandler
	LoadBalancerHandler = NewLoadBalancer()
}

// 根据serviceDetail获取服务对应的负载均衡器
func (l *LoadBalancer) GetLoadBalance(service *ServiceDetail) (load_balance.LoadBalance, error) {
	//step1:查询LoadBalanceSlice中是否已存在对应服务的负载均衡器
	for _, lbItem := range l.LoadBalanceSlice {
		if lbItem.ServiceName == service.Info.ServiceName {
			return lbItem.LoadBalance, nil
		}
	}

	//step2:如无则新建
	schema := "http://"
	if service.HTTPRule.NeedHttps == 1 {
		schema = "https://"
	}
	//tcp和grpc类型 schema为空
	if service.Info.LoadType == public.LoadTypeTCP || service.Info.LoadType == public.LoadTypeGRPC {
		schema = ""
	}

	//获取服务ip列表及权重列表
	ipList := service.LoadBalance.GetIPListByModel()
	weightList := service.LoadBalance.GetWeightListByModel()

	//将服务及权重进行映射并组装2
	ipConf := map[string]string{}
	for index, ip := range ipList {
		ipConf[ip] = weightList[index]
	}
	//生成服务格式化字符串format
	format := fmt.Sprintf("%s%s", schema, "%s")
	//生成负载均衡配置LoadBalanceConf：使用手动发现模式
	loadBalanceConfigCheck, err := load_balance.NewLoadBalanceConfigCheck(ipConf, format)
	if err != nil {
		return nil, err
	}
	//使用负载均衡配置生成负载均衡器
	loadBalance := load_balance.LoadBalanceFactoryWithConf(load_balance.LbType(service.LoadBalance.RoundType), loadBalanceConfigCheck)

	//step3:存入LoadBalanceMap和LoadBalanceSlice
	lbItem := &LoadBalancerItem{
		ServiceName: service.Info.ServiceName,
		LoadBalance: loadBalance,
	}
	l.LoadBalanceSlice = append(l.LoadBalanceSlice, lbItem)
	l.Locker.Lock()
	defer l.Locker.Unlock()
	l.LoadBalanceMap[service.Info.ServiceName] = lbItem
	return loadBalance, nil
}

var TransportorHandler *Transportor

// 存储slice中的服务连接池对象serviceName->LoadBalance
type TransportorItem struct {
	ServiceName string
	Trans       *http.Transport
}

// 存储所有服务的连接池 一个服务对应使用一个连接池serviceName->TransportorItem
type Transportor struct {
	TransportorMap   map[string]*TransportorItem
	TransportorSlice []*TransportorItem
	Locker           sync.RWMutex
}

func NewTransportor() *Transportor {
	return &Transportor{
		TransportorMap:   map[string]*TransportorItem{},
		TransportorSlice: []*TransportorItem{},
		Locker:           sync.RWMutex{},
	}
}

func init() {
	//启动时初始化TransportorHandler
	TransportorHandler = NewTransportor()
}

// 根据serviceDetail获取服务对应的连接池
func (t *Transportor) GetTrans(service *ServiceDetail) (*http.Transport, error) {
	//step1:查询TransportorSlice中是否已存在对应服务的负载均衡器
	for _, tsItem := range t.TransportorSlice {
		if tsItem.ServiceName == service.Info.ServiceName {
			return tsItem.Trans, nil
		}
	}

	//step2:如无则新建
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(service.LoadBalance.UpstreamConnectTimeout) * time.Second, //连接超时
			KeepAlive: 30 * time.Second,                                                        //长连接超时时间
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          service.LoadBalance.UpstreamMaxIdle,                                    //最大空闲连接
		IdleConnTimeout:       time.Duration(service.LoadBalance.UpstreamIdleTimeout) * time.Second,   //空闲超时时间
		TLSHandshakeTimeout:   10 * time.Second,                                                       //tls握手超时时间
		ResponseHeaderTimeout: time.Duration(service.LoadBalance.UpstreamHeaderTimeout) * time.Second, //100-continue超时时间
	}

	//step3:存入TransportorMap和TransportorSlice
	tsItem := &TransportorItem{
		ServiceName: service.Info.ServiceName,
		Trans:       transport,
	}
	t.TransportorSlice = append(t.TransportorSlice, tsItem)
	t.Locker.Lock()
	defer t.Locker.Unlock()
	t.TransportorMap[service.Info.ServiceName] = tsItem
	return transport, nil
}
