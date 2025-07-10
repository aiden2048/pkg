package frame

const (
	ReportStatMsgEvt       = "ReportStatMsg"
	ReportStatMsgEvtStream = "ReportStatMsgStream"
)

type ReportStatMsgEvent struct {
	PlatId             int32  `json:"plat_id"`
	SvrName            string `json:"svr_name,omitempty"`
	SvrID              int32  `json:"svr_id,omitempty"`
	Host               string `json:"host"`
	Key                string `json:"key"`
	Type               int32  `json:"type"`             //0:累计统计, 每次输出后清零, 1: 重置型统计, 每次输出后不清零
	TotalMsgNum        int64  `json:"total_msg_num"`    //!<消息处理总数
	SuccessMsgNum      int32  `json:"success_msg_num"`  //!<消息处理成功的个数
	FailMsgNum         int32  `json:"fail_msg_num"`     //!<消息处理失败的个数
	TimeoutMsgNum      int32  `json:"timeout_msg_num"`  //!<消息处理超时的个数
	MaxProcessTime     int64  `json:"max_process_time"` //!<最大处理耗时
	AvgProcessTime     int64  `json:"avg_process_time"`
	AvgSuccProcessTime int64  `json:"avg_succ_process_time"`
}

func (e *ReportStatMsgEvent) Send(sess ...*Session) {
	if len(sess) > 0 {
		PublishSub(ReportStatMsgEvtStream, ReportStatMsgEvt, 0, e, sess[0])
	} else {
		PublishSub(ReportStatMsgEvtStream, ReportStatMsgEvt, 0, e, NewSessionOnly())
	}
}
