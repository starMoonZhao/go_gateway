package load_balance

import (
	"github.com/pkg/errors"
	"hash/crc32"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// hash函数
type Hash func(data []byte) uint32

type UInt32Slice []uint32

func (s UInt32Slice) Len() int {
	return len(s)
}

func (s UInt32Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s UInt32Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

// 哈希连续的负载均衡器
type ConsistentHashBalance struct {
	conf     LoadBalanceConf   //被观察主体
	mutex    sync.RWMutex      //读写锁
	hash     Hash              //hash函数
	replicas int               //复制因子
	keys     UInt32Slice       //已排序的节点hash切片
	hashMap  map[uint32]string //节点hash和key的map，键是hash值，值是key
}

// 哈希负载均衡器生成
func NewConsistentHashBalance(replicas int, fn Hash) *ConsistentHashBalance {
	if fn == nil {
		//键最多32位 保证是一个2^32-1的环
		fn = crc32.ChecksumIEEE
	}
	c := &ConsistentHashBalance{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[uint32]string),
	}
	return c
}

// 用来添加缓存节点，参数位节点key，比如使用ip
func (c *ConsistentHashBalance) Add(params ...string) error {
	if len(params) == 0 {
		return errors.New("params len 1 at least")
	}
	addr := params[0]
	c.mutex.Lock()
	defer c.mutex.Unlock()
	//结合复制因子计算所有虚拟节点的hash值，并存入c.keys中，同时在c.hashMap中保存hash值和key的映射
	for i := 0; i < c.replicas; i++ {
		hash := c.hash([]byte(strconv.Itoa(i) + addr))
		c.keys = append(c.keys, hash)
		c.hashMap[hash] = addr
	}

	//对所有节点的hash值进行排序 方便后续进行二分查找
	sort.Sort(c.keys)
	return nil
}

// 根据给定的对象获取最靠近它的那个节点
func (c *ConsistentHashBalance) Get(key string) (string, error) {
	if len(c.keys) == 0 {
		return "", errors.New("keys is empty")
	}
	//计算hash值
	hash := c.hash([]byte(key))
	//二分查找获取最有节点：第一个“服务器hash值”大于“数据hash值”的就是最优节点
	index := sort.Search(len(c.keys), func(i int) bool { return c.keys[i] >= hash })
	//如果查找结果大于服务节点hash数组的最大索引，表明该对象hash值位于最后一个节点之后，那么放入第一个节点中
	if index >= len(c.keys) {
		index = 0
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.hashMap[c.keys[index]], nil
}

func (c *ConsistentHashBalance) SetConf(conf LoadBalanceConf) {
	c.conf = conf
}

func (c *ConsistentHashBalance) Update() {
	log.Printf("Update get conf:%v\n", c.conf)
	c.keys = nil
	c.hashMap = make(map[uint32]string)
	for _, ip := range c.conf.GetConf() {
		c.Add(strings.Split(ip, ",")...)
	}
}
