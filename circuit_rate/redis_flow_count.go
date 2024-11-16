package circuit_rate

import (
	"fmt"
	"github.com/e421083458/golang_common/lib"
	"github.com/garyburd/redigo/redis"
	"github.com/starMoonZhao/go_gateway/public"
	"log"
	"sync/atomic"
	"time"
)

// 标识一个服务或租户或系统的流量统计对象
type RedisFlowCount struct {
	ID          string        //标识
	Interval    time.Duration //更新间隔
	QPS         int64
	Unix        int64 //UNIX 时间戳（从 1970 年 1 月 1 日 00:00:00 UTC 到当前时间的秒数）
	TickerCount int64 //间隔时间内产生的访问量
	TotalCount  int64 //总访问量
}

// 为统计对象新建一个流量统计任务
func NewRedisFlowCount(id string, Interval time.Duration) *RedisFlowCount {
	flowCount := &RedisFlowCount{
		ID:       id,
		Interval: Interval,
		QPS:      0,
		Unix:     0,
	}

	//建立协程为该对象建立统计任务
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println("redis flow count err:", err)
			}
		}()

		//创建一个定时器，每隔指定的时间间隔（d）向其 C 通道发送当前时间
		ticker := time.NewTicker(flowCount.Interval)
		//for循环读取C 通道的输出
		for {
			<-ticker.C
			//获取当前统计对象间隔时间内产生的流量
			tickerCount := atomic.LoadInt64(&flowCount.TickerCount)
			//重置间隔时间内的流量奇数
			atomic.StoreInt64(&flowCount.TickerCount, 0)

			now := time.Now()
			//构造此统计对象在redis中存储统计对象的key键
			dayKey := flowCount.GetDayKey(now)
			hourKey := flowCount.GetHourKey(now)
			//将时间间隔内的ticker写入当前统计对象对应的dayKey和hourKey中
			if err := RedisConfPipeline(func(c redis.Conn) {
				//数值递增
				c.Send("INCRBY", dayKey, tickerCount)
				c.Send("INCRBY", hourKey, tickerCount)
				//设置过期时间：两天
				c.Send("EXPIRE", dayKey, 60*60*24*2)
				c.Send("EXPIRE", hourKey, 60*60*24*2)
			}); err != nil {
				log.Println("RedisConfPipeline err:", err)
				continue
			}

			//查询该统计对象实际的TotalCount并存入RedisFlowCount
			totalCount, err := flowCount.GetDayData(now)
			if err != nil {
				log.Println("GetDayData err:", err)
				continue
			}

			//计算实际的tickerCount why：flowCount中的ticker只统计经过此代理的访问 然后再将其汇总到redis中的dayData上
			unix := time.Now().Unix()
			tickerCount = totalCount - flowCount.TotalCount

			//写回
			flowCount.TotalCount = totalCount
			flowCount.QPS = tickerCount / (unix - flowCount.Unix)
			flowCount.Unix = unix
		}
	}()
	return flowCount
}

// 根据时间构造当前流量统计对象存储的日数据的key
func (c *RedisFlowCount) GetDayKey(t time.Time) string {
	dayStr := t.In(lib.TimeLocation).Format("20060102")
	return fmt.Sprintf("%s_%s_%s", public.RedisFlowDayKey, dayStr, c.ID)
}

// 根据时间构造当前流量统计对象存储的小时数据的key
func (c *RedisFlowCount) GetHourKey(t time.Time) string {
	hourStr := t.In(lib.TimeLocation).Format("2006010215")
	return fmt.Sprintf("%s_%s_%s", public.RedisFlowHourKey, hourStr, c.ID)
}

// 根据key查询当前流量统计对象存储的日数据的data
func (c *RedisFlowCount) GetDayData(t time.Time) (int64, error) {
	return redis.Int64(RedisConfDo("GET", c.GetDayKey(t)))
}

// 根据key查询当前流量统计对象存储的小时数据的data
func (c *RedisFlowCount) GetHourData(t time.Time) (int64, error) {
	return redis.Int64(RedisConfDo("GET", c.GetHourKey(t)))
}

// 当发生访问时 增加RedisFlowCount中的TickerCount
func (r *RedisFlowCount) Increase() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Increase err:", err)
			}
		}()
		atomic.AddInt64(&r.TickerCount, 1)
	}()
}
