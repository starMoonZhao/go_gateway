package circuit_rate

import (
	"github.com/e421083458/golang_common/lib"
	"github.com/garyburd/redigo/redis"
)

// 根据redis配置获取redis管道并写入命令
func RedisConfPipeline(pip ...func(c redis.Conn)) error {
	conn, err := lib.RedisConnFactory("default")
	if err != nil {
		return err
	}
	defer conn.Close()
	//设计思想：闭包。外部传入一个以Conn为入参的函数 并实现其定义 在这里传入真正的Conn实现外部像实现的功能
	for _, fn := range pip {
		fn(conn)
	}
	conn.Flush()
	return nil
}

// 获取redis连接执行传入的命令 并返回结果
func RedisConfDo(command string, args ...interface{}) (interface{}, error) {
	conn, err := lib.RedisConnFactory("default")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.Do(command, args...)
}
