package redisDeal

import (
	"context"
	"fmt"
	"time"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/utils/baselib"

	//"github.com/gomodule/redigo/redis"
	"github.com/redis/go-redis/v9"
)

const LogMaxLen = 1024

var bIsLoad = false

const (
	REDIS_INDEX_COMMON     = "default"
	REDIS_INDEX_ECS        = "ecs"
	REDIS_INDEX_USER       = "user"
	REDIS_INDEX_SESSION    = "session"
	REDIS_INDEX_RISK       = "risk"
	REDIS_INDEX_PAY        = "pay"
	REDIS_INDEX_ACTIVITY   = "activity"
	REDIS_INDEX_RECORD     = "record"
	REDIS_INDEX_TOPURLMAIN = "topurlmain" // 短链接主库
	REDIS_INDEX_MAX        = 8
)

// var bRegistRedisConfig bool
var bRedisIsLoads map[string]bool
var nRedisIdleCounts map[string]int
var G_redisPool map[string]redis.UniversalClient

func init() {
	bRedisIsLoads = make(map[string]bool)
	bRedisIsLoads = make(map[string]bool)
	nRedisIdleCounts = make(map[string]int)
	G_redisPool = make(map[string]redis.UniversalClient)
}

func loadConfig() error {
	err, b := frame.LoadRedisConfig()
	if err != nil {
		return err
	}

	logs.Infof("LoadRedisConfig:%+v", frame.GetRedisConfig())
	for k, pool := range G_redisPool {
		if bRedisIsLoads[k] && b[k] && pool != nil {
			logs.Infof("ReOpen G_redisPool[%s]", k)
			StartRedis(nRedisIdleCounts[k])
			go func() {
				time.Sleep(time.Second * 3)
				if pool != nil {
					pool.Close()
				}
			}()
		}
	}

	return nil
}

func OpenRedis(sec string, idleCount int) error {

	if !bIsLoad {
		if err := loadConfig(); err != nil {
			return err
		}
		baselib.RegisterReloadFunc(loadConfig)
	}
	bIsLoad = true
	//加载之后才有配置
	redisCfg := frame.GetRedisConfig().GetRedisNode(sec)
	if redisCfg == nil {
		return fmt.Errorf("没有配置Redis:%s", sec)
	}
	idleCount = redisCfg.IdleCount
	if idleCount == 0 {
		idleCount = 1024
	}
	nRedisIdleCounts[sec] = idleCount

	redisSec := frame.GetRedisConfig().GetRedisNode(sec)
	if redisSec == nil || len(redisSec.Servers) == 0 {
		return fmt.Errorf("Not Find Redis:%s", sec)
	}
	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        redisSec.Servers,
		Password:     redisSec.Password,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		MaxRedirects: 5,
		DialTimeout:  3 * time.Second,

		PoolSize:       100,
		MinIdleConns:   5,
		RouteByLatency: true,
		OnConnect: func(ctx context.Context, conn *redis.Conn) error {
			logs.PrintInfo("新连接建立:", conn)
			return nil
		},
	})
	if rdb.Ping(context.Background()).Err() != nil {
		logs.Errorf("Redis %s 连接失败", sec)
		return fmt.Errorf("Redis %s 连接失败", sec)
	}
	G_redisPool[sec] = rdb

	logs.Infof("Conn to %s Redis: %+v", sec, frame.GetRedisConfig())

	return nil
}

func StartRedis(idleCount ...int) error {
	c := 1024
	if len(idleCount) > 0 {
		c = idleCount[0]
	}
	return OpenRedis(REDIS_INDEX_COMMON, c)
}

// ecs
func StartEcsRedis(idleCount ...int) error {
	c := 256
	if len(idleCount) > 0 {
		c = idleCount[0]
	}
	return OpenRedis(REDIS_INDEX_ECS, c)
}

// user
func StartUserRedis(idleCount ...int) error {
	c := 256
	if len(idleCount) > 0 {
		c = idleCount[0]
	}
	return OpenRedis(REDIS_INDEX_USER, c)
}

// session
func StartSessionRedis(idleCount ...int) error {
	c := 256
	if len(idleCount) > 0 {
		c = idleCount[0]
	}
	return OpenRedis(REDIS_INDEX_SESSION, c)
}

// topurl主库
func StartTopUrlMainRedis(idleCount ...int) error {
	c := 256
	if len(idleCount) > 0 {
		c = idleCount[0]
	}
	return OpenRedis(REDIS_INDEX_TOPURLMAIN, c)
}

// risk
func StartRiskRedis(idleCount ...int) error {
	c := 1024
	if len(idleCount) > 0 {
		c = idleCount[0]
	}
	return OpenRedis(REDIS_INDEX_RISK, c)
}

// pay
func StartPayRedis(idleCount ...int) error {
	c := 256
	if len(idleCount) > 0 {
		c = idleCount[0]
	}
	return OpenRedis(REDIS_INDEX_PAY, c)
}

// activity
func StartActivityRedis(idleCount ...int) error {
	c := 256
	if len(idleCount) > 0 {
		c = idleCount[0]
	}
	return OpenRedis(REDIS_INDEX_ACTIVITY, c)
}

// Record
func StartRecordRedis(idleCount ...int) error {
	c := 256
	if len(idleCount) > 0 {
		c = idleCount[0]
	}
	return OpenRedis(REDIS_INDEX_RECORD, c)
}
func GetRedisPool(sec ...string) redis.UniversalClient {
	s := REDIS_INDEX_COMMON
	if len(sec) > 0 {
		s = sec[0]
	}
	if p, ok := G_redisPool[s]; ok {
		return p
	}
	logs.LogError("找不到redis:%s, 暂时用默认redis代替,马上检查配置以及代码", s)
	return G_redisPool[REDIS_INDEX_COMMON] //避免死循环

}

func GetEcsRedisPool() redis.UniversalClient {
	return GetRedisPool(REDIS_INDEX_ECS)
}

func GetUserRedisPool() redis.UniversalClient {
	return GetRedisPool(REDIS_INDEX_USER)
}
