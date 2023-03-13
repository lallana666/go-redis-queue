package client

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
)

var (
	ErrNoAvailableEndpoints = errors.New("heque_redis_client: no available endpoints")
	ErrNoAvailableKey       = errors.New("heque_redis_client: no available key")
)

const (
	hequeKeyJobs    = "registry:jobs:"
	hequeKeyPending = "registry:pending:"
	hequeKeyRunning = "registry:running:"
	hequeKeyBatches = "registry:batches:"
)

type Client struct {
	redis   *redis.Client
	keyFunc func(key string, name string) (string, error)
}

func New(cfg Config) (*Client, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, ErrNoAvailableEndpoints
	}

	// redis := newRedisClusterClient(cfg.Endpoints)
	redisClient := newRedisClient(cfg.Endpoints)

	return &Client{
		redis:   redisClient,
		keyFunc: DefaultKeyFunc,
	}, nil
}

func DefaultKeyFunc(key string, name string) (string, error) {
	if len(key) == 0 || len(name) == 0 {
		return "", ErrNoAvailableKey
	}
	return key + name, nil
}

func newRedisClusterClient(endpoints []string) *redis.ClusterClient {
	gClient := redis.NewClusterClient(&redis.ClusterOptions{
		//连接信息
		Addrs:         endpoints, //主机名+冒号+端口，默认localhost:6379
		ReadOnly:      true,
		RouteRandomly: true,
	})
	return gClient
}

func newRedisClient(endpoints []string) *redis.Client {
	gClient := redis.NewClient(&redis.Options{
		//连接信息
		Network: "tcp",        //网络类型，tcp or unix，默认tcp
		Addr:    endpoints[0], //主机名+冒号+端口，默认localhost:6379
		DB:      1,            // redis数据库index

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
	})
	return gClient
}

// Enqueue
func (c *Client) Enqueue(spec JobSpec) (*Job, error) {
	// 生成job id
	jobID := uuid.New().String()

	// ****************
	// 开始redis事务
	// ****************
	pipe := c.redis.TxPipeline()
	// job key value
	jobKey, err := c.keyFunc(hequeKeyJobs, jobID)
	if err != nil {
		pipe.Discard()
		return nil, err
	}

	res := pipe.Set(jobKey, spec.Payload, 0*time.Second)
	if res.Err() != nil {
		pipe.Discard()
		return nil, res.Err()
	}

	// pending queue key, push to pending
	queueKey, err := c.keyFunc(hequeKeyPending, spec.QueueName)
	if err != nil {
		pipe.Discard()
		return nil, err
	}

	lres := pipe.LPush(queueKey, jobID)
	if lres.Err() != nil {
		pipe.Discard()
		return nil, lres.Err()
	}

	// batch count 批次统计
	batchKey, err := c.keyFunc(hequeKeyBatches, spec.Batch)
	if err != nil {
		pipe.Discard()
		return nil, err
	}
	hres := pipe.HIncrBy(batchKey, "pending", 1)
	if hres.Err() != nil {
		pipe.Discard()
		return nil, hres.Err()
	}

	// ****************
	// redis事务结束
	// ****************
	_, err = pipe.Exec()
	if err != nil {
		pipe.Discard()
		return nil, err
	}

	var job = &Job{
		ID:   jobID,
		Spec: spec,
		Status: JobStatus{
			Phase: JobPending,
		},
	}

	return job, nil
}

// Dequeue
func (c *Client) Dequeue(queueName string) (*Job, error) {
	// 判断是否有pending jobs
	// 如果pending没有，阻塞
	pendingKey, err := c.keyFunc(hequeKeyPending, queueName)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	runningKey, err := c.keyFunc(hequeKeyRunning, queueName)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	pendingJobString := c.redis.BRPopLPush(pendingKey, runningKey, 0*time.Second)
	if pendingJobString.Err() != nil {
		log.Println(pendingJobString.Err())
		return nil, pendingJobString.Err()
	}

	pendingJobId := pendingJobString.Val()

	jobKey, err := c.keyFunc(hequeKeyJobs, pendingJobId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	jobString := c.redis.Get(jobKey)
	if jobString.Err() != nil {
		log.Println(jobString.Err())
		return nil, jobString.Err()
	}

	// 获取job里的估值参数
	jobStringMap := make(map[string]string)
	err = json.Unmarshal([]byte(jobString.Val()), &jobStringMap)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var job = &Job{
		ID: pendingJobId,
		Spec: JobSpec{
			Payload:   jobString.Val(),
			QueueName: queueName,
			Batch:     jobStringMap["Batch"],
		},
		Status: JobStatus{
			Phase: JobRunning,
		},
	}

	// ****************
	// redis事务开始
	// ****************
	pl := c.redis.TxPipeline()
	if job.Spec.Batch != "" {
		batchKey, err := c.keyFunc(hequeKeyBatches, job.Spec.Batch)
		if err != nil {
			pl.Discard()
			return nil, err
		}

		intCmd := pl.HIncrBy(batchKey, "running", 1)
		if intCmd.Err() != nil {
			log.Println(intCmd.Err())
			pl.Discard()
			return nil, intCmd.Err()
		}

		intCmd = pl.HIncrBy(batchKey, "pending", -1)
		if intCmd.Err() != nil {
			log.Println(intCmd.Err())
			pl.Discard()
			return nil, intCmd.Err()
		}
	}

	// ****************
	// redis事务结束
	// ****************
	if _, err := pl.Exec(); err != nil {
		log.Println(err)
		pl.Discard()
		return nil, err
	}
	return job, nil
}

// MarkAsDone
func (c *Client) MarkAsDone(job *Job) error {
	// ****************
	// redis事务开始
	// ****************
	pl := c.redis.TxPipeline()
	batchKey, err := c.keyFunc(hequeKeyBatches, job.Spec.Batch)
	if err != nil {
		log.Println(err)
		pl.Discard()
		return err
	}

	intCmd := pl.HIncrBy(batchKey, "running", -1)
	if intCmd.Err() != nil {
		log.Println(intCmd.Err())
		pl.Discard()
		return intCmd.Err()
	}

	intCmd = pl.HIncrBy(batchKey, "done", 1)
	if intCmd.Err() != nil {
		log.Println(intCmd.Err())
		pl.Discard()
		return intCmd.Err()
	}

	jobKey, err := c.keyFunc(hequeKeyJobs, job.ID)
	if err != nil {
		log.Println(err)
		return err
	}

	intCmd = pl.Del(jobKey)
	if intCmd.Err() != nil {
		log.Println(intCmd.Err())
		pl.Discard()
		return intCmd.Err()
	}

	//弹出running
	runningKey, err := c.keyFunc(hequeKeyRunning, job.Spec.QueueName)
	if err != nil {
		log.Println(err)
		return err
	}

	stringCmd := pl.RPop(runningKey)
	if stringCmd.Err() != nil {
		log.Println(stringCmd.Err())
		pl.Discard()
		return stringCmd.Err()
	}

	// ****************
	// redis事务结束
	// ****************
	if _, err := pl.Exec(); err != nil {
		log.Println(err)
		pl.Discard()
		return err
	}
	return nil
}

// MarkAsFailed
func (c *Client) MarkAsFailed(job *Job) error {
	// ****************
	// redis事务开始
	// ****************
	pl := c.redis.TxPipeline()
	batchKey, err := c.keyFunc(hequeKeyBatches, job.Spec.Batch)
	if err != nil {
		log.Println(err)
		pl.Discard()
		return err
	}

	intCmd := pl.HIncrBy(batchKey, "running", -1)
	if intCmd.Err() != nil {
		log.Println(intCmd.Err())
		pl.Discard()
		return intCmd.Err()
	}

	intCmd = pl.HIncrBy(batchKey, "failed", 1)
	if intCmd.Err() != nil {
		log.Println(intCmd.Err())
		pl.Discard()
		return intCmd.Err()
	}

	jobKey, err := c.keyFunc(hequeKeyJobs, job.ID)
	if err != nil {
		log.Println(err)
		return err
	}

	intCmd = pl.Del(jobKey)
	if intCmd.Err() != nil {
		log.Println(intCmd.Err())
		pl.Discard()
		return intCmd.Err()
	}

	//弹出running
	runningKey, err := c.keyFunc(hequeKeyRunning, job.Spec.QueueName)
	if err != nil {
		log.Println(err)
		return err
	}

	stringCmd := pl.RPop(runningKey)
	if stringCmd.Err() != nil {
		log.Println(stringCmd.Err())
		pl.Discard()
		return stringCmd.Err()
	}

	// ****************
	// redis事务结束
	// ****************
	if _, err := pl.Exec(); err != nil {
		log.Println(err)
		_ = pl.Discard()
		return err
	}
	log.Println("房屋估值失败......jobId:" + job.ID)
	return nil
}

// progress
func (c *Client) Progress(batch string) (float64, error) {
	var progress float64

	batchKey, err := c.keyFunc(hequeKeyBatches, batch)
	if err != nil {
		return 0, err
	}

	jobStringMap := c.redis.HGetAll(batchKey)
	if jobStringMap.Err() != nil {
		return 0, jobStringMap.Err()
	}

	var j *BatchCount
	jsonBytes, err := json.Marshal(jobStringMap.Val())
	if err != nil {
		return 0, err
	}

	err = json.Unmarshal(jsonBytes, &j)
	if err != nil {
		return 0, err
	}

	var done = 0
	var pending = 0
	var running = 0
	var failed = 0

	if j.Done != nil {
		done, err = strconv.Atoi(*j.Done)
		if err != nil {
			return 0, err
		}
	}

	if j.Pending != nil {
		pending, err = strconv.Atoi(*j.Pending)
		if err != nil {
			return 0, err
		}
	}

	if j.Running != nil {
		running, err = strconv.Atoi(*j.Running)
		if err != nil {
			return 0, err
		}
	}

	if j.Failed != nil {
		failed, err = strconv.Atoi(*j.Failed)
		if err != nil {
			return 0, err
		}
	}

	// 计算进度
	total := done + pending + running + failed
	if total == 0 {
		progress = 1
	} else {
		progress = float64(done+failed) / float64(total)
	}

	return progress, nil
}
