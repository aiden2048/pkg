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
func PulljsEventEx[R any](streamName, sub string, app_id int32, consumer string, logic func(reqs []*R) *errorMsg.ErrRsp, option ...*frame.ConsumeOption) error {
	js := frame.GetNatsStreamContext()
	streamkey := frame.GetNatsStreamKey(streamName, sub)
	streamkey = fmt.Sprintf("%s.%d", streamkey, app_id)
	if err := frame.CreateNatsStream(streamkey); err != nil {
		logs.Errorf("Stream:%s subj:%s Publish Checksubect err:%+v", streamkey, sub, err)
		return err
	}
	op := &frame.ConsumeOption{}
	if len(option) > 0 {
		op = option[0]
	}
	if op.BatchSize <= 0 {
		op.BatchSize = 10000
	}
	if op.MaxWaitMsec <= 0 {
		op.MaxWaitMsec = 10000
	}
	subs, err := js.PullSubscribe(streamkey, consumer)
	if err != nil {
		return err
	}
	runtime.Go(func() {
		defer subs.Unsubscribe()
		for {
			start := time.Now()
			msgs, err := subs.Fetch(op.BatchSize, nats.MaxWait(time.Duration(op.MaxWaitMsec)*time.Millisecond))
			if err != nil {
				t := time.Since(start)
				frame.ReportDoRpcStat("PullNats", streamkey, frame.ESMR_FAILED, t)
				logs.Bill("Pull_Nats_Fail", "streamName:%s, sub:%s, consumer:%s,err:%+v", streamName, sub, consumer, err)
			}
			var reqs []*R
			for _, msg := range msgs {
				qgreq := frame.NatsTransMsg{}
				var req R
				err := jsoniter.Unmarshal(msg.Data, &qgreq)
				if err != nil {
					t := time.Since(start)
					frame.ReportDoRpcStat("Unmarshal", streamkey, frame.ESMR_FAILED, t)
					logs.Bill("Pull_Nats_Fail", "streamName:%s, sub:%s, consumer:%s,data:%s,err:%+v", streamName, sub, consumer, string(msg.Data), err)
					continue
				}
				err = jsoniter.Unmarshal(qgreq.MsgBody, &req)
				if err != nil {
					t := time.Since(start)
					frame.ReportDoRpcStat("Unmarshal", streamkey, frame.ESMR_FAILED, t)
					logs.Bill("Pull_Nats_Fail", "streamName:%s, sub:%s, consumer:%s,data:%s,err:%+v", streamName, sub, consumer, string(msg.Data), err)
					continue
				}
				reqs = append(reqs, &req)
			}
			eventErr := logic(reqs)
			if eventErr != nil {
				logs.Bill("Pull_Nats_Fail", "streamName:%s, sub:%s, consumer:%s,data:%s,err:%+v", streamName, sub, consumer, reqs, eventErr)
				continue
			}
			for _, msg := range msgs {
				doTime := time.Now()
				pubTime := msg.Header.Get("PublishTime")
				if pubTime != "" {
					pt, errs := time.Parse(time.RFC3339Nano, pubTime)
					if errs == nil {
						stat.ReportStat("Nats.PullStream."+streamName+"."+sub, 0, doTime.Sub(pt))
					}
				}
				msg.Ack()
			}
			costMsec := time.Since(start).Milliseconds()
			if 2*len(msgs) < op.BatchSize && costMsec < int64(op.MaxWaitMsec) {
				// 如果取到数据较少就等待
				time.Sleep(time.Duration(int64(op.MaxWaitMsec)-costMsec) * time.Millisecond)
			}
		}
	})
	return nil
}
