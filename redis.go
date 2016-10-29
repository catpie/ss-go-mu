package main

import (
	. "github.com/catpie/ss-go-mu/log"
	redis "gopkg.in/redis.v3"
	"time"
)

const (
	DefaultExpireTime          = 0
	DefaultOnlineKeyExpireTime = time.Minute * 5
)

var Redis = new(RedisClient)

type RedisClient struct {
	client *redis.Client
}

func (r *RedisClient) SetClient(client *redis.Client) {
	r.client = client
}

func (r *RedisClient) Exists(u UserInterface) (bool, error) {
	return r.client.Exists(genUserInfoKey(u)).Result()
}

func (r *RedisClient) Del(u UserInterface) error {
	return r.client.Del(genUserInfoKey(u)).Err()
}

func (r *RedisClient) ClearAll() error {
	return r.client.FlushAll().Err()
}

// traffic
func (r *RedisClient) IncrSize(u UserInterface, size int) error {
	key := genUserFlowKey(u)
	incrSize := int(float32(size))
	isExits, err := r.client.Exists(key).Result()
	if err != nil {
		return err
	}
	if !isExits {
		return r.client.Set(key, incrSize, DefaultExpireTime).Err()
	}
	return r.client.IncrBy(key, int64(incrSize)).Err()
}

func (r *RedisClient) GetSize(u UserInterface) (int64, error) {
	key := genUserFlowKey(u)
	isExits, err := r.client.Exists(key).Result()
	if err != nil {
		return 0, err
	}
	if !isExits {
		return 0, nil
	}
	return r.client.Get(key).Int64()
}

func (r *RedisClient) SetSize(u UserInterface, size int) error {
	key := genUserFlowKey(u)
	return r.client.Set(key, size, DefaultExpireTime).Err()
}

func (r *RedisClient) ClearSize() error {
	return nil
}

func (r *RedisClient) MarkUserOnline(u UserInterface) error {
	key := genUserOnlineKey(u)
	return r.client.Set(key, "1", DefaultOnlineKeyExpireTime).Err()
}

func (r *RedisClient) IsUserOnline(u UserInterface) bool {
	key := genUserOnlineKey(u)
	isExits, err := r.client.Exists(key).Result()
	if err != nil {
		return false
	}
	return isExits
}

func (r *RedisClient) GetOnlineUsersCount(users []UserInterface) int {
	count := 0
	for _, v := range users {
		if r.IsUserOnline(v) {
			count++
		}
	}
	return count
}

func InitRedis() error {
	conf := config.Redis
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Host,
		Password: conf.Pass, // no password set
		DB:       conf.Db,   // use default DB
	})

	pong, err := client.Ping().Result()
	if err != nil {
		return err
	}
	Log.Info(pong)
	Redis.SetClient(client)
	// set storage
	SetStorage(Redis)
	return nil
}
