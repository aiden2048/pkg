package frame

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aiden2048/pkg/frame/stat"
	jsoniter "github.com/json-iterator/go"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/utils"
	"github.com/nats-io/nats.go"
)

const (
	StreamDefaultTTL      = 3 * time.Hour
	StreamDefaultReplicas = 3
)

type ConsumeOption struct {
	BatchSize   int // 取数每批数据条数
	MaxWaitMsec int // 取数最大等待时间：毫秒
}

var gStreamSubjMap = make(map[string]bool) // 主题是否创建
var gStreamSubjMutex sync.Mutex            // 加锁避免重复创建，并确保按顺序添加主题(跨进程可能冲突但多次check最后会修复)

func GetNatsStreamContext() nats.JetStreamContext {
	conn := GetNatsConn()
	if conn == nil {
		logs.Errorf("GetNatsConn Failed")
		return nil
	}
	stream, err := conn.JetStream()
	if err != nil {
		logs.Errorf("GetJetStream Failed")
		return nil
	}
	return stream
}

func CheckNatsStreamSubject(streamName, streamSubj string) error {
	gStreamSubjMutex.Lock()
	defer gStreamSubjMutex.Unlock()
	if _, ok := gStreamSubjMap[streamName+"."+streamSubj]; ok {
		return nil
	}
	stream := GetNatsStreamContext()
	if stream == nil {
		return errors.New("GetStreamContext Failed")
	}
	if info, err := stream.StreamInfo(streamName); info == nil { // stream不存在
		if _, err = stream.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{streamSubj},
			MaxAge:   StreamDefaultTTL,
			Replicas: StreamDefaultReplicas,
		}); err != nil { // 新建失败返回error
			logs.Errorf("AddStream Name:%s Subject:%s err:%+v", streamName, streamSubj, err)
			return err
		}
		logs.Importantf("AddStream Name:%s Subject:%s OK", streamName, streamSubj)
	} else if !utils.InArray(info.Config.Subjects, streamSubj) { // stream存在，但主题不存在
		info.Config.Subjects = append(info.Config.Subjects, streamSubj)
		info.Config.MaxAge = StreamDefaultTTL
		info.Config.Replicas = StreamDefaultReplicas
		if _, err = stream.UpdateStream(&info.Config); err != nil { // 更新失败返回error
			logs.Errorf("UpdateStream Name:%s Subject:%s err:%+v", streamName, streamSubj, err)
			return err
		}
		logs.Importantf("UpdateStream Name:%s Subject:%s OK", streamName, streamSubj)
	} else if info.Config.MaxAge != StreamDefaultTTL || info.Config.Replicas != StreamDefaultReplicas { // 修改参数
		info.Config.MaxAge = StreamDefaultTTL
		info.Config.Replicas = StreamDefaultReplicas
		if _, err = stream.UpdateStream(&info.Config); err != nil { // 更新失败忽略错误
			logs.Errorf("UpdateStream Name:%s MaxAge:%v Replicas:%d err:%+v", streamName, info.Config.MaxAge, info.Config.Replicas, err)
		} else {
			logs.Importantf("UpdateStream Name:%s MaxAge:%v Replicas:%d OK", streamName, info.Config.MaxAge, info.Config.Replicas)
		}
	} // else: 主题已存在，且参数无需修改
	gStreamSubjMap[streamName+"."+streamSubj] = true
	return nil
}

var natsStreamMap = sync.Map{} // (map[string]bool) // 缓存stream是否存在

func CreateNatsStream(streamName string) error {
	if _, ok := natsStreamMap.Load(streamName); ok {
		return nil
	}
	js := GetNatsStreamContext()
	if js == nil {
		return errors.New("GetStreamContext Failed")
	}
	natsStreamMap.Store(streamName, true)
	if stream, _ := js.StreamInfo(streamName); stream == nil {
		_, err := js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{streamName + ".*"},
			MaxAge:   StreamDefaultTTL,
			Replicas: StreamDefaultReplicas,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func GetNatsStreamKey(streamName, sub string) string {
	return fmt.Sprintf("%s_%s", streamName, sub)
}

// PublishSub 按app_id推送就设置app_id 不分开推送设置 app_id=0
func PublishSub(streamName, sub string, app_id int32, v any, sess *Session) error {

	//err := PublishSubEx(streamName, sub, app_id, v, sess)
	//if err != nil {
	//	return err
	//}
	streamName = fmt.Sprintf("%d_%s", GetPlatformId(), streamName)
	return PublishSubEx(streamName, sub, app_id, v, sess)
	//return err
}

func PublishSubEx(streamName, sub string, app_id int32, v any, sess *Session) error {
	stream := GetNatsStreamContext()
	if stream == nil {
		return errors.New("GetStreamContext Failed")
	}
	streamkey := GetNatsStreamKey(streamName, sub)
	if err := CreateNatsStream(streamkey); err != nil {
		logs.Errorf("Stream:%s subj:%s Publish Checksubect err:%+v", streamkey, sub, err)
		return err
	}

	s, err := jsoniter.Marshal(v)
	if err != nil {
		logs.Warnf("Stream:%s subj:%s Publish Marshal data:%+v err:%+v", streamkey, sub, v, err)
		return err
	}
	if sess == nil {
		sess = NewSessionOnlyApp(app_id)
	}
	logs.Infof("pub-events:%+v", v)
	//Session{SvrFE: GetServerName(), SvrID: GetServerID() /*, Cmd: ename*/}
	qgmsg := NatsTransMsg{Sess: *sess}
	//需要修正下发起者为本进程
	qgmsg.GetSession().SvrFE = GetServerName()
	qgmsg.GetSession().SvrID = GetServerID()
	qgmsg.GetSession().Channel = 0
	qgmsg.MsgBody = s
	qs, err := jsoniter.Marshal(qgmsg)
	if err != nil {
		logs.Warnf("Stream:%s subj:%s Publish Marshal data:%+v err:%+v", streamkey, sub, v, err)
		return err
	}
	msg := &nats.Msg{Subject: fmt.Sprintf("%s.%d", streamkey, app_id), Data: qs}
	msg.Header = nats.Header{}
	msg.Header.Set("PublishTime", time.Now().Format(time.RFC3339Nano))
	if _, err = stream.PublishMsg(msg); err != nil {
		logs.Bill("Publish_NatsMsg_Fail", "Stream:%s Subject:%s Publish data:%+v err:%+v", streamName, msg.Subject, v, err)
		return err
	}
	return nil
}

func Publish(streamName, streamSubj string, v any) error {
	stream := GetNatsStreamContext()
	if stream == nil {
		return errors.New("GetStreamContext Failed")
	}

	if err := CheckNatsStreamSubject(streamName, streamSubj); err != nil {
		logs.Errorf("Stream:%s Subj:%s Publish CheckStreamSubject err:%+v", streamName, streamSubj, err)
		return err
	}

	s, err := jsoniter.Marshal(v)
	if err != nil {
		logs.Warnf("Stream:%s Subj:%s Publish Marshal data:%+v err:%+v", streamName, streamSubj, v, err)
		return err
	}
	msg := &nats.Msg{Subject: streamSubj, Data: s}
	msg.Header = nats.Header{}
	msg.Header.Set("PublishTime", time.Now().Format(time.RFC3339Nano))
	if _, err = stream.PublishMsg(msg); err != nil {
		logs.Warnf("Stream:%s Subj:%s Publish data:%+v err:%+v", streamName, streamSubj, v, err)
		return err
	}
	return nil
}

func RegisterNatsPullConsumer(streamName, streamSubj, consumer string, logic func([]*nats.Msg) error, option ...*ConsumeOption) error {
	op := &ConsumeOption{}
	if len(option) > 0 {
		op = option[0]
	}
	if op.BatchSize <= 0 {
		op.BatchSize = 10000
	}
	if op.MaxWaitMsec <= 0 {
		op.MaxWaitMsec = 10000
	}

	stream := GetNatsStreamContext()
	if stream == nil {
		return errors.New("GetStreamContext Failed")
	}

	if err := CheckNatsStreamSubject(streamName, streamSubj); err != nil {
		logs.Errorf("Stream:%s Subj:%s RegisterPullConsumer CheckStreamSubject err:%+v", streamName, streamSubj, err)
		return err
	}

	if info, err := stream.ConsumerInfo(streamName, consumer); info != nil && info.Config.FilterSubject != streamSubj {
		// 自动解绑旧的主题，但如果真的命名冲突则可能导致其他消费者异常
		if err = stream.DeleteConsumer(streamName, consumer); err != nil {
			logs.Errorf("Stream:%s Subj:%s RegisterPullConsumer DeleteConsumer err:%+v", streamName, streamSubj, err)
			return err
		}
		logs.Importantf("Stream:%s Subject:%s RegisterPullConsumer Unbind Old Subject:%s", streamName, streamSubj, info.Config.FilterSubject)
	}

	sub, err := stream.PullSubscribe(streamSubj, consumer)
	if err != nil {
		logs.Errorf("Stream:%s Subject:%s Consumer:%s PullSubscribe err:%+v", streamName, streamSubj, consumer, err)
		return err
	}

	go func() {
		defer sub.Unsubscribe()
		for {
			startTime := time.Now()
			msg, err := sub.Fetch(op.BatchSize, nats.MaxWait(time.Duration(op.MaxWaitMsec)*time.Millisecond))
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, nats.ErrTimeout) {
					logs.Infof("Stream:%s Subject:%s Consumer:%s Waiting New Msg ...", streamName, streamSubj, consumer)
				} else {
					logs.Warnf("Stream:%s Subject:%s Consumer:%s err:%+v", streamName, streamSubj, consumer, err)
				}
			} else if len(msg) > 0 {
				err = logic(msg)
				if err != nil {
					logs.Warnf("Stream:%s Subject:%s Consumer:%s err:%+v", streamName, streamSubj, consumer, err)
				} else {
					doTime := time.Now()
					for _, m := range msg {
						pubTime := m.Header.Get("PublishTime")
						if pubTime != "" {
							pt, errs := time.Parse(time.RFC3339Nano, pubTime)
							if errs == nil {
								stat.ReportStat("Nats.PullStream."+streamName+"."+streamSubj, 0, doTime.Sub(pt))
							}
						}
						err = m.Ack()
						if err != nil {
							logs.Warnf("Stream:%s Subject:%s Consumer:%s Ack Msg:%v err:%+v", streamName, streamSubj, consumer, string(m.Data), err)
						}
					}
				}
			}
			costMsec := time.Now().Sub(startTime).Milliseconds()
			if 2*len(msg) < op.BatchSize && costMsec < int64(op.MaxWaitMsec) {
				// 如果取到数据较少就等待
				time.Sleep(time.Duration(int64(op.MaxWaitMsec)-costMsec) * time.Millisecond)
			}
		}
	}()

	logs.Importantf("Register Stream:%s Subject:%s Consumer:%s Start", streamName, streamSubj, consumer)
	return nil
}

var gConsumerMap = make(map[string]bool)
var gConsumerMutex sync.Mutex

func RegisterNatsMultiPullConsumer(streamPattern, subjPattern, consumerSuffix string, logic func([]*nats.Msg) error, option ...*ConsumeOption) {
	stream := GetNatsStreamContext()
	if stream == nil {
		return
	}
	// 输入正则，或者带%d
	streamRegex := regexp.MustCompile("^" + strings.ReplaceAll(streamPattern, "%d", "\\d+") + "$")
	subjRegex := regexp.MustCompile("^" + strings.ReplaceAll(subjPattern, "%d", "\\d+") + "$")
	go func() {
		for {
			time.Sleep(30 * time.Second)
			logs.Importantf("RegisterMultiPullConsumer Check New Stream Like %s.%s ...", streamPattern, subjPattern)
			for info := range stream.Streams() {
				name := info.Config.Name
				for _, subj := range info.Config.Subjects {
					key := strings.ReplaceAll(subj, ".", "_") + "_" + consumerSuffix
					gConsumerMutex.Lock()
					if _, ok := gConsumerMap[key]; !ok && streamRegex.MatchString(name) && subjRegex.MatchString(subj) {
						// 找到满足模板的新主题，创建消费者
						logs.Importantf("RegisterMultiPullConsumer Find New Stream %s.%s", name, subj)
						err := RegisterNatsPullConsumer(name, subj, key, logic, option...)
						if err != nil {
							logs.Warnf("RegisterMultiPullConsumer Register %s err:%+v", key, err)
							continue
						}
						gConsumerMap[key] = true
					}
					gConsumerMutex.Unlock()
				}
			}
		}
	}()
}
