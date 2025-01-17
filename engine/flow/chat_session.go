package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/gofrs/uuid"
	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/manager"
	"github.com/zone-7/chatflow_engine/engine/meta"
	"github.com/zone-7/chatflow_engine/engine/utils"
)

var Sessions = make(map[string]*ChatSession)

func init() {
	go monitSession()
}

type ChatSession struct {
	Info                 *meta.ChatSessionInfo   `json:"info"`
	Opt                  meta.Option             `json:"option"`
	Chatflow             *meta.ChatFlow          `json:"chatflow"`
	Runtime              *andflow.RuntimeModel   `json:"runtime"`
	Messages             []*meta.ChatFlowMessage `json:"messages"` //消息记录
	IsOpen               bool                    `json:"is_open"`
	Running              bool                    `json:"running"`
	ActiveTime           time.Time               `json:"active_time"`
	ResponseMessageTypes []string                `json:"response_message_types"`
	DoSuspand            bool                    `json:"do_suspand"`

	OutputFunc func(msg meta.ChatFlowMessage)

	input_chan chan meta.ChatFlowMessage
	store_chan chan string
	wg         sync.WaitGroup
}

// 打开会话
func OpenChatSession(opt meta.Option, message meta.ChatFlowMessage, rsesponseMessageTypes []string, output func(message meta.ChatFlowMessage)) (*ChatSession, error) {
	var err error

	if len(message.FlowCode) == 0 {
		return nil, errors.New("flow_code 参数不能为空")
	}

	if len(message.UserId) == 0 {
		uid, _ := uuid.NewV4()
		message.UserId = strings.ReplaceAll(uid.String(), "-", "")
	}
	if len(message.SessionId) == 0 {
		uid, _ := uuid.NewV4()
		message.SessionId = strings.ReplaceAll(uid.String(), "-", "")
	}
	if len(message.FlowSpace) == 0 {
		message.FlowSpace = meta.FLOW_SPACE_PRODUCT
	}

	flow_code := message.FlowCode
	flow_space := message.FlowSpace
	user_id := message.UserId
	session_id := message.SessionId

	if len(flow_space) == 0 {
		flow_space = meta.FLOW_SPACE_PRODUCT
	}

	chatSession := GetChatSession(session_id)
	if chatSession == nil {

		session_manager := manager.NewChatSessionInfoManager(opt)

		info, _ := session_manager.LoadSessionInfo(user_id, flow_code, session_id)
		if info == nil {
			info = &meta.ChatSessionInfo{}
			info.UserId = user_id
			info.FlowCode = flow_code
			info.FlowSpace = flow_space
			info.Id = session_id
		}

		runtime, _ := session_manager.LoadSessionRuntime(user_id, flow_code, session_id)
		history, _ := session_manager.LoadSessionMessages(user_id, flow_code, session_id, 0, 0)

		chatSession, err = CreateChatSession(opt, info, history, runtime, rsesponseMessageTypes)
		if err != nil {
			return chatSession, err
		}
	}

	chatSession.OutputFunc = output
	chatSession.Open()

	return chatSession, nil

}

// 获取会话
func GetChatSession(session_id string) *ChatSession {

	chatSession := Sessions[session_id]

	return chatSession

}

// 通过流程编码关闭所有会话
func CloseChatSessionsByFlowCode(flow_code string) {

	session_ids_del := make([]string, 0)
	for _, s := range Sessions {
		if s.Info.FlowCode == flow_code {
			session_ids_del = append(session_ids_del, s.Info.Id)
		}
	}
	for _, id := range session_ids_del {
		CloseChatSession(id)
	}
}

// 关闭会话
func CloseChatSession(session_id string) {
	session := Sessions[session_id]
	if session == nil {
		return
	}

	session.Close()

	delete(Sessions, session_id)
}

func CloseAllChatSession(user_id, flow_code string) {
	for _, session := range Sessions {
		if session.Info.UserId == user_id && session.Info.FlowCode == flow_code {
			CloseChatSession(session.Info.Id)
		}
	}
}

// 创建会话
func CreateChatSession(opt meta.Option, info *meta.ChatSessionInfo, history []*meta.ChatFlowMessage, runtime *andflow.RuntimeModel, responseMessageTypes []string) (*ChatSession, error) {
	if len(info.Id) == 0 {
		return nil, errors.New("会话ID不能为空")
	}
	if len(info.FlowCode) == 0 {
		return nil, errors.New("流程编码不能为空")
	}
	if len(info.UserId) == 0 {
		return nil, errors.New("用户ID不能为空")
	}
	flow_space := info.FlowSpace
	if len(flow_space) == 0 {
		flow_space = meta.FLOW_SPACE_PRODUCT
	}

	chatFlowManager := manager.ChatFlowManager{Opt: opt}

	chatflow, err := chatFlowManager.LoadChatFlow(flow_space, info.FlowCode)

	if err != nil {
		return nil, err
	}
	if chatflow == nil {
		return nil, errors.New("对话流程不存在")
	}

	flow := chatflow.FlowModel

	if runtime == nil {
		runtime = andflow.CreateRuntime(flow, nil)
	}

	session := &ChatSession{}
	session.Opt = opt
	session.Info = info
	session.Info.CreateTime = time.Now().UnixNano() / 1e6 //毫秒
	session.Chatflow = chatflow                           //流程信息
	session.Messages = history                            //历史消息

	session.Runtime = runtime            //运行状态
	session.Runtime.Id = info.Id         //ID 直接复制给运行时状态ID
	session.Runtime.UserId = info.UserId //用户ID复制给运行时状态的用户ID

	session.Runtime.Flow = flow //用新的流程复制给运行时状态的流程定义

	session.ResponseMessageTypes = responseMessageTypes //可以响应的消息类型列表

	session.ActiveTime = time.Now() //活动时间

	Sessions[info.Id] = session

	return session, nil
}

// 打开通道启动
func (s *ChatSession) Open() {
	if s.IsOpen {
		return
	}
	// 设置启动状态为true

	s.IsOpen = true
	s.ActiveTime = time.Now()

	s.wg = sync.WaitGroup{}

	if s.input_chan == nil {
		s.input_chan = make(chan meta.ChatFlowMessage, 5) //消息接收队列

	}

	if s.store_chan == nil {
		s.store_chan = make(chan string, 5) //存储会话队列

	}

	// 打开通道
	go s.input_process()
	go s.store_process()
}

// 关闭通道
func (s *ChatSession) Close() {

	//关闭通道
	close(s.input_chan)
	close(s.store_chan)

	s.wg = sync.WaitGroup{}
	//设置启动状态为false
	s.IsOpen = false

}

// 停止流程
func (s *ChatSession) Suspand() {
	fmt.Println("用户要求停止")
	flowSession := andflow.GetSession(s.Runtime.Id)
	if flowSession != nil {
		flowSession.Stop()
		s.DoSuspand = true
	}

}

// 重置运行状态
func (s *ChatSession) Reset() {
	if s.Chatflow == nil {
		return
	}

	s.Runtime = andflow.CreateRuntime(s.Chatflow.FlowModel, nil)
	s.Runtime.Id = s.Info.Id         //ID 直接复制给运行时状态ID
	s.Runtime.UserId = s.Info.UserId //用户ID复制给运行时状态的用户ID

	s.Messages = make([]*meta.ChatFlowMessage, 0)
}

// 获取历史消息
func (s *ChatSession) GetMessages() []*meta.ChatFlowMessage {
	return s.Messages
}

// 添加到历史消息
func (s *ChatSession) AddMessage(msg *meta.ChatFlowMessage) {
	if msg.MessageType == meta.CHAT_MESSAGE_TYPE_MESSAGE {
		if s.Messages == nil {
			s.Messages = make([]*meta.ChatFlowMessage, 0)
		}

		//合并消息
		msgs := s.GetMessage(msg.MessageId)
		var oldmsg *meta.ChatFlowMessage
		if msgs != nil && len(msgs) > 0 {
			oldmsg = msgs[0]
		}
		if oldmsg == nil {
			s.Messages = append(s.Messages, msg)
		} else {
			oldmsg.Content = oldmsg.Content + msg.Content
			oldmsg.Finish = msg.Finish
		}

	}
}

// 根据消息ID获取消息
func (s *ChatSession) GetMessage(messageId string) []*meta.ChatFlowMessage {
	msgs := make([]*meta.ChatFlowMessage, 0)
	if s.Messages != nil {
		for _, msg := range s.Messages {
			if msg.MessageId == messageId {
				msgs = append(msgs, msg)
			}
		}
	}
	return msgs
}

// 获取当前请求的消息
func (s *ChatSession) GetCurrentRequestMessages(history_count int) []*meta.ChatFlowMessage {

	requestId := s.Runtime.RequestId

	if history_count <= 0 {
		history_count = 1
	}

	msgs := make([]*meta.ChatFlowMessage, 0)
	if s.Messages != nil {
		for i := len(s.Messages) - 1; i >= 0 && len(msgs) < history_count; i-- {
			msg := s.Messages[i]

			if msg != nil && msg.MessageType == meta.CHAT_MESSAGE_TYPE_MESSAGE && msg.RequestId == requestId && msg.Role == meta.CHAT_MESSAGE_ROLE_USER {
				msgs = append([]*meta.ChatFlowMessage{msg}, msgs...)
			}
		}
	}

	return msgs
}

func (s *ChatSession) GetCurrentRequestMessagesContent(history_count int) string {
	requestContent_route := ""
	messages := s.GetCurrentRequestMessages(history_count)
	if messages != nil && len(messages) > 0 {
		for _, msg := range messages {
			requestContent_route += msg.Content
		}
	}
	return requestContent_route
}

// 获取当请求前返回消息
func (s *ChatSession) GetCurrentResponseMessages() []*meta.ChatFlowMessage {
	requestId := s.Runtime.RequestId

	msgs := make([]*meta.ChatFlowMessage, 0)
	if s.Messages != nil {
		for _, msg := range s.Messages {
			if msg.MessageType == meta.CHAT_MESSAGE_TYPE_MESSAGE && msg.RequestId == requestId && msg.Role == meta.CHAT_MESSAGE_ROLE_ASSISTANT {
				msgs = append(msgs, msg)
			}
		}
	}

	return msgs
}

// 执行
func (s *ChatSession) Execute(msg meta.ChatFlowMessage) error {
	defer func() {
		s.wg.Done()
	}()

	var err error
	if s.Info == nil {
		return nil
	}

	if s.Chatflow == nil {
		return nil
	}

	if s.Runtime == nil {
		return nil
	}

	//内容为空不执行
	if len(msg.Content) == 0 {
		return nil
	}

	//如果正在执行就不执行
	if s.Running {
		return nil
	}
	// 正在执行标志
	s.Running = true

	// 是否中断
	s.DoSuspand = false

	defer func() {
		s.Running = false
		if s.store_chan != nil {
			s.store_chan <- "store"
		}
	}()

	//过滤
	cando := s.doFilter("input", &msg)
	if !cando {
		return nil
	}

	// 补全Msg信息
	msg.FlowCode = s.Info.FlowCode
	msg.SessionId = s.Info.Id
	msg.UserId = s.Info.UserId
	msg.MessageType = meta.CHAT_MESSAGE_TYPE_MESSAGE

	if len(msg.Mid) == 0 {
		uid, _ := uuid.NewV4()
		mid := strings.ReplaceAll(uid.String(), "-", "")
		msg.Mid = mid
	}
	if len(msg.MessageId) == 0 {
		uid, _ := uuid.NewV4()
		mid := strings.ReplaceAll(uid.String(), "-", "")
		msg.MessageId = mid
	}

	if len(msg.RequestId) == 0 {
		uid, _ := uuid.NewV4()
		rid := strings.ReplaceAll(uid.String(), "-", "")
		msg.RequestId = rid
	}

	if len(msg.Role) == 0 {
		msg.Role = meta.CHAT_MESSAGE_ROLE_USER
	}

	if msg.SendTime == 0 {
		msg.SendTime = time.Now().UnixNano() / 1e6
	}

	// 会话标题，默认为发送内容的前面几个字
	if len(s.Info.Title) == 0 {
		s.Info.Title = msg.Content
		s.Info.Title = strings.ReplaceAll(s.Info.Title, "\n", "")
		//截取Title长度，避免过长，用循环的方式可以防止中文乱吗
		s.Info.Title = strings.ReplaceAll(s.Info.Title, "\n", "")
		titleWords := strings.Split(s.Info.Title, "")
		if len(titleWords) > 30 {
			s.Info.Title = strings.Join(titleWords[:30], "")
		}
	}

	// 添加到消息记录列表
	s.AddMessage(&msg)

	// 激活时间
	s.ActiveTime = time.Now()

	// runtime operation
	runtimeOperation := &andflow.CommonRuntimeOperation{}
	runtimeOperation.Init(s.Runtime)

	runtimeOperation.OnChangeFunc = func(event string, runtime *andflow.RuntimeModel) {
		s.ResponseRuntime()
	}
	//设置流程的运行时，requestid = 本次消息的requestId
	runtimeOperation.SetRequestId(msg.RequestId)

	//流程定义参数，不覆盖已有
	for _, p := range s.Chatflow.Params {
		runtimeOperation.SetParam(p.Name, p.Value)
	}

	//message params
	msg_json, err := json.Marshal(msg)
	if err == nil {
		msg_obj := make(map[string]interface{})
		err = json.Unmarshal(msg_json, &msg_obj)
		if err == nil {
			runtimeOperation.SetParam("message", msg_obj)
		}
	}

	//时间相关参数
	runtimeOperation.SetParam("year", time.Now().Year())
	runtimeOperation.SetParam("month", time.Now().Month())
	runtimeOperation.SetParam("weekday", time.Now().Weekday())
	runtimeOperation.SetParam("datetime", time.Now().Format("2006-01-02 15:04:05"))
	runtimeOperation.SetParam("date", time.Now().Format("2006-01-02"))

	//对话用户提交的参数，覆盖
	if msg.Params != nil {
		for k, v := range msg.Params {
			runtimeOperation.SetParam(k, v)
		}
	}

	//flowrouter
	flowRouter := &andflow.CommonFlowRouter{}

	//flowrunner
	flowRunner := &andflow.CommonFlowRunner{}

	flowRunner.SetActionScriptFunc(func(rts *goja.Runtime, session *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) {
		rts.Set("sendWaiting", func(call goja.FunctionCall) goja.Value {
			arg := call.Argument(0)

			msg_v := fmt.Sprintf("%v", arg.Export())

			s.ResponseWaitting(msg_v)

			return goja.NaN()
		})
		rts.Set("sendWaitting", func(call goja.FunctionCall) goja.Value {
			arg := call.Argument(0)

			arg_v := fmt.Sprintf("%v", arg.Export())

			s.ResponseWaitting(arg_v)

			return goja.NaN()
		})

		rts.Set("sendMessage", func(call goja.FunctionCall) goja.Value {
			arg := call.Argument(0)
			format := call.Argument(1)

			arg_v := fmt.Sprintf("%v", arg.Export())

			formatStr := ""
			if format == nil || format == goja.NaN() || format == goja.Null() || len(format.String()) == 0 {
				formatStr = meta.CHAT_MESSAGE_FORMAT_TEXT
			} else {
				formatStr = format.String()
			}

			s.Response(meta.ChatFlowMessage{MessageType: meta.CHAT_MESSAGE_TYPE_MESSAGE, Content: arg_v, Format: formatStr, Finish: "yes"}, true)

			return goja.NaN()
		})

		rts.Set("getMessage", func(call goja.FunctionCall) goja.Value {
			return rts.ToValue(msg.Content)
		})

	})

	flowRunner.SetLinkScriptFunc(func(rts *goja.Runtime, session *andflow.Session, param *andflow.LinkParam, state *andflow.LinkStateModel) {
		rts.Set("sendWaiting", func(call goja.FunctionCall) goja.Value {
			arg := call.Argument(0)

			arg_v := fmt.Sprintf("%v", arg.Export())

			s.Response(meta.ChatFlowMessage{MessageType: meta.CHAT_MESSAGE_TYPE_WAITING, Content: arg_v, Format: meta.CHAT_MESSAGE_FORMAT_TEXT, Finish: "yes"}, true)

			return goja.NaN()
		})
		rts.Set("sendWaitting", func(call goja.FunctionCall) goja.Value {
			arg := call.Argument(0)

			arg_v := fmt.Sprintf("%v", arg.Export())

			s.ResponseWaitting(arg_v)

			return goja.NaN()
		})

		rts.Set("sendMessage", func(call goja.FunctionCall) goja.Value {
			arg := call.Argument(0)
			format := call.Argument(1)

			arg_v := fmt.Sprintf("%v", arg.Export())

			formatStr := ""
			if format == nil || format == goja.NaN() || format == goja.Null() || len(format.String()) == 0 {
				formatStr = meta.CHAT_MESSAGE_FORMAT_TEXT
			} else {
				formatStr = format.String()
			}

			s.Response(meta.ChatFlowMessage{MessageType: meta.CHAT_MESSAGE_TYPE_MESSAGE, Content: arg_v, Format: formatStr, Finish: "yes"}, true)

			return goja.NaN()
		})

		rts.Set("getMessage", func(call goja.FunctionCall) goja.Value {

			return rts.ToValue(msg.Content)
		})
	})

	flowRunner.SetActionFailureFunc(func(session *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel, err error) {

		s.Response(meta.ChatFlowMessage{MessageType: meta.CHAT_MESSAGE_TYPE_ERROR, Code: 1, Content: err.Error(), Finish: "yes"}, true)
	})

	flowRunner.SetLinkFailureFunc(func(session *andflow.Session, param *andflow.LinkParam, state *andflow.LinkStateModel, err error) {
		s.Response(meta.ChatFlowMessage{MessageType: meta.CHAT_MESSAGE_TYPE_ERROR, Code: 1, Content: err.Error(), Finish: "yes"}, true)
	})

	flowRunner.SetActionExecutedFunc(func(session *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel, res andflow.Result, err error) {
		if state.ActionName == "begin" {
			session.Operation.SetState(1)
		}
		if state.ActionName == "end" {
			session.Operation.SetState(2)
		}
	})

	flowRunner.SetLinkExecutedFunc(func(session *andflow.Session, param *andflow.LinkParam, state *andflow.LinkStateModel, res andflow.Result, err error) {

	})

	flowRunner.SetTimeoutFunc(func(session *andflow.Session) {
		s.Response(meta.ChatFlowMessage{MessageType: meta.CHAT_MESSAGE_TYPE_ERROR, Code: 1, Content: "执行超时", Finish: "yes"}, true)
	})

	// 发送“请等待”提示信息
	if len(strings.Trim(s.Chatflow.WaittingText, " ")) > 0 {
		s.ResponseWaitting(s.Chatflow.WaittingText)
	}

	//超时设置
	var timeout int64
	if len(s.Chatflow.FlowModel.Timeout) >= 0 {
		i, err := strconv.ParseInt(s.Chatflow.FlowModel.Timeout, 10, 64)
		if err == nil {
			timeout = i
		}
	}
	// 默认超时时间（MS）
	if timeout <= 0 {
		timeout = 60000 //60秒
	}

	s.ResponseSession()

	// 执行andflow
	andflow.Execute(runtimeOperation, flowRouter, flowRunner, timeout)

	// complete
	s.ResponseComplete()

	//如果是人为中断就需要重置会话
	if s.DoSuspand {
		s.Reset()
	}

	return err

}

// 异步执行对话流程
func (s *ChatSession) ChatAsync(msg meta.ChatFlowMessage) {
	if len(msg.Content) == 0 {
		log.Println("content empty")
		return
	}

	if s.input_chan == nil {
		log.Println("input chan empty")
		return
	}

	//其他消息
	s.wg.Add(1)
	s.input_chan <- msg
}

// 同步执行对话流程
func (s *ChatSession) Chat(msg meta.ChatFlowMessage) {
	s.ChatAsync(msg)
	s.wg.Wait()
}

// 输出消息
func (s *ChatSession) Response(msg meta.ChatFlowMessage, add_history bool) {

	//过滤
	cando := s.doFilter("output", &msg)
	if !cando {
		return
	}

	s.ActiveTime = time.Now()

	uid, _ := uuid.NewV4()
	mid := strings.ReplaceAll(uid.String(), "-", "")
	msg.Mid = mid

	msg.FlowCode = s.Chatflow.FlowModel.Code
	msg.SessionId = s.Info.Id
	msg.UserId = s.Info.UserId
	msg.FlowCode = s.Info.FlowCode

	if len(msg.MessageId) == 0 {
		uid, _ := uuid.NewV4()
		messageId := strings.ReplaceAll(uid.String(), "-", "")
		msg.MessageId = messageId
	}
	if len(msg.RequestId) == 0 {
		msg.RequestId = s.Runtime.RequestId
	}
	if len(msg.MessageType) == 0 {
		msg.MessageType = meta.CHAT_MESSAGE_TYPE_MESSAGE
	}
	if len(msg.Role) == 0 {
		msg.Role = meta.CHAT_MESSAGE_ROLE_ASSISTANT
	}
	if msg.SendTime == 0 {
		msg.SendTime = time.Now().UnixNano() / 1e6
	}

	if len(msg.Finish) == 0 {
		msg.Finish = "yes"
	}

	//根据消息类型判断是否可以输出
	if len(msg.MessageType) > 0 && utils.StringsIndex(s.ResponseMessageTypes, msg.MessageType) < 0 {
		return
	}

	//添加到消息历史记录
	if add_history {
		s.AddMessage(&msg)
	}

	//输出到回调函数
	if s.OutputFunc != nil {
		s.OutputFunc(msg)
	}

}

// 反馈状态
func (s *ChatSession) ResponseRuntime() {

	runtime := s.Runtime
	if runtime == nil {
		return
	}
	runtime_res := andflow.RuntimeModel{}
	runtime_res.ActionStates = runtime.ActionStates
	runtime_res.LinkStates = runtime.LinkStates
	runtime_res.BeginTime = runtime.BeginTime
	runtime_res.EndTime = runtime.EndTime
	runtime_res.FlowState = runtime.FlowState
	runtime_res.Id = runtime.Id
	runtime_res.IsError = runtime.IsError
	runtime_res.IsRunning = runtime.IsRunning
	runtime_res.Param = runtime.Param
	runtime_res.Logs = runtime.Logs

	rt, err := json.Marshal(runtime_res)
	if err != nil {
		return
	}

	content := string(rt)

	s.Response(meta.ChatFlowMessage{RequestId: runtime.RequestId, MessageType: meta.CHAT_MESSAGE_TYPE_RUNTIME, Role: meta.CHAT_MESSAGE_ROLE_SYSTEM, Content: content, Finish: "yes"}, true)
}

func (s *ChatSession) ResponseSession() {

	content, err := json.Marshal(s.Info)
	if err != nil {
		return
	}

	s.Response(meta.ChatFlowMessage{RequestId: s.Runtime.RequestId, MessageType: meta.CHAT_MESSAGE_TYPE_SESSION, Role: meta.CHAT_MESSAGE_ROLE_SYSTEM, Content: string(content), Finish: "yes"}, true)

}

func (s *ChatSession) ResponseWaitting(msg string) {
	s.Response(meta.ChatFlowMessage{MessageType: meta.CHAT_MESSAGE_TYPE_WAITING, Role: meta.CHAT_MESSAGE_ROLE_SYSTEM, Content: msg, Format: meta.CHAT_MESSAGE_FORMAT_TEXT, Finish: "yes"}, true)
}

func (s *ChatSession) ResponseComplete() {
	s.Response(meta.ChatFlowMessage{MessageType: meta.CHAT_MESSAGE_TYPE_COMPLETE, Role: meta.CHAT_MESSAGE_ROLE_SYSTEM, Content: "", Format: meta.CHAT_MESSAGE_FORMAT_TEXT, Finish: "yes"}, true)
}

// 合规过滤
func (s *ChatSession) doFilter(side string, msg *meta.ChatFlowMessage) bool {
	if msg.MessageType != meta.CHAT_MESSAGE_TYPE_MESSAGE {
		return true
	}
	if s.Chatflow.Filters != nil && len(s.Chatflow.Filters) > 0 {
		for _, filter := range s.Chatflow.Filters {

			if strings.Index(filter.Side, side) < 0 {
				continue
			}

			regword := filter.Word
			//忽略大小写
			if filter.IgnoreCase == "false" || filter.IgnoreCase == "0" {
				regword = filter.Word
			} else {
				regword = "(?i)(" + regword + ")"
			}

			reg := regexp.MustCompile(regword)
			indexs := reg.FindStringIndex(msg.Content)

			if indexs != nil {
				//拒绝模式
				if filter.Mode == "reject" {
					return false
				}
				//替换模式
				if filter.Mode == "replace" {

					msg.Content = reg.ReplaceAllString(msg.Content, filter.ReplaceWord)

				}

			}
		}

	}

	return true
}

func (s *ChatSession) input_process() {
	for {
		msg, ok := <-s.input_chan
		if !ok {
			break
		}

		s.Execute(msg)
	}
	s.input_chan = nil
}

// 保存用户会话
func (s *ChatSession) store_process() {
	for {

		_, ok := <-s.store_chan
		if !ok {
			break
		}
		s.StoreSession()

	}
	s.store_chan = nil
}

// 保存用户会话
func (s *ChatSession) StoreSession() {
	session_manager := manager.ChatSessionInfoManager{Opt: s.Opt}
	session_manager.StoreSessionInfo(s.Info)
	session_manager.StoreSessionMessages(s.Info, s.Messages)
	session_manager.StoreSessionRuntime(s.Info, s.Runtime)
}

// 监控会话是否过期，过期就关闭
func monitSession() {
	for {
		for _, s := range Sessions {
			if s == nil || s.Chatflow.SessionTimeout == 0 {
				continue
			}

			if time.Now().Sub(s.ActiveTime).Milliseconds() > s.Chatflow.SessionTimeout {
				fmt.Println("session release: ", s.Info.Id)
				CloseChatSession(s.Info.Id)
			}
		}

		time.Sleep(5 * time.Second)

	}
}
