package circuit_rate

import (
	"sync"
	"time"
)

var FlowCounterHandler *FlowCounter

// 存储所有服务的连接池 一个服务对应使用一个连接池serviceName->TransportorItem
type FlowCounter struct {
	RedisFlowCountMap   map[string]*RedisFlowCount
	RedisFlowCountSlice []*RedisFlowCount
	Locker              sync.RWMutex
}

func NewFlowCounter() *FlowCounter {
	return &FlowCounter{
		RedisFlowCountMap:   map[string]*RedisFlowCount{},
		RedisFlowCountSlice: []*RedisFlowCount{},
		Locker:              sync.RWMutex{},
	}
}

func init() {
	//启动时初始化FlowCounterHandler
	FlowCounterHandler = NewFlowCounter()
}

// 根据serviceDetail获取服务对应的流量统计器
func (f *FlowCounter) GetFlowCounter(id string) (*RedisFlowCount, error) {
	//step1:查询RedisFlowCountSlice中是否已存在对应对象的流量统计器
	for _, rfcItem := range f.RedisFlowCountSlice {
		if rfcItem.ID == id {
			return rfcItem, nil
		}
	}

	//step2:如无则新建
	flowCount := NewRedisFlowCount(id, 1*time.Second)

	//step3:存入RedisFlowCountMap和RedisFlowCountSlice
	f.RedisFlowCountSlice = append(f.RedisFlowCountSlice, flowCount)
	f.Locker.Lock()
	defer f.Locker.Unlock()
	f.RedisFlowCountMap[id] = flowCount
	return flowCount, nil
}
