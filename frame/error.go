package frame

import "errors"

var (
	ErrMsgEncodeErr = errors.New("msg encode error") //
	ErrSeqRunOut    = errors.New("seq run out")      //
	ErrTimeOut      = errors.New("timeout")          //
	ErrCsHeadIsNil  = errors.New("cs head is nil")
)
