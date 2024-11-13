package load_balance

import (
	"fmt"
	"github.com/starMoonZhao/go_gateway/reverse_proxy/zookeeper"
	"log"
)

// 负载均衡可用服务配置：用于zk服务注册查看服务活性
// 实现LoadBalance、Observer接口
type LoadBalanceConfigZk struct {
	observers    []Observer
	path         string   //服务在zk下的注册地址
	zkHosts      []string //zk服务的ip地址
	confIPWeight map[string]string
	activeList   []string
	format       string
}

// 向负载均衡配置中注册观察者对象
func (l *LoadBalanceConfigZk) Attach(o Observer) {
	l.observers = append(l.observers, o)
}

// 返回可用服务列表
func (l *LoadBalanceConfigZk) GetConf() []string {
	confList := make([]string, len(l.activeList))
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
func (l *LoadBalanceConfigZk) WatchConf() {
	//连接zk服务
	zkManager := zookeeper.NewZkManager(l.zkHosts)
	zkManager.Connect()
	//设置zkManager监听的地址
	chanList, chanErr := zkManager.WatchServerListByPath(l.path)
	//使用协程不间断的查询服务可用性
	go func() {
		defer zkManager.Close()
		for {
			//读取通道中的更新列表或错误信息
			select {
			case err := <-chanErr:
				log.Printf("zk change err:%v\n", err)
			case changeList := <-chanList:
				l.UpdateConf(changeList)
				log.Printf("zk change list:%v\n", changeList)
			}
		}
	}()
}

// 更新配置列表
func (l *LoadBalanceConfigZk) UpdateConf(conf []string) {
	l.activeList = conf
	for _, obverse := range l.observers {
		//同时通知观察者更新服务
		obverse.Update()
	}
}

// 默认构造器
func NewLoadBalanceConfigZk(format, path string, zkHosts []string, conf map[string]string) (*LoadBalanceConfigZk, error) {
	//初次加载可用服务列表
	zkManager := zookeeper.NewZkManager(zkHosts)
	zkManager.Connect()
	defer zkManager.Close()
	activeList, err := zkManager.GetServerListByPath(path)
	if err != nil {
		return nil, err
	}
	loadBalanceConfigZk := &LoadBalanceConfigZk{
		format:       format,
		path:         path,
		zkHosts:      zkHosts,
		confIPWeight: conf,
		activeList:   activeList,
	}
	//开启负载均衡配置的服务探活
	loadBalanceConfigZk.WatchConf()
	return loadBalanceConfigZk, nil
}
