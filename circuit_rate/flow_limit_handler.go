package circuit_rate

import (
	"golang.org/x/time/rate"
	"sync"
)

var FlowLimiterHandler *FlowLimiter

// 存储所有服务的限流器 一个服务对应使用一个限流器serviceName->FlowLimiterItem
type FlowLimiter struct {
	FlowLimiterMap   map[string]*FlowLimiterItem
	FlowLimiterSlice []*FlowLimiterItem
	Locker           sync.RWMutex
}

type FlowLimiterItem struct {
	ID      string
	Limiter *rate.Limiter
}

func NewFlowLimiter() *FlowLimiter {
	return &FlowLimiter{
		FlowLimiterMap:   map[string]*FlowLimiterItem{},
		FlowLimiterSlice: []*FlowLimiterItem{},
		Locker:           sync.RWMutex{},
	}
}

func init() {
	//启动时初始化FlowLimiterHandler
	FlowLimiterHandler = NewFlowLimiter()
}

// 根据serviceDetail获取服务对应的限流器
func (f *FlowLimiter) GetFlowLimiter(id string, qps int) (*rate.Limiter, error) {
	//step1:查询FlowLimiterSlice中是否已存在对应对象的限流器
	for _, flItem := range f.FlowLimiterSlice {
		if flItem.ID == id {
			return flItem.Limiter, nil
		}
	}

	//step2:如无则新建
	flowLimiter := rate.NewLimiter(rate.Limit(qps), qps*3)

	//step3:存入FlowLimiterMap和FlowLimiterSlice
	flowLimiterItem := &FlowLimiterItem{ID: id, Limiter: flowLimiter}
	f.FlowLimiterSlice = append(f.FlowLimiterSlice, &FlowLimiterItem{ID: id, Limiter: flowLimiter})
	f.Locker.Lock()
	defer f.Locker.Unlock()
	f.FlowLimiterMap[id] = flowLimiterItem
	return flowLimiter, nil
}
