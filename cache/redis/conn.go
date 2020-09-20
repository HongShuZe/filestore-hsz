package redis

import (
	"github.com/garyburd/redigo/redis"
	"time"
	"fmt"
)

var (
	pool *redis.Pool
	redisHost = "127.0.0.1:6379"
	//redisPass = "testupload"
)

// 创建redis连接池
func newRedisPool() *redis.Pool {
	return &redis.Pool{
		// 最大连接数
		MaxIdle: 50,
		// 再给定时间的最大连接数
		MaxActive: 30,
		// 超时时间
		IdleTimeout: 300 * time.Second,
		// 创建和确认连接
		Dial: func() (redis.Conn, error) {
			// 1.打开连接
			c, err := redis.Dial("tcp", redisHost)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			/*// 2.访问认证
			if _, err = c.Do("AUTH", redisPass); err != nil {
				fmt.Println("AUTH: redis 密码认证失败")
				c.Close()
				return nil, err
			}*/
			return c, nil
		},
		// 检查连接状态
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func init()  {
	pool = newRedisPool()
}

func RedisPool() *redis.Pool {
	return pool
}

