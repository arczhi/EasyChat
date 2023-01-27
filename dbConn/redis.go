package dbConn

import (
	"EasyChat/config"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

var conf = map[string]string{
	"type":     "tcp",
	"address":  config.Cfg.REDIS.Addr,
	"password": config.Cfg.REDIS.Password,
}

func NewRedisPool() *redis.Pool {
	return &redis.Pool{
		// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
		MaxIdle: 16, //最初的连接数量
		// MaxActive:1000000,    //最大连接数量
		MaxActive:   0,                 //连接池最大连接数量,不确定可以用0（0表示自动定义），按需分配
		IdleTimeout: 300 * time.Second, //连接关闭时间 300秒 （300秒不使用自动关闭）
		Dial: func() (redis.Conn, error) { //要连接的redis数据库

			//创建redis连接 //redis-cli -h 127.0.0.1 -p 6379 -n 1 -xxxxxx
			con, err := redis.Dial(conf["type"], conf["address"], redis.DialDatabase(1), redis.DialPassword(conf["password"]))
			if err != nil {
				log.Fatal("[ERROR] Failed to initialize the Redis database.\n", err)
				return nil, err
			}

			return con, nil
		},
	}
}

// // 从池里获取连接
// rc := RedisClient.Get()
// // 用完后将连接放回连接池
// defer rc.Close()
// // 错误判断
// if conn.Err() != nil {
//   //TODO
// }
