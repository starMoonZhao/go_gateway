package load_balance

// 定义负载均衡器的策略 作为负载均衡器的工厂类生成对应策略的负载均衡器
type LbType int

const (
	LbRandom LbType = iota
	LbRoundRobin
	LbWeightRoundRobin
	LbConsistentHash
)

// 获取指定策略的负载均衡器
func LoadBalanceFactory(lbType LbType) LoadBalance {
	switch lbType {
	case LbRandom:
		return &RandomBalance{}
	case LbRoundRobin:
		return &RoundRobinBalance{}
	case LbWeightRoundRobin:
		return &WeightRoundRobinBalance{}
	case LbConsistentHash:
		return &ConsistentHashBalance{}
	default:
		return &RandomBalance{}
	}
}

// 获取指定策略的负载均衡器 同时根据传入配置初始化负载均衡器
func LoadBalanceFactoryWithConf(lbType LbType, conf LoadBalanceConf) LoadBalance {
	switch lbType {
	case LbRandom:
		lb := &RandomBalance{}
		//将负载均衡配置设置到负载均衡器上
		lb.SetConf(conf)
		//将负载均衡器作为观察者注入到负载均衡配置中
		conf.Attach(lb)
		//首次手动更新负载均衡器配置
		lb.Update()
		return lb
	case LbRoundRobin:
		lb := &RoundRobinBalance{}
		lb.SetConf(conf)
		conf.Attach(lb)
		lb.Update()
		return lb
	case LbWeightRoundRobin:
		lb := &WeightRoundRobinBalance{}
		lb.SetConf(conf)
		conf.Attach(lb)
		lb.Update()
		return lb
	case LbConsistentHash:
		lb := NewConsistentHashBalance(10, nil)
		lb.SetConf(conf)
		conf.Attach(lb)
		lb.Update()
		return lb
	default:
		lb := &RoundRobinBalance{}
		lb.SetConf(conf)
		conf.Attach(lb)
		lb.Update()
		return lb
	}
}
