package mongodb

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aiden2048/pkg/public/redisDeal"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/frame/logs"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var db = &sync.Map{} //*mongo.Client
var ordb = &sync.Map{}

// var conforDb *mongo.Client
var mgoUrl string // mono的uri
type WriteC int8

// 默认数据库Key
var DefaultKey = int8(RealKey)

// 是否记录查询数量
var enableQueryRecord = false

// 是否停用索引上报
var disableReportMongoIndex = false

// 索引创建方式 0自己创建 1boot创建 2 boot创建 并创建后台索引
var indexCreateType = IndexCreateByBoot

const (
	RealKey     = 1 // 运行从库
	RealReadKey = 2 // 运行主库
)
const (
	WLevel1 WriteC = 1 // 写关注等级 1个节点确认
	WLevel2 WriteC = 2 // 写关注等级 半数节点确认
)

const (
	IndexCreateTypeSelf          = 0 // 自己创建索引
	IndexCreateByBoot            = 1 // boot 创建索引
	IndexCreateByBootWithOpIndex = 2 // boot 创建索引 并创建后台索引
)

func GetMgoUri() string {
	if mgoUrl == "" {
		mgoUrl, _ = initCfg(frame.GetMgoCoinfig().Real)
	}
	return mgoUrl
}
func EnableQueryRecord() {
	enableQueryRecord = true
}

func DisableMongoReport() {
	disableReportMongoIndex = true
}

func SetIndexCreateType(cType int) {
	indexCreateType = cType
}

func StartMgoDb(wl WriteC, plat_id ...int32) (err error) {
	pid := int32(frame.GetPlatformId())
	if len(plat_id) > 0 {
		pid = plat_id[0]
	}
	client, ok := db.Load(pid)
	if !ok || client == nil {
		dbi, err := startReal(frame.GetMgoCoinfig(pid).Real, wl)
		if err != nil {
			return err
		}
		db.Store(pid, dbi)
	}
	ordbi, ok := ordb.Load(pid)
	if !ok || ordbi == nil {
		ordbi, err = startOnlyRead(frame.GetMgoCoinfig(pid).Real, wl)
		if err != nil {
			return err
		}
		ordb.Store(pid, ordbi)
	}

	// 索引上报依赖redis
	if err = redisDeal.StartRedis(); err != nil {
		log.Fatalf("InitRedis failed: %s", err.Error())
		return err
	}
	return nil

}

func startReal(cfg frame.MgoSvrCfg, writeLevel WriteC) (*mongo.Client, error) {
	uri, cfg := initCfg(cfg)
	wc := writeconcern.New(writeconcern.WMajority())
	if writeLevel == WLevel1 {
		wc = writeconcern.New(writeconcern.W(1))
	}
	opts := options.Client().ApplyURI(uri)
	opts.SetMaxPoolSize(uint64(cfg.PoolNum))                               // 设置最大连接数
	opts.SetMinPoolSize(uint64(cfg.MinPoolNum))                            // 设置最小连接数
	opts.SetWriteConcern(wc)                                               // 写关注为1个节点确认 writeconcern.WMajority() 请求确认写操作传播到大多数mongod实例
	opts.SetReadConcern(readconcern.Majority())                            // 指定查询应返回实例的最新数据确认为，已写入副本集中的大多数成员
	opts.SetReadPreference(readpref.SecondaryPreferred())                  // 优先读从库
	opts.SetMaxConnIdleTime(time.Duration(cfg.ConnIdleTime) * time.Second) // 设置连接空闲时间 超过就会断开
	return mongo.Connect(context.TODO(), opts)
}
func startOnlyRead(cfg frame.MgoSvrCfg, writeLevel WriteC) (*mongo.Client, error) {
	uri, cfg := initCfg(cfg)
	wc := writeconcern.New(writeconcern.WMajority())
	if writeLevel == WLevel1 {
		wc = writeconcern.New(writeconcern.W(1))
	}
	opts := options.Client().ApplyURI(uri)
	opts.SetMaxPoolSize(uint64(cfg.PoolNum))                               // 设置最大连接数
	opts.SetMinPoolSize(uint64(cfg.MinPoolNum))                            // 设置最小连接数
	opts.SetWriteConcern(wc)                                               // 请求确认写操作传播到大多数mongod实例
	opts.SetReadConcern(readconcern.Local())                               // 指定查询应返回实例的最新数据确认为，当前指定节点
	opts.SetReadPreference(readpref.PrimaryPreferred())                    // 优先读主库
	opts.SetMaxConnIdleTime(time.Duration(cfg.ConnIdleTime) * time.Second) // 设置连接空闲时间 超过就会断开
	return mongo.Connect(context.TODO(), opts)
}
func initCfg(cfg frame.MgoSvrCfg) (string, frame.MgoSvrCfg) {
	uri := cfg.SrvUrl
	if uri == "" {
		if cfg.Scheme == "" {
			cfg.Scheme = "mongodb" // mongodb+srv
		}
		uri = cfg.Scheme + "://" + cfg.GetUrl()
	} else {
		uri = fmt.Sprintf("mongodb+srv://%s:%s@%s", cfg.User, cfg.Password, cfg.SrvUrl)
	}
	if cfg.ConnIdleTime == 0 {
		cfg.ConnIdleTime = 30
	}
	if cfg.PoolNum == 0 {
		cfg.PoolNum = 128
	}
	return uri, cfg
}

func GetMongoDb(plat_id ...int32) *mongo.Client {
	pid := int32(frame.GetPlatformId())
	if len(plat_id) > 0 {
		pid = plat_id[0]
	}
	client, ok := db.Load(pid)
	if !ok || client == nil {
		err := StartMgoDb(WLevel2, pid)
		if err != nil {
			logs.Errorf("mongo Db 获取失败  致命错误 %v", err)
			return nil
		}
		client, ok := db.Load(pid)
		if !ok || client == nil {
			logs.Errorf("mongo Db 获取失败  致命错误")
			return nil
		}
		return client.(*mongo.Client)
	}
	return client.(*mongo.Client)
}
func GetOrMongoDb(plat_id ...int32) *mongo.Client {
	pid := int32(frame.GetPlatformId())
	if len(plat_id) > 0 {
		pid = plat_id[0]
	}
	client, ok := db.Load(pid)
	if !ok || client == nil {
		err := StartMgoDb(WLevel2, pid)
		if err != nil {
			logs.Errorf("mongo Db 获取失败  致命错误 %v", err)
			return nil
		}
		client, ok := db.Load(pid)
		if !ok || client == nil {
			logs.Errorf("mongo Db 获取失败  致命错误")
			return nil
		}
		return client.(*mongo.Client)
	}
	return client.(*mongo.Client)
}
func GetDbSession(key int8, plat_id int32) *mongo.Client {
	return getDbSession(key, plat_id)
}

func getDbSession(key int8, plat_id int32) *mongo.Client {
	switch key {
	case RealKey:
		dbi := GetMongoDb(plat_id)
		if dbi == nil {
			logs.Errorf("RealKey mongo 获取失败  致命错误")
			return nil
		}
		return dbi
	case RealReadKey:
		ordbi := GetOrMongoDb(plat_id)
		if ordbi == nil {
			logs.Errorf("RealReadKey mongo 获取失败  致命错误")
			return nil
		}
		return ordbi
	default:
		return nil
	}
}
