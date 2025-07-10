package frame

const (
	ReportLogMsgEvt       = "ReportLogMsg"
	ReportLogMsgEvtStream = "ReportLogMsgStream"
)

type ReportLogMsgEvent struct {
	SvrName    string `json:"svr_name,omitempty"`
	SvrID      int32  `json:"svr_id,omitempty"`
	ServerHost string `json:"server_host"`
	Pid        int    `json:"pid"`
	LogType    string `json:"log_type"`
	LogTag     string `json:"log_tag"`
	LogMsg     string `json:"log_msg"`
}

func (e *ReportLogMsgEvent) Send(sess ...*Session) {
	if len(sess) > 0 {
		PublishSub(ReportLogMsgEvtStream, ReportLogMsgEvt, 0, e, sess[0])
	} else {
		PublishSub(ReportLogMsgEvtStream, ReportLogMsgEvt, 0, e, NewSessionOnly())
	}
}
