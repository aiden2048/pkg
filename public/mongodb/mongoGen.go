package mongodb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/frame/stat"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/public/redisDeal"
	"github.com/aiden2048/pkg/public/redisKeys"
	"github.com/aiden2048/pkg/utils"

	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var maxQryTimeout = time.Second * 5

func SetMaxQryTimeout(timeout time.Duration) {
	maxQryTimeout = timeout
}

var indexMap = sync.Map{}

// 索引上报
var indexReportMap = sync.Map{}

// 查询统计计数
var queryRecordMap = sync.Map{}

type IndexS struct {
	Name     string
	IsUindex bool
	Keys     []string
}
type Mon struct {
	PlatId    int32
	dbKey     int8
	cache     bool
	indexKey  []string
	expireKey []string
	uindexKey []string
	db        string
	tb        string
	Indexs    []*IndexS
}

// type oid struct {
// 	ID primitive.ObjectID `bson:"_id"`
// }

func NewMon(db, tb string, cache bool, index, uindex, expire []string, dbKey ...int8) *Mon {
	conn := &Mon{
		db:        db,
		tb:        tb,
		dbKey:     DefaultKey, // 默认运行从库
		PlatId:    frame.GetPlatformId(),
		cache:     cache,
		indexKey:  index,
		uindexKey: uindex,
		expireKey: expire,
	}
	if len(dbKey) > 0 {
		conn.dbKey = dbKey[0]
	}
	conn.creatIndex()
	return conn
}

func (m *Mon) SetDbKey(key int8) {
	m.dbKey = key
}
func (m *Mon) SetPlatId(plat_id int32) {
	m.PlatId = plat_id
}
func (m *Mon) SetDB(db string) {
	m.db = db
}
func (m *Mon) SetTB(tb string) {
	m.tb = tb
}
func (m *Mon) GetDB() string {
	return m.db
}
func (m *Mon) GetTB() string {
	return m.tb
}
func (m *Mon) GetColl() *mongo.Collection {
	DB := getDbSession(m.dbKey, m.PlatId)
	return DB.Database(m.db).Collection(m.tb)
}
func (m *Mon) ExpireKey(keys []string) {
	m.expireKey = keys
}

// IgnoreConflictInsertOne 简单的检查唯一性查询后插入新数据的逻辑，应该采用IgnoreConflictInsertOne获取更优的性能
// IgnoreConflictInsertOne通过数据库唯一索引检查完成查询后数据不存在插入,省略了多一次查询的性能损耗
func (m *Mon) IgnoreConflictInsertOne(ctx context.Context, data any) (err error) {
	start := time.Now()
	// 检测添加索引

	defer func() {
		logs.Infof("IgnoreConflictInsertOne,db:%s,tb:%s,cost:%+v,err:%+v", m.db, m.tb, data, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("IgnoreConflictInsertOne,db:%s,tb:%s,data:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, data, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	DB := getDbSession(m.dbKey, m.PlatId)
	_, err = DB.Database(m.db).Collection(m.tb).InsertOne(ctx, data)
	ret := 0
	if errors.Is(err, context.DeadlineExceeded) {
		ret = 2
	} else if err != nil {
		var mgoErr mongo.WriteException
		if errors.As(err, &mgoErr) {
			if len(mgoErr.WriteErrors) > 0 && mgoErr.WriteErrors[0].Code == 11000 {
				// 这种可预见错误不做统计
				return nil
			}
		}

		ret = 1
		logs.PrintError("mongo.InsertOne", m.db, m.tb, data, "err", err)
	}
	stat.ReportStat("rp:mongo.InsertOne."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	return err
}

func (m *Mon) InsertOne(ctx context.Context, data any) (err error) {
	start := time.Now()
	// 检测添加索引
	defer func() {
		logs.Infof("InsertOne,db:%s,tb:%s,data:%+v,cost:%+v,err:%+v", m.db, m.tb, data, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("InsertOne,db:%s,tb:%s,data:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, data, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	DB := getDbSession(m.dbKey, m.PlatId)
	_, err = DB.Database(m.db).Collection(m.tb).InsertOne(ctx, data)
	ret := 0
	if err == context.DeadlineExceeded {
		ret = 2
	} else if err != nil {
		var mgoErr mongo.WriteException
		var isDuplicateKeyErr bool
		if errors.As(err, &mgoErr) {
			if len(mgoErr.WriteErrors) > 0 && mgoErr.WriteErrors[0].Code == 11000 {
				isDuplicateKeyErr = true
			}
		}
		if !isDuplicateKeyErr {
			logs.PrintError("mongo.InsertOne", m.db, m.tb, data, "err", err)
		}
		ret = 1

	}
	stat.ReportStat("rp:mongo.InsertOne."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	return err
}

func (m *Mon) InsertMany(ctx context.Context, data []any) (err error) {
	start := time.Now()
	// 检测添加索引
	defer func() {
		logs.Infof("InsertMany,db:%s,tb:%s,data:%+v,cost:%+v,err:%+v", m.db, m.tb, data, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("InsertMany,db:%s,tb:%s,data:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, data, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	DB := getDbSession(m.dbKey, m.PlatId)
	_, err = DB.Database(m.db).Collection(m.tb).InsertMany(ctx, data)
	ret := 0
	if err == context.DeadlineExceeded {
		ret = 2
	} else if err != nil {
		logs.PrintError("mongo.InsertMany", m.db, m.tb, data, "err", err)
		ret = 1
	}
	stat.ReportStat("rp:mongo.InsertMany."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	return err
}

func (m *Mon) UpdateOne(ctx context.Context, filter, update any) (err error) {
	start := time.Now()
	defer func() {
		logs.Infof("UpdateOne,db:%s,tb:%s,filter:%+v,data:%+v,cost:%+v,err:%+v", m.db, m.tb, filter, update, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("UpdateOne,db:%s,tb:%s,filter:%+v,data:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, filter, update, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	go m.RecordMongoQuery(filter, false)
	DB := getDbSession(m.dbKey, m.PlatId)
	// update["updated_at"] = time.Now()
	_, err = DB.Database(m.db).Collection(m.tb).UpdateOne(ctx, filter, update)
	ret := 0
	if err == context.DeadlineExceeded {
		ret = 2
	} else if err != nil {
		ret = 1
	}
	stat.ReportStat("rp:mongo.UpdateOne."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	if err != nil {
		logs.PrintError("mongo.UpdateOne", m.db, m.tb, filter, update, "err", err)
		return err
	}

	return nil
}
func (m *Mon) UpdateMany(ctx context.Context, filter, update any) (err error) {
	start := time.Now()
	defer func() {
		logs.Infof("UpdateMany,db:%s,tb:%s,filter:%+v,data:%+v,cost:%+v,err:%+v", m.db, m.tb, filter, update, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("UpdateMany,db:%s,tb:%s,filter:%+v,data:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, filter, update, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	go m.RecordMongoQuery(filter, false)
	DB := getDbSession(m.dbKey, m.PlatId)
	// update["updated_at"] = time.Now()
	_, err = DB.Database(m.db).Collection(m.tb).UpdateMany(ctx, filter, update)

	ret := 0
	if err == context.DeadlineExceeded {
		ret = 2
	} else if err != nil {
		ret = 1
	}
	stat.ReportStat("rp:mongo.UpdateMany."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	if err != nil {
		logs.PrintError("mongo.UpdateMany", m.db, m.tb, filter, update, "err", err)
		return err
	}

	return nil
}

// 新增或者更新
func (m *Mon) UpSertWithRes(ctx context.Context, filter, update any) (res *mongo.UpdateResult, err error) {
	start := time.Now()
	go m.RecordMongoQuery(filter, false)
	// 检测添加索引
	defer func() {
		logs.Infof("UpSertWithRes,db:%s,tb:%s,filter:%+v,data:%+v,cost:%+v,err:%+v", m.db, m.tb, filter, update, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("UpSertWithRes,db:%s,tb:%s,filter:%+v,data:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, filter, update, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	DB := getDbSession(m.dbKey, m.PlatId)
	opts := options.Update().SetUpsert(true)
	// update["updated_at"] = time.Now()
	res, err = DB.Database(m.db).Collection(m.tb).UpdateOne(ctx, filter, update, opts)
	ret := 0
	if err == context.DeadlineExceeded {
		ret = 2
	} else if err != nil {
		ret = 1
	}
	stat.ReportStat("rp:mongo.UpSert."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	if err != nil {
		logs.PrintError("mongo.UpSert", m.db, m.tb, filter, update, "err", err)
		return nil, err
	}

	return res, nil
}

// 新增或者更新
func (m *Mon) UpSert(ctx context.Context, filter, update any) (err error) {
	start := time.Now()
	go m.RecordMongoQuery(filter, false)
	// 检测添加索引
	defer func() {
		logs.Infof("UpSert,db:%s,tb:%s,filter:%+v,data:%+v,cost:%+v,err:%+v", m.db, m.tb, filter, update, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("UpSert,db:%s,tb:%s,filter:%+v,data:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, filter, update, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	DB := getDbSession(m.dbKey, m.PlatId)
	opts := options.Update().SetUpsert(true)
	// update["updated_at"] = time.Now()
	_, err = DB.Database(m.db).Collection(m.tb).UpdateOne(ctx, filter, update, opts)
	ret := 0
	if err == context.DeadlineExceeded {
		ret = 2
	} else if err != nil {
		ret = 1
	}
	stat.ReportStat("rp:mongo.UpSert."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	if err != nil {
		logs.PrintError("mongo.UpSert", m.db, m.tb, filter, update, "err", err)
		return err
	}

	return nil
}

func (m *Mon) DeleteOne(ctx context.Context, filter any) (err error) {
	start := time.Now()
	defer func() {
		logs.Infof("DeleteOne,db:%s,tb:%s,filter:%+v,cost:%+v,err:%+v", m.db, m.tb, filter, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("DeleteOne,db:%s,tb:%s,filter:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, filter, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	DB := getDbSession(m.dbKey, m.PlatId)
	_, err = DB.Database(m.db).Collection(m.tb).DeleteOne(ctx, filter)
	ret := 0
	if err == context.DeadlineExceeded {
		ret = 2
	} else if err != nil {
		ret = 1
	}
	stat.ReportStat("rp:mongo.DeleteOne."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	if err != nil {
		return err
	}
	return nil
}
func (m *Mon) DelCache(ctx context.Context, keys []string) {
	for _, k := range keys {
		ks := fmt.Sprintf("cache.%s.%s.%s", m.db, m.tb, k)
		key := &redisKeys.RedisKeys{
			Name: redisDeal.REDIS_INDEX_COMMON,
			Key:  ks,
		}
		err := redisDeal.RedisDoDel(key)
		if err != nil {
			logs.Errorf("更新缓存数据出错,db:%s,tb:%s,err:%+v", m.db, m.tb, err)
		}
	}
}
func (m *Mon) DeleteMany(ctx context.Context, filter any) (err error) {
	start := time.Now()
	defer func() {
		logs.Infof("DeleteMany,db:%s,tb:%s,filter:%+v,cost:%+v,err:%+v", m.db, m.tb, filter, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("DeleteMany,db:%s,tb:%s,filter:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, filter, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	DB := getDbSession(m.dbKey, m.PlatId)
	_, err = DB.Database(m.db).Collection(m.tb).DeleteMany(ctx, filter)
	ret := 0
	if err == context.DeadlineExceeded {
		ret = 2
	} else if err != nil {
		ret = 1
	}
	stat.ReportStat("rp:mongo.DeleteMany."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	if err != nil {
		return err
	}
	//stat.ReportStat("rp:mongo.FindOne."+m.db+"."+m.tb, 0, time.Now().Sub(start))
	return nil
}

func (m *Mon) FindOne(ctx context.Context, filter any, rsp any) (err error) {
	start := time.Now()
	defer func() {
		logs.Infof("FindOne,db:%s,tb:%s,filter:%+v,cost:%+v,err:%+v", m.db, m.tb, filter, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("FindOne,db:%s,tb:%s,filter:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, filter, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	go m.RecordMongoQuery(filter, false)
	DB := getDbSession(m.dbKey, m.PlatId)
	if DB == nil {
		err = errors.New("dbconnet is nil")
		return
	}
	err = DB.Database(m.db).Collection(m.tb).FindOne(ctx, filter).Decode(rsp)
	ret := 0
	if err == context.DeadlineExceeded {
		ret = 2
	} else if err != nil {
		ret = 1
	}
	stat.ReportStat("rp:mongo.FindOne."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	//stat.ReportStat("rp:mongo.FindOne."+m.db+"."+m.tb, 0, time.Now().Sub(start))
	return
}

func (m *Mon) FindOneByCacheKey(ctx context.Context, filter bson.M, ckey string, rsp any) (err error) {
	start := time.Now()
	defer func() {
		logs.Infof("FindOneByCacheKey,db:%s,tb:%s,filter:%+v,cost:%+v,err:%+v", m.db, m.tb, filter, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("FindOneByCacheKey,db:%s,tb:%s,filter:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, filter, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	DB := getDbSession(m.dbKey, m.PlatId)
	if DB == nil {
		err = errors.New("dbconnet is nil")
		return
	}
	if !m.cache {
		err = DB.Database(m.db).Collection(m.tb).FindOne(ctx, filter).Decode(rsp)
		return
	}
	defer func() {
		ret := 0
		if err != nil {
			ret = 1
		}
		stat.ReportStat("rp:mongo.FindOneByCacheKey."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	}()
	// 读缓存
	ks := fmt.Sprintf("cache.%s.%s.%s", m.db, m.tb, ckey)
	key := &redisKeys.RedisKeys{
		Name: redisDeal.REDIS_INDEX_COMMON,
		Key:  ks,
	}
	str := redisDeal.RedisDoGetStr(key)
	if str != "" {
		err := jsoniter.UnmarshalFromString(str, rsp)
		if err == nil {
			return nil
		}
	}

	// 读库
	err = DB.Database(m.db).Collection(m.tb).FindOne(ctx, filter).Decode(rsp)
	if err != nil {
		return err
	}
	// 写缓存
	err = redisDeal.RedisDoSet(key, rsp, utils.RandInt64(3600, 7200))
	if err != nil {
		logs.Errorf("更新缓存数据出错,db:%s,tb:%s,err:%+v", m.db, m.tb, err)
	}

	return nil

}
func (m *Mon) FindOneByID(ctx context.Context, id string, rsp any) (err error) {
	start := time.Now()
	defer func() {
		logs.Infof("FindOneByIDdb:%s,tb:%s,id:%+v,cost:%+v,err:%+v", m.db, m.tb, id, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("FindOneByIDdb:%s,tb:%s,id:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, id, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	DB := getDbSession(m.dbKey, m.PlatId)
	oid, err := primitive.ObjectIDFromHex(id)
	defer func() {
		ret := 0
		if err != nil {
			ret = 1
		}
		stat.ReportStat("rp:mongo.FindOneByID."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	}()
	if err != nil {
		return err
	}

	return DB.Database(m.db).Collection(m.tb).FindOne(ctx, bson.M{"_id": oid}).Decode(rsp)
}
func (m *Mon) FindAllV1(ctx context.Context, filter any, selector ...interface{}) (curs *mongo.Cursor, err error) {
	start := time.Now()
	// if !m.checkUseIndex(filter) {

	// }
	defer func() {
		logs.Infof("FindAll:%s,tb:%s,filter:%+v,cost:%+v,err:%+v", m.db, m.tb, filter, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("FindAll:%s,tb:%s,filter:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, filter, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	go m.RecordMongoQuery(filter, true)
	DB := getDbSession(m.dbKey, m.PlatId)
	opts := options.Find()
	if len(selector) > 0 {
		opts.SetProjection(selector[0])
	}
	if filter == nil {
		filter = Filter{}
	}

	curs, err = DB.Database(m.db).Collection(m.tb).Find(ctx, filter, opts)
	ret := 0
	if err != nil {
		ret = 1
	}
	stat.ReportStat("rp:mongo.FindAll."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	return
}
func (m *Mon) FindAllBySortV1(ctx context.Context, filter any, sort bson.D, selector ...interface{}) (curs *mongo.Cursor, err error) {

	start := time.Now()
	defer func() {
		logs.Infof("FindAll:%s,tb:%s,filter:%+v,cost:%+v,err:%+v", m.db, m.tb, filter, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("FindAll:%s,tb:%s,filter:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, filter, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	go m.RecordMongoQuery(filter, false)
	DB := getDbSession(m.dbKey, m.PlatId)
	opts := options.Find()
	if sort != nil {
		opts.SetSort(sort)
	}

	if len(selector) > 0 {
		opts.SetProjection(selector[0])
	}
	if filter == nil {
		filter = Filter{}
	}
	defer func() {
		ret := 0
		if err != nil {
			ret = 1
		}
		stat.ReportStat("rp:mongo.FindAllBySortV1."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	}()
	curs, err = DB.Database(m.db).Collection(m.tb).Find(ctx, filter, opts)
	return
}
func (m *Mon) FindManyV1(ctx context.Context, filter any, sort bson.D, limit int64, selector ...interface{}) (curs *mongo.Cursor, err error) {
	start := time.Now()
	defer func() {
		logs.Infof("FindMany:%s,tb:%s,filter:%+v,cost:%+v,err:%+v", m.db, m.tb, filter, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("FindMany:%s,tb:%s,filter:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, filter, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	go m.RecordMongoQuery(filter, false)
	DB := getDbSession(m.dbKey, m.PlatId)
	opts := options.Find()
	if sort != nil {
		opts.SetSort(sort)
	}
	opts.SetLimit(limit)
	if len(selector) > 0 {
		opts.SetProjection(selector[0])
	}
	if filter == nil {
		filter = bson.M{}
	}
	defer func() {
		ret := 0
		if err != nil {
			ret = 1
		}
		stat.ReportStat("rp:mongo.FindManyV1."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	}()
	curs, err = DB.Database(m.db).Collection(m.tb).Find(ctx, filter, opts)
	//stat.ReportStat("rp:mongo.FindMany."+m.db+"."+m.tb, 0, time.Now().Sub(start))
	return
}
func (m *Mon) FindManyByPageV1(ctx context.Context, filter any, sort bson.D, offset, limit int64, selector ...interface{}) (curs *mongo.Cursor, err error) {
	start := time.Now()
	defer func() {
		logs.Infof("FindManyByPage:%s,tb:%s,filter:%+v,cost:%+v,err:%+v", m.db, m.tb, filter, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("FindManyByPage:%s,tb:%s,filter:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, filter, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	go m.RecordMongoQuery(filter, false)
	DB := getDbSession(m.dbKey, m.PlatId)
	opts := options.Find()
	if sort != nil {
		opts.SetSort(sort)
	}
	opts.SetSkip(offset)
	opts.SetLimit(limit)
	if len(selector) > 0 {
		opts.SetProjection(selector[0])
	}
	if filter == nil {
		filter = bson.M{}
	}
	defer func() {
		ret := 0
		if err != nil {
			ret = 1
		}
		stat.ReportStat("rp:mongo.FindManyByPageV1."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	}()
	curs, err = DB.Database(m.db).Collection(m.tb).Find(ctx, filter, opts)
	//stat.ReportStat("rp:mongo.FindManyByPage."+m.db+"."+m.tb, 0, time.Now().Sub(start))
	return
}
func (m *Mon) FindCount(ctx context.Context, filter any) (id int64, err error) {
	start := time.Now()
	defer func() {
		logs.Infof("FindCount,db:%s,tb:%s,filter:%+v,err:%+v", m.db, m.tb, filter, err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("FindCount:%s,tb:%s,id:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, filter, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
		go m.RecordMongoQuery(filter, false)
		ret := 0
		if err != nil {
			ret = 1
		}
		stat.ReportStat("rp:mongo.FindCount."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	}()

	DB := getDbSession(m.dbKey, m.PlatId)
	return DB.Database(m.db).Collection(m.tb).CountDocuments(ctx, filter)
}
func (m *Mon) Aggregate(ctx context.Context, pipe []bson.D, opts ...*options.AggregateOptions) (ret []bson.M, err error) {
	start := time.Now()
	defer func() {
		logs.Infof("Aggregate:%s,tb:%s,pipe:%+v,cost:%+v,err:%+v", m.db, m.tb, pipe, time.Now().Sub(start), err)
		if time.Now().Sub(start) > maxQryTimeout {
			logs.Warnf("Aggregate:%s,tb:%s,filter:%+v,执行时间 > %d ms,cost:%+v,err:%+v", m.db, m.tb, pipe, maxQryTimeout.Milliseconds(), time.Now().Sub(start), err)
		}
	}()
	go m.RecordMongoQuery(pipe, true)

	DB := getDbSession(m.dbKey, m.PlatId)
	curs, _err := DB.Database(m.db).Collection(m.tb).Aggregate(ctx, pipe, opts...)
	if _err != nil {
		err = _err
		return nil, err
	}
	defer curs.Close(ctx)
	defer func() {
		ret := 0
		if err != nil {
			ret = 1
		}
		stat.ReportStat("rp:mongo.Aggregate."+m.db+"."+m.tb, ret, time.Now().Sub(start))
	}()
	ret = []bson.M{}
	err = curs.All(ctx, &ret)

	if err != nil {
		return
	}
	//stat.ReportStat("rp:mongo.Aggregate."+m.db+"."+m.tb, 0, time.Now().Sub(start))
	return ret, nil
}
func (m *Mon) Trans(ctx context.Context, fn func(mongo.SessionContext) error) (err error) {
	DB := getDbSession(m.dbKey, m.PlatId)
	err = DB.UseSession(ctx, func(sc mongo.SessionContext) error {
		var err error
		err = sc.StartTransaction()
		if err != nil {
			return err
		}
		err = fn(sc)
		if err != nil {
			logs.Errorf("事务撤销出错,db:%s,tb:%s,err:%+v", m.db, m.tb, err)
			sc.AbortTransaction(context.TODO())
			return err
		}
		err = sc.CommitTransaction(context.TODO())
		if err != nil {
			logs.Errorf("事务提交出错,db:%s,tb:%s,err:%+v", m.db, m.tb, err)
		}
		return err
	})
	return err
}

func (m *Mon) clearIndexs(ctx context.Context, keys []string) {

	//特殊处理. 外接游戏订单库不清理索引
	if m.db == "global_provider_game" {
		return
	}
	dbi := getDbSession(m.dbKey, m.PlatId)
	cursor, err := dbi.Database(m.db).Collection(m.tb).Indexes().List(ctx)
	if err != nil {
		logs.PrintError("获取索引失败", m.db, m.tb)
		return
	}
	for cursor.Next(ctx) {
		var index *bson.D
		errs := cursor.Decode(&index)
		if errs != nil || index == nil {
			continue
		}
		var kname string
		for _, v := range *index {
			if v.Key == "name" {
				kname, _ = v.Value.(string)
				break
			}
		}
		if utils.InArray(keys, kname) {
			continue
		}
		if kname == "_id_" {
			continue
		}
		if strings.HasPrefix(kname, "expire_at") {
			continue
		}
		if strings.HasPrefix(kname, "mindex_") {
			continue
		}
		if strings.HasPrefix(kname, "top_") {
			continue
		}
		if strings.HasPrefix(kname, "op_") {
			continue
		}
		if strings.HasSuffix(kname, "_hash") {
			continue
		}
		if strings.HasSuffix(kname, "_hashed") {
			continue
		}
		if strings.HasPrefix(m.db, "mg_config") /*|| strings.HasPrefix(m.db, "mg_static")*/ {
			logs.PrintError("+++ 准备删除老索引 +++", m.db, m.tb, kname, index)
			dbi.Database(m.db).Collection(m.tb).Indexes().DropOne(ctx, kname, options.DropIndexes())
		} else {
			logs.PrintBill("建议删除老索引", m.db, m.tb, kname, index)

		}

	}
	cursor.Close(ctx)
}
func (m *Mon) creatIndex(tctx ...context.Context) {
	uindexKey := m.uindexKey
	indexKey := []string{}
	expireKey := []string{}
	if indexCreateType == IndexCreateTypeSelf {
		indexKey = m.indexKey
		expireKey = m.expireKey
	}
	if len(indexKey) == 0 && len(uindexKey) == 0 && len(expireKey) == 0 {
		return
	}
	key := fmt.Sprintf("%s.%s", m.db, m.tb)
	_, ok := indexMap.Load(key)
	indexMap.Store(key, true)
	if ok {
		return
	}
	ctx := context.TODO()
	// index 按库和表生成redis
	// 读缓存
	allKey := []string{} // 最后生成redis
	models := []mongo.IndexModel{}
	allKeyNames := []string{}
	for _, in := range indexKey {
		allKey = append(allKey, in)
		model := mongo.IndexModel{}
		bsonD := bson.D{}
		kname := "mindex"
		idx := &IndexS{IsUindex: false}
		for _, k := range strings.Split(in, ",") {
			value := 1
			if strings.HasPrefix(k, "-") {
				k = k[1:]
				value = -1
			}
			bsonD = append(bsonD, bson.E{Key: k, Value: value})
			kname += "_" + k
			idx.Keys = append(idx.Keys, k)
		}
		model.Keys = bsonD
		model.Options = options.Index().SetName(kname)
		models = append(models, model)

		allKeyNames = append(allKeyNames, kname)

		idx.Name = kname
		m.Indexs = append(m.Indexs, idx)

	}
	for _, in := range expireKey {
		allKey = append(allKey, in)
		model := mongo.IndexModel{}
		bsonD := bson.D{}
		kname := "mindex"
		for _, k := range strings.Split(in, ",") {
			value := 1
			if strings.HasPrefix(k, "-") {
				k = k[1:]
				value = -1
			}
			bsonD = append(bsonD, bson.E{Key: k, Value: value})
			kname += "_" + k
		}
		model.Keys = bsonD
		model.Options = options.Index().SetName(kname)
		model.Options.SetExpireAfterSeconds(0)
		models = append(models, model)
		allKeyNames = append(allKeyNames, kname)
	}
	for _, in := range uindexKey {
		allKey = append(allKey, in)
		model := mongo.IndexModel{}
		bsonD := bson.D{}
		kname := "mindex"
		idx := &IndexS{IsUindex: true}
		for _, k := range strings.Split(in, ",") {
			value := 1
			if strings.HasPrefix(k, "-") {
				k = k[1:]
				value = -1
			}
			bsonD = append(bsonD, bson.E{Key: k, Value: value})
			kname += "_" + k
			idx.Keys = append(idx.Keys, k)
		}
		model.Keys = bsonD
		model.Options = options.Index().SetUnique(true).SetBackground(true).SetName(kname)
		models = append(models, model)

		allKeyNames = append(allKeyNames, kname)

		idx.Name = kname
		m.Indexs = append(m.Indexs, idx)

	}
	//清理老索引

	m.clearIndexs(ctx, allKeyNames)
	if len(allKey) == 0 {
		return
	}
	DB := getDbSession(m.dbKey, m.PlatId)
	_, err := DB.Database(m.db).Collection(m.tb).Indexes().CreateMany(ctx, models)
	if err != nil {
		if strings.HasPrefix(m.db, "mg_config") /*|| strings.HasPrefix(m.db, "mg_static")*/ {
			logs.PrintError("CreatIndex_Err", "创建索引失败:", m.db, m.tb, err, allKey)
		} else {
			logs.PrintBill("CreatIndex_Err", "创建索引失败:", m.db, m.tb, err, allKey)
		}
	}
}

func isArrayEqual(s []string, filter map[string]bool) bool {
	for _, k := range s {
		if _, ok := filter[k]; !ok {
			return false
		}
	}
	return true
}
func getKeysFromFilter(filter bson.M) map[string]bool {
	//递归遍历filter, 找出所有的key, 去重
	return map[string]bool{}
}
func (m *Mon) checkFilterIndex(filter bson.M) (bool, *IndexS) {
	if filter == nil || len(filter) == 0 {
		return false, nil
	}
	keys := getKeysFromFilter(filter)
	if keys == nil || len(keys) == 0 {
		return false, nil
	}
	for _, index := range m.Indexs {
		if len(index.Keys) < len(keys) {
			continue
		}
		if isArrayEqual(index.Keys[:len(keys)], keys) {
			return true, index
		}
	}
	return false, nil
}

func (m *Mon) CreateIndex(tctx ...context.Context) {
	m.creatIndex()
}
func (m *Mon) CreatIndexByIndexKey(ctx context.Context, indexKey []string) {
	if len(indexKey) == 0 {
		return
	}
	// index 按库和表生成redis
	models := []mongo.IndexModel{}

	for _, in := range indexKey {
		key := fmt.Sprintf("%s.%s.%s", m.db, m.tb, in)
		_, ok := indexMap.Load(key)
		if ok {
			continue
		}
		model := mongo.IndexModel{}
		bsonD := bson.D{}
		for _, k := range strings.Split(in, ",") {
			value := 1
			if strings.HasPrefix(k, "-") {
				k = k[1:]
				value = -1
			}
			bsonD = append(bsonD, bson.E{Key: k, Value: value})
		}
		model.Keys = bsonD
		model.Options = options.Index().SetBackground(true)
		models = append(models, model)
	}
	DB := getDbSession(m.dbKey, m.PlatId)
	_, err := DB.Database(m.db).Collection(m.tb).Indexes().CreateMany(ctx, models)
	if err != nil {
		logs.Errorf("创建索引失败,db:%s,tb:%s,err:%+v，index:%s", m.db, m.tb, err, indexKey)
	}
	for _, in := range indexKey {
		key := fmt.Sprintf("%s.%s.%s", m.db, m.tb, in)
		indexMap.Store(key, 1)
	}
}
func (m *Mon) DeleteIndex(ctx context.Context, indexName string) error {
	DB := getDbSession(m.dbKey, m.PlatId)
	_, err := DB.Database(m.db).Collection(m.tb).Indexes().DropOne(ctx, indexName)
	return err
}

// 上报索引
func (m *Mon) ReportMongoIndex(opIndexList []string, expireKey []string, version int, expire int) {
	if disableReportMongoIndex {
		return
	}
	localKey := fmt.Sprintf("%s_%s", m.db, m.tb)
	if _, ok := indexReportMap.Load(localKey); ok {
		return
	}
	reportInfo := &ReportIndexMsgEvent{}
	reportInfo.DbName = m.db
	reportInfo.TbName = m.tb
	reportInfo.IndexList = m.indexKey
	reportInfo.UniqueIndexList = m.uindexKey
	reportInfo.Version = version
	reportInfo.Expire = expire
	reportInfo.ExpireKeyList = expireKey
	reportInfo.SvrName = frame.GetServerName()
	reportInfo.Send()
	// redisKey := redisKeys.GenMongoIndexReportSetKey()
	// if redisDeal.GetRedisPool(redisKey.Name) != nil {
	// 	go redisDeal.RedisSendSadd(redisKey, reportInfo)
	// }
	indexReportMap.Store(localKey, 1)
}

func (m *Mon) RecordMongoQuery(query any, isAggregate bool) {
	if !enableQueryRecord {
		return
	}
	if query == nil {
		return
	}

	queryKey := genQueryKey(query, isAggregate)
	if queryKey == "" {
		queryKey = "{}"
	}
	key := fmt.Sprintf("%s|%s|%s", m.db, m.tb, queryKey)
	storeCount, ok := queryRecordMap.Load(key)
	if !ok {
		queryRecordMap.Store(key, int64(1))
	} else {
		count, ok := storeCount.(int64)
		if !ok {
			count = 0
		}
		count++
		queryRecordMap.Store(key, count)
	}

	if ticker == nil {
		startLocalTicker()
	}

}

func (m *Mon) checkUseIndex(filter Filter) bool {
	if len(filter) == 0 {
		return true
	}
	kstr := ""
	if m.dbKey == RealKey || m.dbKey == RealReadKey {
		for _, E := range filter {
			kstr += fmt.Sprintf("%s,", E.Key)
		}
		for _, index := range m.indexKey {
			if strings.HasPrefix(index, kstr) {
				return true
			}
		}
	}
	return false
}
func ToBsonM(data interface{}) (bson.M, error) {
	bsonData, err := bson.Marshal(data)
	if err != nil {
		return nil, err
	}
	var bsonM bson.M
	err = bson.Unmarshal(bsonData, &bsonM)
	if err != nil {
		return nil, err
	}
	return bsonM, nil
}
