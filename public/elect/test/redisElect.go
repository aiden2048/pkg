package main

import (
	"fmt"
	"time"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/public/elect"
	"github.com/aiden2048/pkg/public/mongodb"
	"github.com/aiden2048/pkg/public/redisDeal"
	uuid "github.com/satori/go.uuid"
)

func main() {
	if err := frame.InitConfig("electTest", &frame.FrameOption{}); err != nil {
		fmt.Printf("InitConfig error:%s\n", err.Error())
		return
	}

	// 初始化mongo
	if err := mongodb.StartMgoDb(mongodb.WLevel1); err != nil {
		fmt.Printf("InitMongodb %+v failed\n: %s", frame.GetMgoCoinfig(), err.Error())
		return
	}

	if err := redisDeal.StartRedis(); err != nil {
		fmt.Printf("InitRedis failed: %s\n", err.Error())
		return
	}

	// 初始化REDIS
	if err := redisDeal.StartUserRedis(); err != nil {
		fmt.Printf("InitUserRedis failed: %s\n", err.Error())
		return
	}

	// if err := redisWrapper.SetRedisString("FK", "BAR"); err != nil {
	// 	fmt.Printf("err:%v", err)
	// 	return
	// }
	// fmt.Println(redis.Int(redisWrapper.RedisDo("TTL", "FK")))

	// err := redisWrapper.RedisSend("EXPIRE", "FK", 15)
	// if err != nil {
	// 	fmt.Println("err:", err)
	// 	return
	// }
	// fmt.Println(redis.Int(redisWrapper.RedisDo("TTL", "FK")))

	e, err := elect.NewRedisElect("testElect", uuid.NewV4().String())
	if err != nil {
		fmt.Println("err:", err)
	}

	e2, err := elect.NewRedisElect("testElect", uuid.NewV4().String())
	if err != nil {
		fmt.Println("err:", err)
	}

	e3, err := elect.NewRedisElect("testElect", uuid.NewV4().String())
	if err != nil {
		fmt.Println("err:", err)
	}

	e.Run()
	time.Sleep(time.Second)
	e2.Run()
	e3.Run()

	go func() {
		time.Sleep(time.Second * 10)
		e.Stop()
	}()

	for {
		fmt.Printf("e is master:%v\n", e.IsMaster())
		fmt.Printf("e2 is master:%v\n", e2.IsMaster())
		fmt.Printf("e3 is master:%v\n", e3.IsMaster())
		time.Sleep(time.Second * 3)
	}

}
