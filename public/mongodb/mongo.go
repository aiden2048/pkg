package mongodb

import (
	"context"
	"fmt"
	"log"
	"sync"
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

var dbcs *sync.Map = &sync.Map{} // map[int8]*mongo.Client
var ordbcs *sync.Map = &sync.Map{}
var confDb *sync.Map = &sync.Map{}
var logDb *sync.Map = &sync.Map{}             // log库
var imageRepositoryDb *sync.Map = &sync.Map{} // 镜像库
var topDb *sync.Map = &sync.Map{}             // top库

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

func GetMgoUri(platId int32) string {
	if mgoUrl == "" {
		mgoUrl, _ = initCfg(frame.GetMgoCoinfig(platId).Real)
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
	StartPlatMgoDb(wl, frame.GetPlatformId(), dbs...)
	return nil

}
func StartPlatMgoDb(wl WriteC, platId int32, dbs ...int) (err error) {
	if len(dbs) == 0 || utils.InArray(dbs, RealKey) {
		_, ok := dbcs.Load(platId)
		if !ok {
			db, err := startReal(frame.GetMgoCoinfig(platId).Real, wl)
			if err != nil {
				return err
			}
			dbcs.Store(platId, db)
		}
		_, ok = ordbcs.Load(platId)
		if !ok {
			ordb, err := startOnlyRead(frame.GetMgoCoinfig(platId).Real, wl)
			if err != nil {
				return err
			}
			ordbcs.Store(platId, ordb)
		}
	}
	if len(dbs) == 0 || utils.InArray(dbs, ConfKey) {
		_, ok := confDb.Load(platId)

		if !ok {
			cDb, err := startReal(frame.GetMgoCoinfig(platId).Conf, wl)
			if err != nil {
				return err
			}
			confDb.Store(platId, cDb)
		}
	}
	if len(dbs) == 0 || utils.InArray(dbs, LogKey) {
		_, ok := logDb.Load(platId)
		if !ok {
			lDb, err := startReal(frame.GetMgoCoinfig(platId).Log, wl)
			if err != nil {
				return err
			}
			logDb.Store(platId, lDb)
		}
	}
	if utils.InArray(dbs, ImageRepositoryKey) {
		_, ok := imageRepositoryDb.Load(platId)
		if !ok {
			iDb, err := startReal(frame.GetMgoCoinfig(platId).ImageRepository, wl)
			if err != nil {
				return err
			}
			imageRepositoryDb.Store(platId, iDb)
		}
	}

	// 必须指定库才需要初始化
	if utils.InArray(dbs, TopKey) {
		_, ok := topDb.Load(platId)
		if !ok {
			tDb, err := startReal(frame.GetMgoCoinfig(platId).Top, wl)
			if err != nil {
				return err
			}
			topDb.Store(platId, tDb)
		}
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

func GetDbSession(key int8, platId int32) *mongo.Client {
	return getDbSession(key, platId)
}

func getDbSession(key int8, platId int32) *mongo.Client {
	switch key {
	case RealKey:
		db, ok := dbcs.Load(platId)
		if !ok {
			logs.Errorf("RealKey mongo 获取失败  致命错误")
			return nil
		}
		return db.(*mongo.Client)
	case RealReadKey:
		ordb, ok := ordbcs.Load(platId)
		if !ok {
			logs.Errorf("RealReadKey mongo 获取失败 致命错误")
			return nil
		}
		return ordb.(*mongo.Client)
	case ConfKey:
		cDb, ok := confDb.Load(platId)
		if !ok {
			logs.Errorf("ConfKey mongo 获取失败 致命错误")
			return nil
		}
		return cDb.(*mongo.Client)
	case ImageRepositoryKey:
		iDb, ok := imageRepositoryDb.Load(platId)
		if !ok {
			logs.Errorf("ImageRepositoryKey mongo 获取失败 致命错误")
			return nil
		}
		return iDb.(*mongo.Client)
	case TopKey:
		tDb, ok := topDb.Load(platId)
		if !ok {
			logs.Errorf("TopKey mongo 获取失败 致命错误")
			return nil
		}
		return tDb.(*mongo.Client)
	case LogKey:
		logDb, ok := logDb.Load(platId)
		if !ok {
			logs.Errorf("LogKey mongo 获取 失败 致命错误")
			return nil
		}
		return logDb.(*mongo.Client)
	default:
		return nil
	}
}
