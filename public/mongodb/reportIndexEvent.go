package mongodb

import "github.com/aiden2048/pkg/frame"

const (
	ReportIndexMsgEvt       = "ReportIndexMsg"
	ReportIndexMsgEvtStream = "ReportIndexMsgStream"
)

type ReportIndexMsgEvent struct {
	DbName          string   `json:"db_name"`
	TbName          string   `json:"tb_name"`
	Version         int      `json:"version"`           // 版本号
	Expire          int      `json:"expire"`            // 过期时间
	IndexList       []string `json:"index_list"`        // 业务索引
	UniqueIndexList []string `json:"unique_index_list"` // 唯一索引
	ExpireKeyList   []string `json:"expire_key_list"`   // 过期索引
	SvrName         string   `json:"svr_name"`          // 上报进程
}

func (e *ReportIndexMsgEvent) Send() {
	frame.PublishSub(ReportIndexMsgEvtStream, ReportIndexMsgEvt, 0, e, frame.NewSessionOnly())
}
