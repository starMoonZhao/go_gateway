package load_balance

import (
	"github.com/pkg/errors"
	"log"
	"math/rand"
	"strings"
)

// 随机策略的负载均衡器
type RandomBalance struct {
	conf     LoadBalanceConf //被观察主体
	curIndex int             //当前使用的服务下标
	css      []string        //当前负载均衡器可用的服务列表
}

func (r *RandomBalance) Add(params ...string) error {
	if len(params) == 0 {
		return errors.New("params is empty")
	}
	r.css = append(r.css, params[0])
	return nil
}

func (r *RandomBalance) Get(key string) (string, error) {
	if len(r.css) == 0 {
		return "", errors.New("css is empty")
	}
	//使用随机数获取可用服务下标
	r.curIndex = rand.Intn(len(r.css))
	return r.css[r.curIndex], nil
}

func (r *RandomBalance) SetConf(conf LoadBalanceConf) {
	r.conf = conf
}

// 本负载均衡器作为观察者注入到了LoadBalanceConf对象中
// 当LoadBalanceConf发生变化时，会调用所有观察者的所有Update方法同步变化
// 负载均衡器将变化同步到服务列表中
func (r *RandomBalance) Update() {
	log.Printf("Update get conf:%v\n", r.conf)
	r.css = []string{}
	for _, ip := range r.conf.GetConf() {
		r.Add(strings.Split(ip, ",")...)
	}
}
