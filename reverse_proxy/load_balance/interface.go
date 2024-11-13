package load_balance

// 观察者接口:负载均衡器会实现该接口中的Update方法
// LoadBalance作为观察者嵌入LoadBalanceConf对象中，当LoadBalanceConf配置发生变化时调用Update方法使LoadBalance同步配置
type Observer interface {
	Update()
}

// 配置主体接口
type LoadBalanceConf interface {
	Attach(o Observer)   //观察者嵌入
	GetConf() []string   //读取该负载均衡配置对象的配置列表
	WatchConf()          //监听服务列表的可用情况 实时更新负载均衡配置
	UpdateConf([]string) //更新负载均衡配置
}

// 负载均衡器
type LoadBalance interface {
	Add(params ...string) error   //添加服务
	Get(string) (string, error)   //获取服务配置列表
	SetConf(conf LoadBalanceConf) //设置负载均衡配置
}
