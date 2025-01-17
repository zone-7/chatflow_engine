package flow

// 意图识别
import (
	"strings"

	"github.com/gofrs/uuid"
	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/meta"
)

func init() {
	andflow.RegistActionRunner("route", &RouteRunner{})
}

type RouteRunner struct {
	BaseRunner
}

func (r *RouteRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {
	var err error

	action := s.GetFlow().GetAction(param.ActionId)
	chatSession := r.getChatSession(s)

	prop, err := r.getActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	route_source := prop["route_source"]
	route_source_temp := prop["route_source_temp"]
	route_method := prop["route_method"]
	route_flow := prop["route_flow"]
	route_ask := prop["route_ask"]
	route_reject_words := prop["route_reject_words"]

	// 获取消息文本
	requestContent_route := ""
	if route_source == "temp" && len(route_source_temp) > 0 {
		requestContent_route = route_source_temp
	} else {
		requestContent_route = chatSession.GetCurrentRequestMessagesContent(1)
	}

	//如果出现限制词，就清空消息，当作没说话
	reject_words := getWords(route_reject_words)
	if len(reject_words) > 0 {
		for _, w := range reject_words {
			if strings.Contains(requestContent_route, w) {
				requestContent_route = ""
				break
			}
		}
	}

	if len(requestContent_route) > 0 {
		keyword := ""

		// 调用子流程
		if route_method == "flow" {
			msg := meta.ChatFlowMessage{}
			msg.Content = requestContent_route
			msg.FlowCode = route_flow
			msg.FlowSpace = meta.FLOW_SPACE_PRODUCT

			opt := chatSession.Opt

			subChatSession, err := OpenChatSession(opt, msg, []string{meta.CHAT_MESSAGE_TYPE_MESSAGE}, nil)
			if err != nil {
				return andflow.RESULT_FAILURE, err
			}

			subChatSession.Chat(msg)
			responses := subChatSession.GetCurrentResponseMessages()
			for _, m := range responses {
				keyword += m.Content
			}
		}

		if len(keyword) == 0 {
			keyword = requestContent_route
		}

		routeNextActions := r.getNextActionsByKeyword(s, action, keyword)
		if routeNextActions == nil || len(routeNextActions) == 0 {
			routeNextActions = r.getNextActionsByEmptyKeyword(s, action)
		}
		//执行路由下的路径
		if routeNextActions != nil && len(routeNextActions) > 0 {
			for _, a := range routeNextActions {
				state.NextActionIds = append(state.NextActionIds, a.Id)
			}

			return andflow.RESULT_SUCCESS, err
		}
	}

	// 如果有不明确，有设置提问词，就提问
	if len(route_ask) > 0 {
		uid, _ := uuid.NewV4()
		mid := strings.ReplaceAll(uid.String(), "-", "")
		chatSession.Response(meta.ChatFlowMessage{MessageId: mid, Content: route_ask, MessageType: meta.CHAT_MESSAGE_TYPE_MESSAGE}, true)

	}

	return andflow.RESULT_REJECT, err
}
