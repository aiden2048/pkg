package natsStream

import (
	"fmt"
	"sync"
	"time"

	"github.com/aiden2048/pkg/frame"
	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/frame/runtime"
	"github.com/aiden2048/pkg/frame/stat"
	"github.com/aiden2048/pkg/public/errorMsg"
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
)

var appIds []int32
var streamMap = sync.Map{}

// SubPullEvent 消费者是进程名,SubPullEvent
func SubPullEvent[R any](streamName, sub string, logic func(req []*R) *errorMsg.ErrRsp, option ...*frame.ConsumeOption) error {
	streamMap.Store(frame.GetNatsStreamKey(streamName, sub), true)
	return PulljsEvent(streamName, sub, 0, frame.GetServerName(), logic, option...)
}

// // SubConfig 消费者是进程名_进程id,调用PushjsEvent
// func SubConfig[R any](streamName, sub string, logic func(req *R) *errorMsg.ErrRsp, option ...nats.SubOpt) error {
// 	return PushjsEvent(streamName, sub, fmt.Sprintf("%s_%d", frame.GetServerName(), frame.GetServerID()), logic, option...)
// }

// PulljsEvent 按app_id
func FixStreamName(streamName string) string {
	return fmt.Sprintf("%d_%s", frame.GetPlatformId(), streamName)
}
func PulljsEvent[R any](streamName, sub string, app_id int32, consumer string, logic func(reqs []*R) *errorMsg.ErrRsp, option ...*frame.ConsumeOption) error {
	PulljsEventEx(streamName, sub, app_id, consumer, logic)
	streamName = FixStreamName(streamName)
	return PulljsEventEx(streamName, sub, app_id, consumer, logic)
}

func PulljsEventEx[R any](
	streamName string,
	sub string,
	app_id int32,
	consumer string,
	logic func(reqs []*R) *errorMsg.ErrRsp,
	option ...*frame.ConsumeOption,
) error {

	streams := frame.GetNatsStreamContext()

	streamkey := frame.GetNatsStreamKey(streamName, sub)
	streamkey = fmt.Sprintf("%s.%d", streamkey, app_id)

	if err := frame.CreateNatsStream(streamkey); err != nil {
		logs.Errorf(
			"Stream:%s subj:%s Publish Checksubect err:%+v",
			streamkey, sub, err,
		)
		return err
	}

	op := &frame.ConsumeOption{}
	if len(option) > 0 && option[0] != nil {
		op = option[0]
	}
	if op.BatchSize <= 0 {
		op.BatchSize = 10000
	}
	if op.MaxWaitMsec <= 0 {
		op.MaxWaitMsec = 10000
	}

	for _, js := range streams {
		subs, err := js.PullSubscribe(streamkey, consumer)
		if err != nil {
			return err
		}

		// 把所有外部变量一次性隔离
		func(
			subs *nats.Subscription,
			streamKey string,
			streamName string,
			sub string,
			consumer string,
			op frame.ConsumeOption,
			logic func(reqs []*R) *errorMsg.ErrRsp,
		) {
			runtime.Go(func() {
				defer subs.Unsubscribe()

				for {
					start := time.Now()

					msgs, err := subs.Fetch(
						op.BatchSize,
						nats.MaxWait(time.Duration(op.MaxWaitMsec)*time.Millisecond),
					)
					if err != nil {
						t := time.Since(start)
						frame.ReportDoRpcStat("PullNats", streamKey, frame.ESMR_FAILED, t)
						logs.Bill(
							"Pull_Nats_Fail",
							"streamName:%s, sub:%s, consumer:%s, err:%+v",
							streamName, sub, consumer, err,
						)
						continue
					}

					var reqs []*R
					for _, msg := range msgs {
						var (
							qgreq frame.NatsTransMsg
							req   R
						)

						if err := jsoniter.Unmarshal(msg.Data, &qgreq); err != nil {
							t := time.Since(start)
							frame.ReportDoRpcStat("Unmarshal", streamKey, frame.ESMR_FAILED, t)
							logs.Bill(
								"Pull_Nats_Fail",
								"streamName:%s, sub:%s, consumer:%s, data:%s, err:%+v",
								streamName, sub, consumer, string(msg.Data), err,
							)
							continue
						}

						if err := jsoniter.Unmarshal(qgreq.MsgBody, &req); err != nil {
							t := time.Since(start)
							frame.ReportDoRpcStat("Unmarshal", streamKey, frame.ESMR_FAILED, t)
							logs.Bill(
								"Pull_Nats_Fail",
								"streamName:%s, sub:%s, consumer:%s, data:%s, err:%+v",
								streamName, sub, consumer, string(msg.Data), err,
							)
							continue
						}

						reqs = append(reqs, &req)
					}

					if err := logic(reqs); err != nil {
						logs.Bill(
							"Pull_Nats_Fail",
							"streamName:%s, sub:%s, consumer:%s, data:%+v, err:%+v",
							streamName, sub, consumer, reqs, err,
						)
						continue
					}

					for _, msg := range msgs {
						doTime := time.Now()
						if pubTime := msg.Header.Get("PublishTime"); pubTime != "" {
							if pt, err := time.Parse(time.RFC3339Nano, pubTime); err == nil {
								stat.ReportStat(
									"Nats.PullStream."+streamName+"."+sub,
									0,
									doTime.Sub(pt),
								)
							}
						}
						msg.Ack()
					}

					costMsec := time.Since(start).Milliseconds()
					if 2*len(msgs) < op.BatchSize && costMsec < int64(op.MaxWaitMsec) {
						time.Sleep(
							time.Duration(int64(op.MaxWaitMsec)-costMsec) *
								time.Millisecond,
						)
					}
				}
			})
		}(
			subs,
			streamkey,
			streamName,
			sub,
			consumer,
			*op, // ★ 拷贝值，避免共享修改
			logic,
		)
	}

	return nil
}
