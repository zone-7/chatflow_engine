package meta

// 状态信息

type ChatSessionInfo struct {
	Id         string `json:"id"`
	UserId     string `json:"user_id"`
	FlowCode   string `json:"flow_code"`
	FlowSpace  string `json:"flow_space"`
	Title      string `json:"title"`
	CreateTime int64  `json:"create_time"` //毫秒
}
