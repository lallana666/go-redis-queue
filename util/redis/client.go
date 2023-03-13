package redis

import (
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v7"
)

/**
redis client工具类
*/

// redis连接池
var gClient *redis.Client

// redis 基础数据类型
// set string expire
func Set(key, value string, time int) error {
	res := gClient.Do("SET", key, value)
	if res.Err() != nil {
		log.Println(res.Err())
		return res.Err()
	}

	res = gClient.Do("expire", key, time)
	if res.Err() != nil {
		log.Println(res.Err())
		return res.Err()
	}
	return nil
}

// set string
func SetString(key, value string) error {
	res := gClient.Do("SET", key, value)
	if res.Err() != nil {
		log.Println(res.Err())
		return res.Err()
	}
	return nil
}

// get string
func GetString(key string) (string, error) {
	res := gClient.Do("GET", key)
	if res.Err() != nil {
		log.Println(res.Err())
		return "", res.Err()
	}
	return res.String(), nil
}

func Exist(key string) (bool, error) {
	res := gClient.Do("EXISTS", key)
	if res.Err() != nil {
		log.Println(res.Err())
		return false, res.Err()
	}

	value, err := res.Bool()
	if err != nil {
		log.Println(err)
		return false, err
	}
	return value, nil
}

// delete key
func Delete(key string) error {
	res := gClient.Do("DEL", key)
	if res.Err() != nil {
		log.Println(res.Err())
		return res.Err()
	}

	_, err := res.Int64()
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// set out time
func SetExpire(key string, time int) error {
	res := gClient.Do("expire", key, time)
	if res.Err() != nil {
		log.Println(res.Err())
		return res.Err()
	}
	return nil
}

// hash
func HashSet(keyValue ...string) error {
	res := gClient.Do("HSET", keyValue[0], keyValue[1], keyValue[2])
	if res.Err() != nil {
		log.Println(res.Err())
		return res.Err()
	}
	return nil
}

func HashMSet(key string, keyValue ...string) error {
	res := gClient.HMSet(key, keyValue)
	if res.Err() != nil {
		log.Println(res.Err())
		return res.Err()
	}
	return nil
}

func HashGet(keyValue ...string) (string, error) {
	res := gClient.Do("HGET", keyValue[0], keyValue[1])
	if res.Err() != nil {
		log.Println(res.Err())
		return "", res.Err()
	}

	value, err := res.Text()
	if err != nil {
		log.Println(err)
		return "", err
	}
	return value, nil
}

func HGetAll(key string) (map[string]string, error) {
	res := gClient.HGetAll(key)
	if res.Err() != nil {
		log.Println(res.Err())
		return nil, res.Err()
	}

	return res.Val(), nil
}

func LIndex(key string, index int64) (string, error) {
	stringCmd := gClient.LIndex(key, index)
	if stringCmd.Err() != nil {
		log.Println(stringCmd.Err())
		return "", stringCmd.Err()
	}

	return stringCmd.Val(), nil
}

func Brpop(queueName string, timeout int) (data string, err error) {
	res := gClient.Do("brpop", queueName, timeout)
	if res.Err() != nil {
		log.Println(res.Err())
		return "", res.Err()
	}
	return data, nil
}

func Lpush(keyValue ...string) (err error) {
	res := gClient.Do("lpush", keyValue[0], keyValue[1])
	if res.Err() != nil {
		log.Println(res.Err())
		return res.Err()
	}
	return
}

func Llen(key string) (len int, err error) {
	res := gClient.Do("llen", key)
	if res.Err() != nil {
		log.Println(res.Err())
		return 0, res.Err()
	}
	return res.Int()
}

func Hincrby(hash string, field string, incr int) (err error) {
	res := gClient.Do("HINCRBY", hash, field, incr)
	if res.Err() != nil {
		log.Println(res.Err())
		return res.Err()
	}
	return
}

func Brpoplpush(source string, destination string, timeout time.Duration) (data string, err error) {
	res := gClient.BRPopLPush(source, destination, timeout)
	if res.Err() != nil {
		log.Println(res.Err())
		return "", res.Err()
	}
	return res.Val(), nil
}

// redis事务
func Multi() (pipeliner redis.Pipeliner) {
	pipe := gClient.TxPipeline()
	return pipe
}

// redis连接池
func Init(address string) error {
	gClient = redis.NewClient(&redis.Options{
		//连接信息
		Network: "tcp",   //网络类型，tcp or unix，默认tcp
		Addr:    address, //主机名+冒号+端口，默认localhost:6379
		DB:      0,       // redis数据库index

		//连接池容量及闲置连接数量
		PoolSize:     15, // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
		MinIdleConns: 10, //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；。

		//超时
		DialTimeout:  5 * time.Second, //连接建立超时时间，默认5秒。
		ReadTimeout:  3 * time.Second, //读超时，默认3秒， -1表示取消读超时
		WriteTimeout: 3 * time.Second, //写超时，默认等于读超时
		PoolTimeout:  4 * time.Second, //当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒。

		//闲置连接检查包括IdleTimeout，MaxConnAge
		IdleCheckFrequency: 60 * time.Second, //闲置连接检查的周期，默认为1分钟，-1表示不做周期性检查，只在客户端获取连接时对闲置连接进行处理。
		IdleTimeout:        5 * time.Minute,  //闲置超时，默认5分钟，-1表示取消闲置超时检查
		MaxConnAge:         0 * time.Second,  //连接存活时长，从创建开始计时，超过指定时长则关闭连接，默认为0，即不关闭存活时长较长的连接

		//命令执行失败时的重试策略
		MaxRetries:      0,                      // 命令执行失败时，最多重试多少次，默认为0即不重试
		MinRetryBackoff: 8 * time.Millisecond,   //每次计算重试间隔时间的下限，默认8毫秒，-1表示取消间隔
		MaxRetryBackoff: 512 * time.Millisecond, //每次计算重试间隔时间的上限，默认512毫秒，-1表示取消间隔

		//钩子函数
		OnConnect: func(conn *redis.Conn) error { //仅当客户端执行命令时需要从连接池获取连接时，如果连接池需要新建连接时则会调用此钩子函数
			fmt.Printf("conn=%v\n", conn)
			return nil
		},
	})
	return nil
}
