package load_balance

import (
	"github.com/pkg/errors"
	"log"
	"strconv"
	"strings"
)

// 权重轮询的负载均衡器
type WeightRoundRobinBalance struct {
	conf LoadBalanceConf //被观察主体
	rss  []*WeightNode   //当前负载均衡器可用的服务权重列表
}

// 权重节点
type WeightNode struct {
	addr            string //服务地址
	weight          int    //服务权重
	currentWeight   int    //服务当前权重
	effectiveWeight int    //服务有效权重 默认等于weight
}

// 手动添加可用服务
// param1:地址；param2：权重
func (w *WeightRoundRobinBalance) Add(params ...string) error {
	if len(params) != 2 {
		return errors.New("params len 2 at least.")
	}
	//权重转数字
	weight, err := strconv.ParseInt(params[1], 10, 64)
	if err != nil {
		return err
	}
	weightNode := &WeightNode{
		weight:          int(weight),
		effectiveWeight: int(weight),
		currentWeight:   0,
		addr:            params[0],
	}
	w.rss = append(w.rss, weightNode)
	return nil
}

// 设计思想：根据每个节点的权重值统计出总的权重值->获取当前权重值最大的节点->使用->减去总的权重值->每一巡每个节点加上节点自身的权重
func (w *WeightRoundRobinBalance) Get(key string) (string, error) {
	//总的权重值
	totalWeight := 0
	//当前权重最大节点
	var bestNode *WeightNode
	for _, node := range w.rss {
		//step1: 统计所有权重之和
		totalWeight += node.weight

		//step2: 变更节点临时权重为临时权重+有效权重
		node.currentWeight += node.effectiveWeight

		//step3: 有效权重默认与权重相同，通讯异常时-1, 通讯成功+1，直到恢复到weight大小
		if node.effectiveWeight < node.weight {
			node.effectiveWeight++
		}
		//step4: 找出临时权重最大的节点
		if bestNode == nil || node.currentWeight > bestNode.currentWeight {
			bestNode = node
		}
	}
	if bestNode == nil {
		return "", errors.New("no best node found")
	}
	//被选中节点减去总的权重
	bestNode.currentWeight -= totalWeight
	return bestNode.addr, nil
}

func (w *WeightRoundRobinBalance) SetConf(conf LoadBalanceConf) {
	w.conf = conf
}

func (w *WeightRoundRobinBalance) Update() {
	log.Printf("Update get conf:%v\n", w.conf)
	w.rss = nil
	for _, ip := range w.conf.GetConf() {
		w.Add(strings.Split(ip, ",")...)
	}
}
