package flow

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"

	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/meta"
)

const (
	EXTRACT_TYPE_FULL   = "full"
	EXTRACT_TYPE_FLOW   = "flow"
	EXTRACT_TYPE_FORMAT = "format"
)

func init() {
	andflow.RegistActionRunner("params", &ParamRunner{})
}

type ParamItem struct {
	Id            int64  `json:"id" orm:"pk;auto"`
	Ask           string `json:"ask" orm:"size(100);null"`            //提问表达
	Name          string `json:"name" orm:"size(50);null"`            //名称 英文
	Label         string `json:"label" orm:"size(50);null"`           //中文名称
	InputType     string `json:"input_type" orm:"size(20);null"`      //输入类型
	DataType      string `json:"data_type" orm:"size(20);null"`       //数据类型
	ExtractType   string `json:"extract_type" orm:"size(20);null"`    //实体提取方式
	ExtractFlow   string `json:"extract_flow" orm:"size(200);null"`   //实体提取流程
	ExtractFormat string `json:"extract_format" orm:"size(200);null"` //实体提取正则表达式
	Options       string `json:"options" orm:"size(200);null"`        //选项
	Scope         string `json:"scope" orm:"size(200);null"`          //取值范围
	OrderNo       int    `json:"order_no" orm:"default(0)"`           //顺序
}

type ParamRunner struct {
	BaseRunner
}

func (r *ParamRunner) Properties() []andflow.Prop {
	return []andflow.Prop{}
}
func (r *ParamRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {
	var err error
	action := s.GetFlow().GetAction(param.ActionId)
	chatSession := r.GetChatSession(s)

	prop, err := r.GetActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	// 提示词
	param_ask := prop["param_ask"]

	// 返回
	param_back_links := prop["param_back_links"]
	param_back_words := prop["param_back_words"]

	//是否需要最后确认
	param_check := prop["param_check"]
	//确认词
	param_check_words_yes := prop["param_check_words_yes"]
	param_check_words_no := prop["param_check_words_no"]
	param_check_links := prop["param_check_links"]

	param_source := prop["param_source"]           //参数信息来自参数还是输入
	param_source_temp := prop["param_source_temp"] //参数信息来自哪个参数

	paramsJson := prop["params"] //需要填充的参数
	//需要填充的参数
	params := make([]*ParamItem, 0)
	if len(paramsJson) > 0 {
		err = json.Unmarshal([]byte(paramsJson), &params)
	}

	// 获取消息
	requestContent_param := ""
	if param_source == "temp" && len(param_source_temp) > 0 {
		requestContent_param = param_source_temp
	} else {
		requestContent_param = chatSession.GetCurrentRequestMessagesContent(1)
	}

	// 判断返回
	is_back := false
	words_back := getWords(param_back_words)
	for _, bw := range words_back {
		if strings.Index(requestContent_param, bw) >= 0 {
			is_back = true
			break
		}
	}

	if is_back {
		// 返回节点
		backActions := r.GetNextActionsByName(s, action, param_back_links)
		state.NextActionIds = make([]string, 0)
		for _, act := range backActions {
			state.NextActionIds = append(state.NextActionIds, act.Id)
		}

		r.clearStates(s, action)
		r.clearAnswers(s, action, params)
		return andflow.RESULT_SUCCESS, err

	}

	/*参数*/

	//没有消息就发送提示词
	if r.isFirstAsk(s, action, params) {
		chatSession.Response(meta.ChatFlowMessage{Content: param_ask, MessageType: meta.CHAT_MESSAGE_TYPE_MESSAGE}, true)
		action.SetParam("param_first_ask_"+action.Id, "false")
	}

	//正在提问，由用户提供的实体，用户答案
	asking_param_name := action.GetParam("asking_param_name_" + action.Id)
	if len(asking_param_name) > 0 {
		r.fillAnswers(s, action, params, asking_param_name, requestContent_param)
	}

	//根据参数提出疑问
	if r.askParam(s, action, chatSession, params) {
		return andflow.RESULT_REJECT, nil
	}

	// 如果需要最终确认
	if param_check == "true" {

		//如果正在提问确认
		asking_param_check := action.GetParam("asking_param_check_" + action.Id)

		if asking_param_check == "true" {
			action.SetParam("asking_param_check_"+action.Id, "")

			//处理关键词
			words_yes := getWords(param_check_words_yes)
			words_no := getWords(param_check_words_no)

			//判断是否通过
			var isNo bool
			var isOk bool
			for _, wn := range words_no {
				if strings.Index(requestContent_param, wn) >= 0 {
					isNo = true
					break
				}
			}
			for _, wy := range words_yes {
				if strings.Index(requestContent_param, wy) >= 0 {
					isOk = true
					break
				}
			}

			//不通过：清空答案
			if isNo {
				action.SetParam("param_checked_"+action.Id, "false")

				//回答不正确就删除所有用户回答的参数
				r.clearAnswers(s, action, params)

				//清空状态
				r.clearStates(s, action)

				//没有消息就发送提示词
				if r.isFirstAsk(s, action, params) {
					chatSession.Response(meta.ChatFlowMessage{Content: param_ask, MessageType: meta.CHAT_MESSAGE_TYPE_MESSAGE}, true)
					action.SetParam("param_first_ask_"+action.Id, "false")
				}

				//根据参数提出疑问
				r.askParam(s, action, chatSession, params)

				return andflow.RESULT_REJECT, nil
			}
			//通过：设置状态
			if isOk {
				action.SetParam("param_checked_"+action.Id, "true")
			}

		}

		param_checked := action.GetParam("param_checked_" + action.Id)
		//如果还没有回答就提示用户确认
		if param_checked == "" || param_checked == "false" {
			param_check_ask := r.GetActionParam(action, "param_check_ask", s.GetParamMap())

			if len(param_check_ask) > 0 {

				chatSession.Response(meta.ChatFlowMessage{Content: param_check_ask, MessageType: meta.CHAT_MESSAGE_TYPE_MESSAGE}, true)

				action.SetParam("asking_param_check_"+action.Id, "true")

			}
			return andflow.RESULT_REJECT, nil
		}

	}

	// 后续节点
	nextActions := r.GetNextActionsByName(s, action, param_check_links)
	state.NextActionIds = make([]string, 0)
	for _, act := range nextActions {
		state.NextActionIds = append(state.NextActionIds, act.Id)
	}

	//清空状态
	r.clearStates(s, action)

	return andflow.RESULT_SUCCESS, err
}

// 根据卡槽问题提出疑问
func (r *ParamRunner) askParam(s *andflow.Session, action *andflow.ActionModel, imsession *ChatSession, params []*ParamItem) bool {
	//判断哪些卡槽还没填，反问
	for _, p := range params {
		ps := s.GetParamMap()
		log.Println(ps)

		if s.GetParam(p.Name) == nil {
			action.SetParam("asking_param_name_"+action.Id, p.Name)

			//反问
			imsession.Response(meta.ChatFlowMessage{Content: p.Ask, MessageType: meta.CHAT_MESSAGE_TYPE_MESSAGE}, true)

			return true
		}
	}

	return false
}

// 清空状态
func (r *ParamRunner) clearStates(s *andflow.Session, action *andflow.ActionModel) {
	action.SetParam("asking_param_name_"+action.Id, "")
	action.SetParam("asking_param_check_"+action.Id, "")
	action.SetParam("param_checked_"+action.Id, "")
	action.SetParam("param_first_ask_"+action.Id, "")
}

func (r *ParamRunner) isFirstAsk(s *andflow.Session, action *andflow.ActionModel, params []*ParamItem) bool {

	f := action.GetParam("param_first_ask_" + action.Id)
	if f == "" || f == "true" {
		return true
	}
	return false
}

// 删除所有用户参数
func (r *ParamRunner) clearAnswers(s *andflow.Session, action *andflow.ActionModel, params []*ParamItem) {
	for _, p := range params {
		name := p.Name
		s.SetParam(name, nil)
	}
}

// 解析用户提问或者回答内容，答案提取
func (r *ParamRunner) fillAnswers(s *andflow.Session, action *andflow.ActionModel, params []*ParamItem, param_name string, messageContent string) {
	chatSession := r.GetChatSession(s)
	opt := chatSession.Opt

	for _, p := range params {
		name := p.Name
		if param_name != name {
			continue
		}

		if p.ExtractType == EXTRACT_TYPE_FULL { //全部提取
			s.SetParam(name, messageContent)
		} else if p.ExtractType == EXTRACT_TYPE_FORMAT && len(p.ExtractFormat) > 0 { //使用正则表达式提取

			reg := regexp.MustCompile(p.ExtractFormat)
			data := reg.Find([]byte(messageContent))
			if data != nil {
				s.SetParam(name, string(data))
			}
		} else if p.ExtractType == EXTRACT_TYPE_FLOW && len(p.ExtractFlow) > 0 { //调用模型服务提取实体信息

			msg := meta.ChatFlowMessage{}
			msg.Content = messageContent
			msg.FlowCode = p.ExtractFlow
			msg.FlowSpace = meta.FLOW_SPACE_PRODUCT

			subChatSession, err := OpenChatSession(opt, msg, []string{meta.CHAT_MESSAGE_TYPE_MESSAGE}, nil)
			if err != nil {
				continue
			}

			subChatSession.Chat(msg)
			responses := subChatSession.GetCurrentResponseMessages()
			data := ""
			for _, m := range responses {
				data += m.Content
			}
			if len(data) > 0 {
				s.SetParam(name, data)
			}

		}

	}

}
