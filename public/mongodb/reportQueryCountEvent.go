package mongodb

import "github.com/aiden2048/pkg/frame"

const (
	ReportQueryCountMsgEvt       = "ReportQueryCountMsg"
	ReportQueryCountMsgEvtStream = "ReportQueryCountMsgStream"
)

type ReportQueryCountMsgEvent struct {
	DbName   string `json:"db_name"`
	TbName   string `json:"tb_name"`
	QueryKey string `json:"query_key"` // 查询key
	Count    int64  `json:"count"`     // 查询次数
	SvrName  string `json:"svr_name"`  // 上报进程
}

func (e *ReportQueryCountMsgEvent) Send() {
	frame.PublishSub(ReportQueryCountMsgEvtStream, ReportQueryCountMsgEvt, 0, e, frame.NewSessionOnly())
}
