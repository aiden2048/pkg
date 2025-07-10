// seq.go 收发包seq管理
package frame

import (
	"sync/atomic"
)

const MAX_PROXY_SEQ_NUM = 0xffff + 1    // 65536
const RPC_SEQ_BEGIN = MAX_PROXY_SEQ_NUM // 65536
var rpcSequence uint64

func init() {
	// Proxy Seq范围从0 ~ 65535，预先分配

	// Rpc Seq从65536开始递增
	rpcSequence = RPC_SEQ_BEGIN
}

func NewSeq(protoType int32) (seq uint64, err error) {

	seq = atomic.AddUint64(&rpcSequence, 1)
	return seq, err
}

func FreeSeq(seq uint64) {

}
