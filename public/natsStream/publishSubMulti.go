package natsStream

import (
	"github.com/aiden2048/pkg/frame"
	"github.com/nats-io/nats.go"
)

func RegisterMultiPullConsumer(streamPattern, subjPattern, consumerSuffix string, logic func([]*nats.Msg) error, option ...*frame.ConsumeOption) {
	frame.RegisterNatsMultiPullConsumer(streamPattern, subjPattern, consumerSuffix, logic)
}
