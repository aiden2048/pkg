package natsStream

import (
	"fmt"
	"time"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/frame/runtime"
	"github.com/aiden2048/pkg/public/errorMsg"
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
)

// SubEvent 消费者是进程名,调用PushjsEvent
func SubEvent[R any](streamName, sub string, logic func(req *R) *errorMsg.ErrRsp, option ...nats.SubOpt) error {
	PushjsEvent(streamName, sub, frame.GetServerName(), logic, option...)
	streamName = FixStreamName(streamName)
	return PushjsEvent(streamName, sub, frame.GetServerName(), logic, option...)
}

// // SubConfig 消费者是进程名_进程id,调用PushjsEvent
// func SubConfig[R any](streamName, sub string, logic func(req *R) *errorMsg.ErrRsp, option ...nats.SubOpt) error {
// 	return PushjsEvent(streamName, sub, fmt.Sprintf("%s_%d", frame.GetServerName(), frame.GetServerID()), logic, option...)
// }

// PushjsEvent 目前PushjsEvent 不按app_id区分
func PushjsEvent[R any](streamName, sub, consumer string, logic func(req *R) *errorMsg.ErrRsp, option ...nats.SubOpt) error {
	js := frame.GetNatsStreamContext()
	streamkey := frame.GetNatsStreamKey(streamName, sub)
	if err := frame.CreateNatsStream(streamkey); err != nil {
		logs.Errorf("Stream:%s subj:%s Publish Checksubect err:%+v", streamkey, sub, err)
		return err
	}

	consumer = consumer + "_groups"
	option = append(option, nats.ManualAck(), nats.DeliverNew())
	msgCh := make(chan *nats.Msg)
	for i := 0; i < 128; i++ {
		runtime.Go(func() {
			for {
				msg := <-msgCh
				func() {
					defer msg.Ack()
					start := time.Now()
					qgreq := frame.NatsTransMsg{}
					var req R
					err := jsoniter.Unmarshal(msg.Data, &qgreq)
					if err != nil {
						t := time.Since(start)
						frame.ReportDoRpcStat("Unmarshal", streamkey, frame.ESMR_FAILED, t)
						logs.Bill("Push_Nats_Fail", "streamName:%s, sub:%s, consumer:%s,data:%s,err:%+v", streamName, sub, consumer, string(msg.Data), err)
						return
					}
					err = jsoniter.Unmarshal(qgreq.MsgBody, &req)
					if err != nil {
						t := time.Since(start)
						frame.ReportDoRpcStat("Unmarshal", streamkey, frame.ESMR_FAILED, t)
						logs.Bill("Push_Nats_Fail", "streamName:%s, sub:%s, consumer:%s,data:%s,err:%+v", streamName, sub, consumer, string(msg.Data), err)
						return
					}
					var gpid int64
					ts := fmt.Sprintf("%02d%02d%02d", start.Hour(), start.Minute(), start.Second())
					if qgreq.Sess.Trace == nil {
						qgreq.Sess.Trace, gpid = logs.CreateTraceId(&qgreq.Sess)
						qgreq.Sess.Trace.TraceId += ":" + sub + ":" + ts
					} else {
						qgreq.GetSession().Trace.TraceId = qgreq.GetSession().Trace.TraceId +
							fmt.Sprintf(":%s%d:%s:%s", qgreq.GetSession().GetServerFE(), qgreq.GetSession().GetServerID(), sub, ts)
						gpid = logs.StoreTraceId(qgreq.Sess.Trace, &qgreq.Sess)
					}
					logs.PrintInfo("sub-event", streamkey, req, err)
					qgreq.NatsMsg = msg
					qgreq.Sess.RpcType = "nats"
					eventErr := logic(&req)
					ret := int32(0)
					if eventErr != nil {
						ret = 1
					}
					t := time.Since(start)
					s := fmt.Sprintf("%s.%d", qgreq.GetSession().SvrFE, qgreq.GetSession().SvrID)
					frame.ReportDoRpcStat(s, streamkey, ret, t)
					//用完移出traceid映射
					if gpid != 0 {
						logs.RemoveTraceId(gpid)
					}
				}()
			}
		})
	}
	_, err := js.QueueSubscribe(streamkey+".*", consumer, func(msg *nats.Msg) {
		msgCh <- msg
	}, option...)
	if err != nil {
		return err
	}
	return nil
}
