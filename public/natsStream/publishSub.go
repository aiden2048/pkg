package natsStream

import (
	"github.com/aiden2048/pkg/frame"
	"github.com/nats-io/nats.go"
)

// PublishSub 按app_id推送就设置app_id 不分开推送设置 app_id=0
func PublishSub(streamName, sub string, app_id int32, v any, sess *frame.Session) error {
	return frame.PublishSub(streamName, sub, app_id, v, sess)
}

func Publish(streamName, streamSubj string, v any) error {
	return frame.Publish(streamName, streamSubj, v)
}

func RegisterPullConsumer(streamName, streamSubj, consumer string, logic func([]*nats.Msg) error, option ...*frame.ConsumeOption) error {
	return frame.RegisterNatsPullConsumer(streamName, streamSubj, consumer, logic)
}
