package meta

import (
	"github.com/zone-7/andflow_go/andflow"
)

const (
	FLOW_SPACE_DEVELOP  = "develop"  //设计
	FLOW_SPACE_TEMPLATE = "template" //模板
	FLOW_SPACE_PRODUCT  = "product"  //生产
)

const (
	FLOW_STATUS_STORE   = "store"
	FLOW_STATUS_PUBLISH = "publish"
)

const (
	FLOW_TYPE_APP    = "app"    //主应用流程
	FLOW_TYPE_WIDGET = "widget" //工具流程
)
const (
	FLOW_ACTIVE_ALL   = ""
	FLOW_ACTIVE_TRUE  = "true"
	FLOW_ACTIVE_FALSE = "false"
)

// 消息类型
const (
	CHAT_MESSAGE_TYPE_MESSAGE  = "message"
	CHAT_MESSAGE_TYPE_RUNTIME  = "runtime"
	CHAT_MESSAGE_TYPE_WAITING  = "waiting"
	CHAT_MESSAGE_TYPE_COMPLETE = "complete"
	CHAT_MESSAGE_TYPE_SESSION  = "session"
	CHAT_MESSAGE_TYPE_SYSTEM   = "system"
	CHAT_MESSAGE_TYPE_ERROR    = "error"

	CHAT_MESSAGE_ROLE_USER      = "user"
	CHAT_MESSAGE_ROLE_ASSISTANT = "assistant"
	CHAT_MESSAGE_ROLE_SYSTEM    = "system"

	CHAT_MESSAGE_FORMAT_TEXT = "text"
	CHAT_MESSAGE_FORMAT_JSON = "json"
)

// 流程过滤
type ChatFlowFilter struct {
	Word        string `json:"word"`         //词
	ReplaceWord string `json:"replace_word"` //替换词
	Side        string `json:"side"`         //过滤端： input 输入， output输出，input&output输入输出。
	Mode        string `json:"mode"`         //过滤模式：reject拒绝，replace关键词替换，
	IgnoreCase  string `json:"ignore_case"`  //是否忽略大小写
}

// 流程参数
type ChatFlowParam struct {
	Name        string `json:"name"`        //名称
	Label       string `json:"label"`       //标签
	Value       string `json:"value"`       //默认值
	Description string `json:"description"` //描述
	DebugValue  string `json:"debug_value"` //调试时使用的值
	InputType   string `json:"input_type"`  //text, number,password, select,textarea 输入框HTML
	Visible     string `json:"visible"`     //是否可见 true、false
}

// 用户的参数
type UserChatFlowParams struct {
	UserId   string           `json:"user_id"`
	FlowCode string           `json:"flow_code"`
	Params   []*ChatFlowParam `json:"params"`
}

// 流程信息
type ChatFlowInfo struct {
	Code           string `json:"code" yaml:"code"`
	Name           string `json:"name" yaml:"name"`
	SubTitle       string `json:"sub_title" yaml:"sub_title"`
	Version        string `json:"version" yaml:"version"`
	Edition        int64  `json:"edition" yaml:"edition"`
	Published      string `json:"published" yaml:"published"`
	Description    string `json:"description" yaml:"description"`
	Active         string `json:"active" yaml:"active"`
	SessionTimeout int64  `json:"session_timeout" yaml:"session_timeout"`
	WaittingText   string `json:"waitting_text" yaml:"waitting_text"`
	CreateTime     int64  `json:"create_time" yaml:"create_time"` //毫秒
	OrderNum       int64  `json:"order_num" yaml:"order_num"`
	Icon           string `json:"icon" yaml:"icon"`
	FlowSpace      string `json:"flow_space" yaml:"flow_space"`
	FlowType       string `json:"flow_type" yaml:"flow_type"`
}

// 用于对话的的信息
type ChatFlowMeta struct {
	ChatFlowInfo
	Params []*ChatFlowParam `json:"params"`
}

// 流程定义信息
type ChatFlow struct {
	ChatFlowInfo
	Params    []*ChatFlowParam   `json:"params"`  //参数
	Filters   []*ChatFlowFilter  `json:"filters"` //过滤器
	FlowModel *andflow.FlowModel `json:"flow_model"`
}

type ChatFlowMessage struct {
	Mid            string            `json:"mid"`             //唯一ID
	RequestId      string            `json:"request_Id"`      //每次请求一个ID
	FlowSpace      string            `json:"flow_space"`      //流程空间
	FlowType       string            `json:"flow_type"`       //流程类型
	FlowCode       string            `json:"flow_code"`       //流程ID
	SessionId      string            `json:"session_id"`      //会话ID
	UserId         string            `json:"user_id"`         //用户ID,最终用户
	MessageId      string            `json:"message_id"`      //每个消息一个ID，一个ID可以分几个多次发送
	MessageType    string            `json:"message_type"`    //消息类型
	MessageAccepts string            `json:"message_accepts"` //可接收的消息类型。
	Role           string            `json:"role"`            //发送者角色
	Params         map[string]string `json:"params"`          //参数
	Content        string            `json:"content"`         //消息内容
	Images         []string          `json:"images"`          //图片
	Format         string            `json:"format"`          //消息格式
	Stream         bool              `json:"stream"`          //流式输出
	Code           int               `json:"code"`            //异常编码
	Index          int               `json:"index"`           //消息序号
	SendTime       int64             `json:"send_time"`       //发送时间毫秒
	Finish         string            `json:"finish"`          //一个消息多次发送的话，表示是否结束,结束原因
}
