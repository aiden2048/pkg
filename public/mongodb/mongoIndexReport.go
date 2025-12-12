package mongodb

import (
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/aiden2048/pkg/frame/runtime"
)

// 定时器
var ticker *time.Ticker

// mongo 结构体上报
type MongoIndexReportInfo struct {
	DbName          string   `json:"db_name"`
	TbName          string   `json:"tb_name"`
	Version         int      `json:"version"`           // 版本号
	Expire          int      `json:"expire"`            // 过期时间
	IndexList       []string `json:"index_list"`        // 业务索引
	UniqueIndexList []string `json:"unique_index_list"` // 唯一索引
	ExpireKeyList   []string `json:"expire_key_list"`   // 过期索引
	OpIndexList     []string `json:"op_index_list"`     // 后台索引
	SvrName         string   `json:"svr_name"`          // 上报进程
	DdKey           int8     `json:"db_key"`            // 所属库
	IndexCreateType int      `json:"index_create_type"` // 索引创建方式 0 本进程创建 1 boot创建 2
}

// 记录查询所用到的key  按字母顺序排序
func genQueryKey(query any, isAggregate bool) string {
	if query == nil {
		return ""
	}
	resultArr := []string{}
	if isAggregate { // 如果是聚合 只匹配最前面的match
		if arrD, ok := query.([]primitive.D); ok {
			for _, bsonD := range arrD {
				for _, bsonE := range bsonD {
					if bsonE.Key != "$match" {
						break
					}
					resultArr = append(resultArr, genOneQueryKey("", bsonE.Value)...)
				}
			}
		}
	} else {
		resultArr = genOneQueryKey("", query)
	}
	if len(resultArr) == 0 {
		return ""
	}
	keyMap := map[string]bool{}
	for _, key := range resultArr {
		keyMap[key] = true
	}

	finalArr := []string{}
	for key, _ := range keyMap {
		finalArr = append(finalArr, key)
	}
	sort.Strings(finalArr)
	return strings.Join(finalArr, ",")
}

func genOneQueryKey(key string, value any) []string {
	// 普通key直接返回
	if key != "" && !strings.Contains(key, "$") {
		return []string{key}
	}
	if value == nil {
		return []string{}
	}
	resultArr := []string{}
	mux := sync.Mutex{}
	if bsonM, ok := value.(primitive.M); ok {
		// bsonM 有可能被修改
		mux.Lock()
		for k, v := range bsonM {
			resultArr = append(resultArr, genOneQueryKey(k, v)...)
		}
		mux.Unlock()
	}
	if bsonD, ok := value.(primitive.D); ok {
		for _, bsonE := range bsonD {
			resultArr = append(resultArr, genOneQueryKey(bsonE.Key, bsonE.Value)...)
		}
	}
	if arrM, ok := value.([]primitive.M); ok {
		for _, bsonM := range arrM {
			for k, v := range bsonM {
				resultArr = append(resultArr, genOneQueryKey(k, v)...)
			}
		}
	}
	if arrD, ok := value.([]primitive.D); ok {
		for _, bsonD := range arrD {
			for _, bsonE := range bsonD {
				resultArr = append(resultArr, genOneQueryKey(bsonE.Key, bsonE.Value)...)
			}
		}
	}
	if filter, ok := value.(Filter); ok {
		for _, bsonE := range filter.D() {
			resultArr = append(resultArr, genOneQueryKey(bsonE.Key, bsonE.Value)...)
		}
	}
	return resultArr
}

func startLocalTicker() {
	if ticker != nil {
		return
	}
	runtime.Go(func() {
		if ticker != nil {
			return
		}
		ticker = time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				queryRecordMap.Range(func(k, v any) bool {
					key, ok := k.(string)
					if !ok {
						return true
					}
					count, ok := v.(int64)
					arr := strings.Split(key, "|")
					if len(arr) != 3 {
						return true
					}
					dbName := GetRealMongoName(arr[0])
					tbName := GetRealMongoName(arr[1])
					queryKey := arr[2]
					event := &ReportQueryCountMsgEvent{
						DbName:   dbName,
						TbName:   tbName,
						QueryKey: queryKey,
						Count:    count,
					}
					event.Send()
					// redisKey := redisKeys.GenMongoQueryHashKey(dbName, tbName, 0)
					// if redisDeal.GetRedisPool(redisKey.Name) != nil {
					// 	redisDeal.RedisSendHincrby(redisKey, queryKey, count, 4*24*3600)
					// }
					return true
				})
				locker := sync.Mutex{}
				locker.Lock()
				queryRecordMap = sync.Map{}
				locker.Unlock()
			}
		}
	})
}

func GetRealMongoName(name string) string {
	arr := strings.Split(name, "_")
	if len(arr) == 0 {
		return name
	}
	nameArr := []string{}
	for _, value := range arr {
		if !IsNumeric(value) {
			nameArr = append(nameArr, value)
		}
	}
	return strings.Join(nameArr, "_")
}

func IsNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
