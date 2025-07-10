package mongodb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aiden2048/pkg/public/redisDeal"
	"github.com/aiden2048/pkg/utils"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/frame/logs"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var db *mongo.Client
var ordb *mongo.Client
var confDb *mongo.Client
var logDb *mongo.Client             // log库
var imageRepositoryDb *mongo.Client // 镜像库
var topDb *mongo.Client             // top库

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
	RealKey            = 1 // 运行从库
	RealReadKey        = 2 // 运行主库
	ConfKey            = 3 // 配置从库
	ImageRepositoryKey = 4 // 镜像库
	TopKey             = 5 // top 数据库
	LogKey             = 6 // 配置从库
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

func StartMgoDb(wl WriteC, dbs ...int) (err error) {
	if len(dbs) == 0 || utils.InArray(dbs, RealKey) {
		if db == nil {
			db, err = startReal(frame.GetMgoCoinfig().Real, wl)
			if err != nil {
				return err
			}
		}
		if ordb == nil {
			ordb, err = startOnlyRead(frame.GetMgoCoinfig().Real, wl)
			if err != nil {
				return err
			}
		}
	}
	if len(dbs) == 0 || utils.InArray(dbs, ConfKey) {

		if confDb == nil {
			confDb, err = startReal(frame.GetMgoCoinfig().Conf, wl)
			if err != nil {
				return err
			}
		}
	}
	if len(dbs) == 0 || utils.InArray(dbs, LogKey) {
		if logDb == nil {
			logDb, err = startReal(frame.GetMgoCoinfig().Log, wl)
			if err != nil {
				return err
			}
		}
	}
	// 必须指定库才需要初始化
	if utils.InArray(dbs, ImageRepositoryKey) {
		if imageRepositoryDb == nil {
			imageRepositoryDb, err = startReal(frame.GetMgoCoinfig().ImageRepository, wl)
			if err != nil {
				return err
			}
		}
	}

	// 必须指定库才需要初始化
	if utils.InArray(dbs, TopKey) {
		if topDb == nil {
			topDb, err = startReal(frame.GetMgoCoinfig().Top, wl)
			if err != nil {
				return err
			}
		}
	}
	// 索引上报依赖redis
	if err = redisDeal.StartRedis(); err != nil {
		log.Fatalf("InitRedis failed: %s", err.Error())
		return err
	}
	return nil

}

// 设置默认数据库 （top使用）
func SetDefaultKey(key int8) error {
	DefaultKey = key
	if getDbSession(key) == nil {
		logs.Errorf("mongo 获取失败  致命错误 key:%d", key)
		return errors.New("mongo 获取失败  致命错误 ")
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
	//if cfg.MinPoolNum == 0 {
	//	cfg.MinPoolNum = 16
	//}
	if cfg.ConnIdleTime == 0 {
		cfg.ConnIdleTime = 30
	}
	if cfg.PoolNum == 0 {
		cfg.PoolNum = 128
	}
	return uri, cfg
}

// GetMongoDb change stream 使用，一般不推荐使用这个接口
func GetMongoDb() *mongo.Client {
	if db == nil {
		logs.Errorf("mongo Db 获取失败  致命错误")
		return nil
	}
	return db
}

// GetMongoLogDb change stream 使用，一般不推荐使用这个接口
func GetMongoLogDb() *mongo.Client {
	if logDb == nil {
		logs.Errorf("mongo logDb 获取失败  致命错误")
		return nil
	}
	return logDb
}
func GetDbSession(key int8) *mongo.Client {
	return getDbSession(key)
}

func GetMongoImageDb() *mongo.Client {
	if imageRepositoryDb == nil {
		logs.Errorf("mongo imageDb 获取失败  致命错误")
		return nil
	}
	return imageRepositoryDb
}

func getDbSession(key int8) *mongo.Client {
	switch key {
	case RealKey:
		if db == nil {
			logs.Errorf("RealKey mongo 获取失败  致命错误")
			return nil
		}
		return db
	case RealReadKey:
		if ordb == nil {
			logs.Errorf("RealReadKey mongo 获取失败 致命错误")
			return nil
		}
		return ordb
	case ConfKey:
		if confDb == nil {
			logs.Errorf("ConfKey mongo 获取失败 致命错误")
			return nil
		}
		return confDb
	case ImageRepositoryKey:
		if imageRepositoryDb == nil {
			logs.Errorf("ImageRepositoryKey mongo 获取失败 致命错误")
			return nil
		}
		return imageRepositoryDb
	case TopKey:
		if topDb == nil {
			logs.Errorf("TopKey mongo 获取失败 致命错误")
			return nil
		}
		return topDb
	case LogKey:
		if logDb == nil {
			logs.Errorf("LogKey mongo 获取 失败 致命错误")
			return nil
		}
		return logDb
	default:
		return nil
	}
}
