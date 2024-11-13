package zookeeper

import (
	"github.com/samuel/go-zookeeper/zk"
	"log"
	"time"
)

// zk服务管理器
type ZkManager struct {
	hosts []string
	conn  *zk.Conn
}

func NewZkManager(hosts []string) *ZkManager {
	return &ZkManager{hosts: hosts}
}

// 连接zk服务器
func (z *ZkManager) Connect() error {
	conn, _, err := zk.Connect(z.hosts, time.Second*5)
	if err != nil {
		return err
	}
	z.conn = conn
	return nil
}

// 断开zk服务
func (z *ZkManager) Close() {
	z.conn.Close()
	return
}

// 获取zk指定路径下的服务节点
func (z *ZkManager) GetServerListByPath(path string) (list []string, err error) {
	list, _, err = z.conn.Children(path)
	return
}

// zk watch机制：当有服务断开或者重连事件发生时，收到该事件
func (z *ZkManager) WatchServerListByPath(path string) (chan []string, chan error) {
	conn := z.conn
	//构造服务更新事件通知管道
	snapshots := make(chan []string)
	errors := make(chan error)
	go func() {
		for {
			snapshot, _, events, err := conn.ChildrenW(path)
			if err != nil {
				errors <- err
			}
			snapshots <- snapshot
			//输出事件变更
			select {
			case event := <-events:
				if event.Err != nil {
					errors <- event.Err
				}
				log.Printf("ChildrenW Event Path:%v, Type:%v\n", event.Path, event.Type)
			}
		}
	}()
	return snapshots, errors
}

// 获取服务节点配置
func (z *ZkManager) GetPathData(path string) ([]byte, *zk.Stat, error) {
	return z.conn.Get(path)
}

// 更新服务节点配置
func (z *ZkManager) SetPathData(path string, config []byte) error {
	//查看该节点是否存在
	exists, _, _ := z.conn.Exists(path)

	if !exists {
		//不存在 直接新建
		z.conn.Create(path, config, 0, zk.WorldACL(zk.PermAll))
		return nil
	}

	//已存在 更新
	_, stat, err := z.GetPathData(path)
	if err != nil {
		return err
	}
	//设置新值
	_, err = z.conn.Set(path, config, stat.Version)
	if err != nil {
		log.Printf("Update zk node path err: %v\n", err)
		return err
	}
	log.Printf("Update zk node path success: %v\n", path)
	return nil
}
