package mouselib

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
)

// redis客户端操作out-of-box函数

func NewRedisClient(addr string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})
}

// StrSet 设置单个string值
func StrSet(rds *redis.Client, key, val string) error {
	ctx := context.Background()
	return rds.Set(ctx, key, val, 0).Err()
}

// StrGet 获取单个string值
func StrGet(rds *redis.Client, key string) string {
	ctx := context.Background()
	return rds.Get(ctx, key).Val()
}

// Keys 查询keys
func Keys(rds *redis.Client, query string) []string {
	ctx := context.Background()
	return rds.Keys(ctx, query).Val()
}

// Del 获取单个string值
func Del(rds *redis.Client, keys ...string) error {
	ctx := context.Background()
	return rds.Del(ctx, keys...).Err()
}

// Lock 获取锁
func Lock(rds *redis.Client, key string, exp time.Duration) bool {
	ctx := context.Background()
	return rds.SetNX(ctx, key, "", exp).Val()
}

// UnLock 释放锁
func UnLock(rds *redis.Client, key string) error {
	return Del(rds, key)
}

// ListSet 创建一个列表
func ListSetAndLpush(rds *redis.Client, key string, vals ...string) error {
	ctx := context.Background()
	return rds.RPush(ctx, key, vals).Err()
}

// ListRange 查找列表的值
func ListRange(rds *redis.Client, key string, start, end int64) []string {
	ctx := context.Background()
	return rds.LRange(ctx, key, start, end).Val()
}

// ListLen 查找列表的长度
func ListLen(rds *redis.Client, key string, start, end int64) int {
	ctx := context.Background()
	return int(rds.LLen(ctx, key).Val())
}

// ListModify 根据索引值修改列表的值
func ListModify(rds *redis.Client, key string, idx int64, newVal string) error {
	ctx := context.Background()
	return rds.LSet(ctx, key, idx, newVal).Err()
}

// ListLpop 列表头移除值并返回
func ListLpop(rds *redis.Client, key string) string {
	ctx := context.Background()
	return rds.LPop(ctx, key).String()
}

// ListRpop 列表尾移除值并返回
func ListRpop(rds *redis.Client, key string) string {
	ctx := context.Background()
	return rds.RPop(ctx, key).String()
}

// HashSet 插入hash值
func HashSet(rds *redis.Client, key string, kv map[string]string) error {
	params := make([]string, 0, len(kv)*2)
	for k, v := range kv {
		params = append(params, k, v)
	}
	ctx := context.Background()
	return rds.HSet(ctx, key, params).Err()
}

// HashGet 获取hash中某个属性的值
func HashGet(rds *redis.Client, key string, field string) string {
	ctx := context.Background()
	return rds.HGet(ctx, key, field).String()
}

// HashGetAll 获取hash的所有值
func HashGetAll(rds *redis.Client, key string) map[string]string {
	ctx := context.Background()
	return rds.HGetAll(ctx, key).Val()
}

// HashKeys 获取hash中所有的key
func HashKeys(rds *redis.Client, key string) []string {
	ctx := context.Background()
	return rds.HKeys(ctx, key).Val()
}

// HashLen 获取hash的长度
func HashLen(rds *redis.Client, key string) int {
	ctx := context.Background()
	return int(rds.HLen(ctx, key).Val())
}

// HashDel 删除hash中某些属性的值
func HashDel(rds *redis.Client, key string, fields ...string) error {
	ctx := context.Background()
	return rds.HDel(ctx, key, fields...).Err()
}

// ZSetAdd 创建/添加zset值
func ZSetAdd(rds *redis.Client, key string, members map[float64]string) error {
	ctx := context.Background()
	return rds.ZAddArgs(ctx, key, redis.ZAddArgs{
		NX: true,
		Members: func(m map[float64]string) []redis.Z {
			r := make([]redis.Z, 0)
			for score, mem := range m {
				r = append(r, redis.Z{
					Score:  score,
					Member: mem,
				})
			}
			return r
		}(members),
	}).Err()
}

// ZCount 根据score获取范围内值的个数
func ZCount(rds *redis.Client, key string, min, max float64) int {
	mi, mx := fmt.Sprint(min), fmt.Sprint(max)
	ctx := context.Background()
	return int(rds.ZCount(ctx, key, mi, mx).Val())
}

// ZCard 获取zset的值的个数
func ZCard(rds *redis.Client, key string) int {
	ctx := context.Background()
	return int(rds.ZCard(ctx, key).Val())
}

// ZRange 根据score范围获取所有的值
func ZRange(rds *redis.Client, key string, min, max float64) []string {
	ctx := context.Background()
	return rds.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:     key,
		ByScore: true,
		Start:   min,
		Stop:    max,
	}).Val()
}
