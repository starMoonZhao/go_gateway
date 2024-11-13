package load_balance

import (
	"fmt"
	"net"
	"reflect"
	"sort"
	"time"
)

const (
	DefaultCheckTimeout   = 2
	DefaultCheckMaxErrNum = 2
	DefaultCheckInterval  = 5
)

// 负载均衡可用服务配置：用于主动探测服务活性
// 实现LoadBalance、Observer接口
type LoadBalanceConfigCheck struct {
	observers    []Observer        //观察者列表
	confIPWeight map[string]string //权重列表 原始服务列表
	activeList   []string          //活跃服务列表
	format       string            //服务格式化字符串
}

// 向负载均衡配置中注册观察者对象
func (l *LoadBalanceConfigCheck) Attach(o Observer) {
	l.observers = append(l.observers, o)
}

// 返回可用服务列表
func (l *LoadBalanceConfigCheck) GetConf() []string {
	confList := []string{}
	for _, ip := range l.activeList {
		weight, ok := l.confIPWeight[ip]
		if !ok {
			weight = "50"
		}
		confList = append(confList, fmt.Sprintf(l.format, ip)+","+weight)
	}
	return confList
}

// 监听服务可用性
func (l *LoadBalanceConfigCheck) WatchConf() {
	//使用协程不间断的查询服务可用性
	go func() {
		confIpErrNum := map[string]int{}
		for {
			//新的可用服务列表
			newActiveList := []string{}
			//遍历原始服务列表
			for item, _ := range l.confIPWeight {
				//使用tcp连接探活
				conn, err := net.DialTimeout("tcp", item, time.Duration(DefaultCheckTimeout)*time.Second)
				if err != nil {
					//探测失败 为该服务的错误次数+1
					if _, ok := confIpErrNum[item]; !ok {
						confIpErrNum[item] = 1
					} else {
						confIpErrNum[item] += 1
					}
				} else {
					conn.Close()
					//探测成功 将失败次重置
					if _, ok := confIpErrNum[item]; ok {
						confIpErrNum[item] = 0
					}
				}
				//如果错误次数小于最大错误次数 将其添加到可用服务列表中
				if confIpErrNum[item] < DefaultCheckMaxErrNum {
					newActiveList = append(newActiveList, item)
				}
			}
			//查看可用服务列表是否发生变化 如发生变化将其更新
			sort.Strings(l.activeList)
			sort.Strings(newActiveList)
			if !reflect.DeepEqual(newActiveList, l.activeList) {
				l.UpdateConf(newActiveList)
			}

			//间隔DefaultCheckInterval时间后继续探活
			time.Sleep(time.Duration(DefaultCheckInterval) * time.Second)
		}
	}()
}

// 更新配置列表
func (l *LoadBalanceConfigCheck) UpdateConf(conf []string) {
	l.activeList = conf
	for _, obverse := range l.observers {
		//同时通知观察者更新服务
		obverse.Update()
	}
}

// 默认构造器
func NewLoadBalanceConfigCheck(conf map[string]string, format string) (*LoadBalanceConfigCheck, error) {
	//将原始服务列表直接设置为可用服务列表
	activeList := []string{}
	for item, _ := range conf {
		activeList = append(activeList, item)
	}
	loadBalanceConfig := &LoadBalanceConfigCheck{
		format:       format,
		confIPWeight: conf,
		activeList:   activeList,
	}
	//开启负载均衡配置的服务探活
	loadBalanceConfig.WatchConf()
	return loadBalanceConfig, nil
}
