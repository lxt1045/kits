/*
	redis相关的功能
		1. 初始化
			Init(redisCfg *RedisConfigST) error

		2. 使用接口
			Get(key string) (reply interface{}, error) 查询单个Key
			GetMulti(keys []string) (replys []interface{}, error) 查询多个KEY
			Set(key string, val interface{}, timeout int) error 设置单个Key
			Del(key string) error 删除key
			IncrBy(key string, step int) error key自增
			DecrBy(key string, step int) error key自减
			Do(command string, args ...interface{}) (reply interface{}, error) 未定义的接口使用

			// ID生成器，唯一的ID生成
			NewID(key string, base int64) (int64, error)
*/
//放全局函数
package redis

import (
	"sync"
)

var (
	defaultPool *Pool

	IDsMap sync.Map //用于缓存ID，ID即是需要全局唯一，递增的整型变量
)

func init() {
}

func Init(cfg *Config) (err error) {
	defaultPool, err = New(cfg)
	return
}

type ID struct {
	current int64
	max     int64
}

// 二元组运算的实现 max = b>a ? b : a -> max = CondV(b>a, b, a).(a.Type())
func CondV(cond bool, v1 interface{}, v2 interface{}) interface{} {
	if cond {
		return v1
	} else {
		return v2
	}
}

// 所有未定义的接口使用
func Do(command string, args ...interface{}) (reply interface{}, err error) {
	return defaultPool.Do(command, args...)
}

func Get(key string) (interface{}, error) {
	return defaultPool.Get(key)
}

// 获取Json对象的数据
func GetJsonObj(key string, data interface{}) error {
	return defaultPool.GetJsonObj(key, data)
}

// 如果redis不支持MGET，则需要修改为一个一个的查询
func MGet(keys []interface{}) (interface{}, error) {
	return defaultPool.MGet(keys)
}

// 如果redis不支持MGET，则需要修改为一个一个的查询
func HMGet(keys ...interface{}) (interface{}, error) {
	return defaultPool.HMGet(keys)
}

// 如果redis不支持MGET，则需要修改为一个一个的查询
func HGet(keys ...interface{}) (interface{}, error) {
	return defaultPool.HGet(keys)
}

// 如果redis不支持MGET，则需要修改为一个一个的查询
func HMSet(keys ...interface{}) error {
	return defaultPool.HMSet(keys)
}

// 如果redis不支持MGET，则需要修改为一个一个的查询
func HSet(keys ...interface{}) error {
	return defaultPool.HSet(keys)
}

// timeout 超时时间，单位秒
func Set(key string, val interface{}, timeout int) error {
	return defaultPool.Set(key, val, timeout)
}

func Del(key string) error {
	return defaultPool.Del(key)
}

func IncrBy(key string, step int) (interface{}, error) {
	return defaultPool.IncrBy(key, step)
}

func DecrBy(key string, step int) (interface{}, error) {
	return defaultPool.DecrBy(key, step)
}

func Expire(key string, timeout int) error {
	return defaultPool.Expire(key, timeout)
}

func IDIncU64(key string) (uint64, error) {
	x, err := defaultPool.IDInc(key)
	return uint64(x), err
}
func IDInc(key string) (int64, error) {
	return defaultPool.IDInc(key)
}
func IDIncDirect(key string) (int64, error) {
	return defaultPool.IDIncDirect(key, 1)
}

//
