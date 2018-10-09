package redis

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/chasex/redis-go-cluster"

	"github.com/lxt1045/kits/log"
)

const (
	connDefTimeOutMs = 500  // ms
	readTimeOutMs    = 1000 // ms
	writeTimeOutMs   = 1000 // ms
	keepAliveNumMs   = 16
	aliveTimeMs      = 60

	idRange = 2000
)

var (
	errorNotFound = fmt.Errorf("not found")

	slowLog = true
)

//Config Redis的配置信息
type Config struct {
	Server       []string
	ConnTimeOut  int // 连接超时时间 单位秒
	ReadTimeOut  int
	WriteTimeOut int
	KeepAlive    int
	AliveTime    int
}

//Pool Redis的连接池封装
type Pool struct {
	*redis.Cluster
}

//New 根据配置创建连接池
func New(cfg *Config) (*Pool, error) {
	if nil == cfg {
		return nil, fmt.Errorf("input cfg is nil")
	}
	if 0 == len(cfg.Server) {
		return nil, fmt.Errorf("redis server not configed")
	}
	cfg.ConnTimeOut = CondV(cfg.ConnTimeOut == 0, connDefTimeOutMs, cfg.ConnTimeOut).(int)
	cfg.ReadTimeOut = CondV(cfg.ReadTimeOut == 0, readTimeOutMs, cfg.ReadTimeOut).(int)
	cfg.WriteTimeOut = CondV(cfg.WriteTimeOut == 0, writeTimeOutMs, cfg.WriteTimeOut).(int)
	cfg.KeepAlive = CondV(cfg.KeepAlive == 0, keepAliveNumMs, cfg.KeepAlive).(int)
	cfg.AliveTime = CondV(cfg.AliveTime == 0, aliveTimeMs, cfg.AliveTime).(int)

	log.Printf("Init.redis.config: %+v\n", cfg)

	cluster, err := redis.NewCluster(
		&redis.Options{
			StartNodes:   cfg.Server,
			ConnTimeout:  time.Duration(cfg.ConnTimeOut) * time.Millisecond,
			ReadTimeout:  time.Duration(cfg.ReadTimeOut) * time.Millisecond,
			WriteTimeout: time.Duration(cfg.WriteTimeOut) * time.Millisecond,
			KeepAlive:    cfg.KeepAlive,
			AliveTime:    time.Duration(cfg.AliveTime) * time.Second,
		})
	if nil != err {
		return nil, fmt.Errorf("redis.NewCluster.%v", err)
	}
	return &Pool{Cluster: cluster}, nil
}

// // 所有未定义的接口使用
// func (p *Pool) Do(command string, args ...interface{}) (reply interface{}, err error) {
// 	return p.Cluster.Do(command, args...)
// }

func (p *Pool) Get(key string) (interface{}, error) {
	return p.Cluster.Do("GET", key)
}

// 如果redis不支持MGET，则需要修改为一个一个的查询
func (p *Pool) MGet(keys []interface{}) (interface{}, error) {
	return p.Cluster.Do("MGET", keys...)
}

func (p *Pool) HMGet(keys []interface{}) (interface{}, error) {
	return p.Cluster.Do("HMGET", keys...)
}
func (p *Pool) HGet(keys []interface{}) (interface{}, error) {
	return p.Cluster.Do("HGET", keys...)
}

// 获取Json对象的数据
func (p *Pool) GetJsonObj(key string, data interface{}) error {
	buf, err := p.Cluster.Do("GET", key)
	if nil != err {
		return err
	}

	if nil == buf {
		return errorNotFound
	}

	if err = json.Unmarshal(buf.([]byte), data); nil != err {
		return fmt.Errorf("json.Unmarshal err: %v", err)
	}

	return nil
}

// timeout 超时时间，单位秒
func (p *Pool) Set(key string, val interface{}, timeout int) error {
	var err error
	if timeout <= 0 {
		_, err = p.Cluster.Do("SET", key, val)
	} else {
		_, err = p.Cluster.Do("SETEX", key, timeout, val)
	}

	return err
}
func (p *Pool) HMSet(keys []interface{}) error {
	_, err := p.Cluster.Do("HMSET", keys...)
	return err
}
func (p *Pool) HSet(keys []interface{}) error {
	_, err := p.Cluster.Do("HSET", keys...)
	return err
}

func (p *Pool) Del(key string) error {
	_, err := p.Cluster.Do("DEL", key)
	if err != nil {
		return err
	}

	return nil
}

func (p *Pool) IncrBy(key string, step int) (interface{}, error) {
	return p.Cluster.Do("INCRBY", key, step)
}

func (p *Pool) DecrBy(key string, step int) (interface{}, error) {
	return p.Cluster.Do("DECRBY", key, step)
}

func (p *Pool) Expire(key string, timeout int) error {
	_, err := p.Cluster.Do("EXPIRE", key, timeout)
	if err != nil {
		return err
	}

	return nil
}

//IDInc 为减少访问Redis次数，在此做缓存
func (p *Pool) IDInc(key string) (int64, error) {
	I, ok := IDsMap.Load(key)
	if ok {
		if p, ok := I.(*ID); ok {
			return atomic.AddInt64(&p.current, 1), nil
		}
	}

	max, err := p.IDIncDirect(key, idRange)
	if err != nil {
		return time.Now().UnixNano(), err
	}

	id := &ID{
		current: max - idRange + 1,
		max:     max,
	}
	I, loaded := IDsMap.LoadOrStore(key, id)
	if loaded {
		if id, ok := I.(*ID); ok {
			return atomic.AddInt64(&id.current, 1), nil
		}
	}

	return max - idRange, nil
}

//IDIncDirect 直接从Redis获取，不缓存
func (p *Pool) IDIncDirect(key string, n int) (int64, error) {
	resp, err := p.Do("INCRBY", key, n)
	if err != nil {
		return time.Now().UnixNano(), fmt.Errorf("redis.IDInc(%s) error:%v", key, err)
	}
	id, ok := resp.(int64)
	if !ok {
		return time.Now().UnixNano(), fmt.Errorf("redis.IDInc(%s) error, get:%v", key, resp)
	}

	return id, nil
}
