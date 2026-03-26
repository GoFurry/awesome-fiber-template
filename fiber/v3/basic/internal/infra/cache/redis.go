package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	env "github.com/GoFurry/awesome-go-template/fiber/v3/basic/config"
	log "github.com/GoFurry/awesome-go-template/fiber/v3/basic/internal/infra/logging"
	"github.com/GoFurry/awesome-go-template/fiber/v3/basic/pkg/common"
	"github.com/redis/go-redis/v9"
)

var client *redis.Client
var ctx = context.Background()

func GetRedisService() *redis.Client { return client }

func RedisReady() bool { return client != nil }

func InitRedisOnStart() error {
	return connect()
}

func connect() error {
	connCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client = redis.NewClient(&redis.Options{
		Addr:      env.GetServerConfig().Redis.RedisAddr,
		Password:  env.GetServerConfig().Redis.RedisPassword,
		DB:        0,
		OnConnect: OnConnectFunc,
	})
	_, err := client.Ping(connCtx).Result()
	if err != nil {
		client = nil
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Debug("connected to redis ok.")
	return nil
}

func OnConnectFunc(ctx context.Context, cn *redis.Conn) error {
	log.Debug("new redis connect...")
	return nil
}

func Del(keys ...string) common.Error {
	if client == nil {
		return common.NewServiceError("redis service is not ready")
	}
	err := client.Del(ctx, keys...).Err()
	if err != nil {
		log.Error("删除缓存失败..." + err.Error())
		return common.NewServiceError("删除缓存失败.")
	}
	return nil
}

func SetNX(key string, value any, expiration time.Duration) bool {
	if client == nil {
		return false
	}
	boolVal, err := client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		log.Error("设置缓存失败..." + err.Error())
		return false
	}
	return boolVal
}

func Set(key string, value any) common.Error {
	return SetExpire(key, value, 0)
}

func SetExpire(key string, value any, expiration time.Duration) common.Error {
	if client == nil {
		return common.NewServiceError("redis service is not ready")
	}
	err := client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		log.Error("设置缓存失败..." + err.Error())
		return common.NewServiceError("设置缓存失败.")
	}
	return nil
}

func Get(key string) *redis.Cmd {
	if client == nil {
		cmd := redis.NewCmd(ctx)
		cmd.SetErr(errors.New("redis service is not ready"))
		return cmd
	}
	return client.Do(ctx, "get", key)
}

func GetString(key string) (data string, gfsError common.Error) {
	if client == nil {
		return "", common.NewServiceError("redis service is not ready")
	}
	val, err := client.Get(ctx, key).Result()

	switch {
	case errors.Is(err, redis.Nil):
		return "", nil
	case err != nil:
		log.Error("获取缓存失败..." + err.Error())
		return "", common.NewServiceError("获取缓存失败.")
	}
	return strings.TrimSpace(val), nil
}

func HSetMap(key string, kvMap map[string]string) common.Error {
	if client == nil {
		return common.NewServiceError("redis service is not ready")
	}
	err := client.HSet(ctx, key, kvMap).Err()
	if err != nil {
		log.Error("设置缓存失败..." + err.Error())
		return common.NewServiceError("设置缓存失败.")
	}
	return nil
}

func HSet(key string, fieldName string, fieldVal string) common.Error {
	if client == nil {
		return common.NewServiceError("redis service is not ready")
	}
	err := client.HSet(ctx, key, fieldName, fieldVal).Err()
	if err != nil {
		log.Error("设置缓存失败..." + err.Error())
		return common.NewServiceError("设置缓存失败.")
	}
	return nil
}

func HGet(key string, fieldName string) (data string, gfsError common.Error) {
	if client == nil {
		return "", common.NewServiceError("redis service is not ready")
	}
	res, err := client.HGet(ctx, key, fieldName).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return "", common.NewServiceError(key + "缓存不存在")
	case err != nil:
		log.Error("获取缓存失败..." + err.Error())
		return "", common.NewServiceError("获取缓存失败.")
	}
	return res, nil
}

func HMGet(key string, fields ...string) (data []interface{}, gfsError common.Error) {
	if client == nil {
		return nil, common.NewServiceError("redis service is not ready")
	}
	res, err := client.HMGet(ctx, key, fields...).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return nil, common.NewServiceError(key + "缓存不存在")
	case err != nil:
		log.Error("获取缓存失败..." + err.Error())
		return nil, common.NewServiceError("获取缓存失败.")
	}
	return res, nil
}

func HGetAll(key string) (data map[string]string, gfsError common.Error) {
	if client == nil {
		return nil, common.NewServiceError("redis service is not ready")
	}
	res, err := client.HGetAll(ctx, key).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return nil, common.NewServiceError(key + "缓存不存在")
	case err != nil:
		log.Error("获取缓存失败..." + err.Error())
		return nil, common.NewServiceError("获取缓存失败.")
	}
	return res, nil
}

func HDel(key string, fields ...string) (res int64, gfsError common.Error) {
	if client == nil {
		return 0, common.NewServiceError("redis service is not ready")
	}
	intVal, err := client.HDel(ctx, key, fields...).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return 0, common.NewServiceError(key + "缓存不存在")
	case err != nil:
		log.Error("获取缓存失败..." + err.Error())
		return intVal, nil
	}
	return intVal, nil
}

func Incr(key string) {
	if client == nil {
		return
	}
	client.Incr(ctx, key)
}

func CountByPrefix(prefix string) (res int64, gfsError common.Error) {
	if client == nil {
		return 0, common.NewServiceError("redis service is not ready")
	}
	var cursor uint64
	var count int
	pattern := prefix + "*"

	for {
		keys, newCursor, err := client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return 0, common.NewServiceError("redis统计失败.")
		}

		count += len(keys)
		cursor = newCursor

		if cursor == 0 {
			break
		}
	}

	return int64(count), nil
}

func DelByPrefix(prefix string) common.Error {
	if client == nil {
		return common.NewServiceError("redis service is not ready")
	}
	var cursor uint64
	pattern := prefix + "*"

	for {
		keys, newCursor, err := client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Error(fmt.Sprintf("redis scan err:%v", err))
			return common.NewServiceError(err.Error())
		}
		if len(keys) != 0 {
			err := Del(keys...)
			if err != nil {
				log.Error(fmt.Sprintf("redis del err:%v", err))
				return err
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

func FindByPrefix(prefix string) ([]string, common.Error) {
	if client == nil {
		return nil, common.NewServiceError("redis service is not ready")
	}
	var cursor uint64
	var resList []string
	pattern := prefix + "*"

	for {
		keys, newCursor, err := client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Error(fmt.Sprintf("redis scan err:%v", err))
			return nil, common.NewServiceError(err.Error())
		}
		if len(keys) != 0 {
			resList = append(resList, keys...)
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}
	return resList, nil
}

func PipelineExec(fn func(pipe redis.Pipeliner)) common.Error {
	if client == nil {
		return common.NewServiceError("redis service is not ready")
	}
	pipe := client.Pipeline()
	fn(pipe)
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Error("Pipeline执行失败: " + err.Error())
		return common.NewServiceError("缓存批量操作失败")
	}
	return nil
}
