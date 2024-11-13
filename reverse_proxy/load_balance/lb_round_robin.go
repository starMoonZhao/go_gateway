package load_balance

import (
	"github.com/pkg/errors"
	"log"
	"strings"
)

// 轮询策略的负载均衡器
type RoundRobinBalance struct {
	conf     LoadBalanceConf //被观察主体
	curIndex int             //当前使用的服务下标
	css      []string        //当前负载均衡器可用的服务列表
}

func (r *RoundRobinBalance) Add(params ...string) error {
	if len(params) == 0 {
		return errors.New("params is empty")
	}
	r.css = append(r.css, params[0])
	return nil
}

func (r *RoundRobinBalance) Get(key string) (string, error) {
	if len(r.css) == 0 {
		return "", errors.New("css is empty")
	}
	if r.curIndex >= len(r.css) {
		r.curIndex = 0
	}
	curAddr := r.css[r.curIndex]
	r.curIndex = (r.curIndex + 1) % len(r.css)
	return curAddr, nil
}

func (r *RoundRobinBalance) SetConf(conf LoadBalanceConf) {
	r.conf = conf
}

func (r *RoundRobinBalance) Update() {
	log.Printf("Update get conf:%v\n", r.conf)
	r.css = []string{}
	for _, ip := range r.conf.GetConf() {
		r.Add(strings.Split(ip, ",")...)
	}
}
