// channel.go 中转层，框架层与底层交互的管道，负责连接池管理、打包发送、回包转发
package frame

import (
	"time"

	"github.com/aiden2048/pkg/frame/logs"
	"github.com/aiden2048/pkg/utils/baselib"
)

type Channel struct {
	Uid uint64

	Host     string
	Uri      string
	Sequence uint64
	respChan chan *NatsTransMsg
	closed   chan struct{}
	Start    time.Time
	SendNum  int
	LastSend time.Time
}

var channelMap = new(baselib.Map)

func NewChannel(sess *Session) (*Channel, error) {
	ch := &Channel{respChan: make(chan *NatsTransMsg, 256) /*session: session,*/, closed: make(chan struct{})}
	var err error
	ch.Sequence, err = NewSeq(0)
	if err != nil {
		return nil, err
	}
	ch.Uid = sess.GetUid()

	ch.Start = time.Now()
	channelMap.Store(ch.Sequence, ch)
	return ch, nil
}

func PushChannelMsg(channel uint64, m *NatsTransMsg) {
	if v, ok := channelMap.Load(channel); ok {
		if ch, ok := v.(*Channel); ok {
			select {
			case ch.respChan <- m:
			default:
				logs.LogUserError(ch.GetUid(), "ch.respChan full, seq %d, uid:%d", channel, ch.GetUid())
			}
		} else {
			logs.Errorf("Channel Map Error %d", channel)
		}
	} else {
		logs.LogDebug("Seq Error %d", channel)
	}
}
func (ch *Channel) GetUid() uint64 {
	return ch.Uid
}
func (ch *Channel) GetSeq() uint64 {
	return ch.Sequence
}

func (ch *Channel) GetStart() time.Time {
	return ch.Start
}

func (ch *Channel) Close() {
	//先从map删除, 再关闭, 免得有人找到了
	channelMap.Delete(ch.Sequence)
	close(ch.closed)
	FreeSeq(ch.Sequence)

}

func (ch *Channel) IsClosed() <-chan struct{} {
	return ch.closed
}

func (ch *Channel) RecvMsgPara(timeout int) (*NatsTransMsg, error) {
	recvMsg, err := ch.Recv(timeout)
	if err != nil {
		return nil, err
	}
	return recvMsg, err
}
func (ch *Channel) Recv(timeout int) (m *NatsTransMsg, err error) {
	timer := GetTimer(time.Second * time.Duration(timeout))
	defer PutTimer(timer)
	select {
	case m = <-ch.respChan:
		return m, nil
	case <-timer.C:
		return m, ErrTimeOut
	}
}

func (ch *Channel) PushMsg(m *NatsTransMsg) {
	ch.respChan <- m
}

func (ch *Channel) RecvChan() <-chan *NatsTransMsg {
	return ch.respChan
}
