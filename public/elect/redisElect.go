package elect

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/frame/runtime"
	"github.com/aiden2048/pkg/public/redisDeal"
	"github.com/aiden2048/pkg/public/redisKeys"
)

type RedisElect struct {
	name       string
	isMaster   bool
	nodeId     string
	clusterKey *redisKeys.RedisKeys
	stopCh     chan struct{}
}

func NewRedisElect(name, nodeId string) (*RedisElect, error) {
	if name == "" {
		return nil, errors.New("name is empty")
	}
	if nodeId == "" {
		return nil, errors.New("node id is empty")
	}
	return &RedisElect{
		name:       name,
		nodeId:     nodeId,
		clusterKey: &redisKeys.RedisKeys{Name: redisKeys.REDIS_INDEX_COMMON, Key: fmt.Sprintf("electCluster.%s.MasterNodeId", name)},
		stopCh:     make(chan struct{}),
	}, nil
}

func (r *RedisElect) Stop() {
	r.stopCh <- struct{}{}
}

func (r *RedisElect) getNodeIdFromRedis() string {
	var nodeId string
	val := redisDeal.RedisDoGetStr(r.clusterKey)
	info := strings.SplitN(val, ":", 2)
	if len(info) > 1 {
		nodeId = info[1]
	} else {
		nodeId = val
	}
	return nodeId
}

func (r *RedisElect) Run() {
	pTraceId := "redisElect:" + runtime.GenTraceStrId()
	runtime.Go(func() {
		firstLoop := true
		for {
			trace := runtime.GetTrace()
			trace.TraceId = pTraceId + "." + runtime.GenTraceStrId()
			runtime.StoreTrace(trace)
			if firstLoop {
				firstLoop = false
			} else {
				tm := time.NewTimer(time.Second * 2)
				select {
				case <-tm.C:
				case <-r.stopCh:
					r.isMaster = false
					tm.Stop()
					goto out
				}
				tm.Stop()
			}

			// 当前不是master
			if !r.isMaster {

				succ := redisDeal.RedisDoSetnx(r.clusterKey, fmt.Sprintf("%d:%s", time.Now().Unix(), r.nodeId), 15)
				logs.Debugf("elect redis setnx key:%s succ:%d", r.clusterKey.Key, succ)
				if succ == 0 {

					ttl, err := redisDeal.RedisGetTtl(r.clusterKey)
					if err != nil {
						logs.Errorf("err:%v", err)
						continue
					}
					logs.Debugf("elect redis setnx key:%s ttl:%d", r.clusterKey.Key, ttl)
					// 有可能上次进程退出时候还没来的及设置expire就退出了
					if ttl == -1 || ttl > 60 {

						if ttl == -1 {
							if err = redisDeal.RedisDoDel(r.clusterKey); err != nil {
								logs.Errorf("err:%v", err)
								continue
							}
						}
						val := redisDeal.RedisDoGetStr(r.clusterKey)
						info := strings.SplitN(val, ":", 2)
						if len(info) < 2 {
							if err = redisDeal.RedisDoDel(r.clusterKey); err != nil {
								logs.Errorf("err:%v", err)
							}
							continue
						}

						lastSuccAtStr := info[0]
						lastSuccAt, err := strconv.ParseInt(lastSuccAtStr, 10, 64)
						if err != nil {
							if err = redisDeal.RedisDoDel(r.clusterKey); err != nil {
								logs.Errorf("err:%v", err)
							}
							continue
						}

						if time.Now().Unix()-lastSuccAt > 60 {

							if err = redisDeal.RedisDoDel(r.clusterKey); err != nil {
								logs.Errorf("err:%v", err)
								continue
							}
						}
					}

					continue
				}

				r.isMaster = true

				err := redisDeal.RedisSetTtl(r.clusterKey, 15)
				if err != nil {
					logs.Errorf("err:%v", err)
					continue
				}

				continue
			}

			if r.getNodeIdFromRedis() != r.nodeId {
				logs.Debugf("elect redis setnx key:%s nodeId:%s r.nodeId:%s", r.clusterKey.Key, r.getNodeIdFromRedis(), r.nodeId)
				r.isMaster = false
				continue
			}

			err := redisDeal.RedisSetTtl(r.clusterKey, 12)
			if err != nil {
				logs.Errorf("err:%v", err)
				continue
			}
		}
	out:
		return
	})
}

func (r *RedisElect) GetClusterKey() string {
	return r.clusterKey.Key
}

func (r *RedisElect) GetName() string {
	return r.name
}

func (r *RedisElect) GetNodeId() string {
	return r.nodeId
}

func (r *RedisElect) IsMaster() bool {
	return r.isMaster
}
