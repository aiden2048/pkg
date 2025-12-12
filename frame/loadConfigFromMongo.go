package frame

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/aiden2048/pkg/frame/logs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"github.com/BurntSushi/toml"

	"time"
)

var db *mongo.Client

func initMongo(cfg *MgoSvrCfg) (*mongo.Client, error) {
	if db != nil {
		return db, nil
	}
	uri := cfg.SrvUrl
	if uri == "" {
		if cfg.Scheme == "" {
			cfg.Scheme = "mongodb" // mongodb+srv
		}
		uri = cfg.Scheme + "://" + cfg.GetUrl()
	} else {
		uri = fmt.Sprintf("mongodb+srv://%s:%s@%s", cfg.User, cfg.Password, cfg.SrvUrl)
	}
	log.Println("mongo url", uri)
	opts := options.Client().ApplyURI(uri)
	opts.SetMaxPoolSize(uint64(cfg.PoolNum))                               // 设置最大连接数
	opts.SetMinPoolSize(0)                                                 // 设置最小连接数
	opts.SetWriteConcern(writeconcern.New(writeconcern.WMajority()))       //请求确认写操作传播到大多数mongod实例
	opts.SetReadConcern(readconcern.Majority())                            //指定查询应返回实例的最新数据确认为，已写入副本集中的大多数成员
	opts.SetReadPreference(readpref.SecondaryPreferred())                  // 优先读从库
	opts.SetMaxConnIdleTime(time.Duration(cfg.ConnIdleTime) * time.Second) // 设置连接空闲时间 超过就会断开
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Println(uri)
		return nil, err
	}
	db = client
	return db, nil
}

type EnvConfig struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id"`
	Encrypt int32              `json:"encrypt" bson:"encrypt"`
	Config  string             `bson:"config" json:"config"` //配置内容
}

func LoadConfigFromMongo(key string, conf interface{}) error {
	if defFrameOption.DisableMgo {
		return nil
	}
	db, err := initMongo(&GetMgoCoinfig().Real)
	if err != nil {
		return err
	}
	cfg := &EnvConfig{}
	filter := bson.M{"file_name": key}
	dbname := fmt.Sprintf("mg_config_%d", GetPlatformId())
	opts := options.FindOne()
	opts.SetSort(bson.D{{"plat_id", -1}})
	err = db.Database(dbname).Collection("c_environment").FindOne(context.TODO(), filter, opts).Decode(cfg)
	logs.Infof("db.Query filter %+v err:%+v dbname:%s", filter, err, dbname)
	if err != nil {

		return err
	}
	logs.Infof("LoadConfigFromMongo %s: %+v", key, cfg)
	if cfg.Config == "" {
		return errors.New("NoConfig")
	}
	if cfg.Encrypt != 0 {
		cfg.Config = Decrypt(cfg.Config)
	}
	if strings.HasSuffix(key, ".json") {
		err = jsoniter.UnmarshalFromString(cfg.Config, conf)
	} else {
		_, err = toml.Decode(cfg.Config, conf)
	}

	if err != nil {
		logs.Infof("LoadConfigFromMongo %s: Decode:%+v", key, err)
		return err
	}
	logs.Infof("LoadConfigFromMongo %s: %+v", key, conf)
	return nil
}
