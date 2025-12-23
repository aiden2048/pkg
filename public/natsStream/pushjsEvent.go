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

// PushjsEvent 目前PushjsEvent 不按app_id区分
func PushjsEvent[R any](
	streamName string,
	sub string,
	consumer string,
	logic func(req *R) *errorMsg.ErrRsp,
	option ...nats.SubOpt,
) error {

	streams := frame.GetNatsStreamContext()

	streamkey := frame.GetNatsStreamKey(streamName, sub)
	if err := frame.CreateNatsStream(streamkey); err != nil {
		logs.Errorf(
			"Stream:%s subj:%s Publish Checksubect err:%+v",
			streamkey, sub, err,
		)
		return err
	}

	consumer = consumer + "_groups"

	option = append(option,
		nats.ManualAck(),
		nats.DeliverNew(),
	)

	// ★ 必须是有缓冲的 channel，避免阻塞 NATS 回调
	msgCh := make(chan *nats.Msg, 1024)

	workerNum := 128

	for i := 0; i < workerNum; i++ {
		func(
			streamKey string,
			streamName string,
			sub string,
			consumer string,
			logic func(req *R) *errorMsg.ErrRsp,
		) {
			runtime.Go(func() {
				for {
					msg, ok := <-msgCh
					if !ok {
						return
					}

					func(msg *nats.Msg) {
						defer msg.Ack()

						start := time.Now()

						var (
							qgreq frame.NatsTransMsg
							req   R
						)

						if err := jsoniter.Unmarshal(msg.Data, &qgreq); err != nil {
							t := time.Since(start)
							frame.ReportDoRpcStat("Unmarshal", streamKey, frame.ESMR_FAILED, t)
							logs.Bill(
								"Push_Nats_Fail",
								"streamName:%s, sub:%s, consumer:%s, data:%s, err:%+v",
								streamName, sub, consumer, string(msg.Data), err,
							)
							return
						}

						if err := jsoniter.Unmarshal(qgreq.MsgBody, &req); err != nil {
							t := time.Since(start)
							frame.ReportDoRpcStat("Unmarshal", streamKey, frame.ESMR_FAILED, t)
							logs.Bill(
								"Push_Nats_Fail",
								"streamName:%s, sub:%s, consumer:%s, data:%s, err:%+v",
								streamName, sub, consumer, string(msg.Data), err,
							)
							return
						}

						var gpid int64
						ts := fmt.Sprintf("%02d%02d%02d",
							start.Hour(),
							start.Minute(),
							start.Second(),
						)

						if qgreq.Sess.Trace == nil {
							qgreq.Sess.Trace, gpid = logs.CreateTraceId(&qgreq.Sess)
							qgreq.Sess.Trace.TraceId += ":" + sub + ":" + ts
						} else {
							qgreq.GetSession().Trace.TraceId =
								qgreq.GetSession().Trace.TraceId +
									fmt.Sprintf(
										":%s%d:%s:%s",
										qgreq.GetSession().GetServerFE(),
										qgreq.GetSession().GetServerID(),
										sub,
										ts,
									)
							gpid = logs.StoreTraceId(qgreq.Sess.Trace, &qgreq.Sess)
						}

						qgreq.NatsMsg = msg
						qgreq.Sess.RpcType = "nats"

						eventErr := logic(&req)

						ret := int32(0)
						if eventErr != nil {
							ret = 1
						}

						t := time.Since(start)
						s := fmt.Sprintf(
							"%s.%d",
							qgreq.GetSession().SvrFE,
							qgreq.GetSession().SvrID,
						)

						frame.ReportDoRpcStat(s, streamKey, ret, t)

						if gpid != 0 {
							logs.RemoveTraceId(gpid)
						}
					}(msg)
				}
			})
		}(
			streamkey,
			streamName,
			sub,
			consumer,
			logic,
		)
	}

	// ★ 对所有 JetStream Context 建立 QueueSubscribe
	for _, js := range streams {
		_, err := js.QueueSubscribe(
			streamkey+".*",
			consumer,
			func(msg *nats.Msg) {
				msgCh <- msg
			},
			option...,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
