package flow

// 意图识别
import (
	"encoding/json"

	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/meta"
	"github.com/zone-7/chatflow_engine/engine/utils"
)

func init() {
	andflow.RegistActionRunner("subflow", &SubflowRunner{})
}

type SubflowRunner struct {
	BaseRunner
}

func (r *SubflowRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {
	var err error

	action := s.GetFlow().GetAction(param.ActionId)
	chatSession := r.getChatSession(s)

	prop, err := r.getActionParams(action, s.GetParamMap())

	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	source := prop["source"]
	source_temp := prop["source_temp"]

	flow_code := prop["flow_code"]
	response_params := prop["response_params"]
	flow_params_json := r.getActionParam(action, "flow_params", nil)

	flow_params := make(map[string]string)

	if len(flow_params_json) > 0 {
		json.Unmarshal([]byte(flow_params_json), &flow_params)
	}

	for k, v := range flow_params {
		vv, err := replaceTemplate(v, "temp_"+action.Id+"_params_"+k, s.GetParamMap())
		if err == nil {
			flow_params[k] = vv
		}
	}

	// 获取消息文本
	requestContent := ""
	if source == "temp" && len(source_temp) > 0 {
		requestContent = source_temp
	} else {
		requestContent = chatSession.GetCurrentRequestMessagesContent(1)
	}

	if len(requestContent) > 0 {

		msg := meta.ChatFlowMessage{}
		msg.Content = requestContent
		msg.FlowCode = flow_code
		msg.FlowSpace = meta.FLOW_SPACE_PRODUCT
		msg.Params = flow_params
		msg.RequestId = s.Operation.GetRequestId()

		opt := chatSession.Opt

		subChatSession, err := OpenChatSession(opt, msg, []string{meta.CHAT_MESSAGE_TYPE_MESSAGE}, func(message meta.ChatFlowMessage) {
			chatSession.Response(message, true)
		})

		if err != nil {
			return andflow.RESULT_FAILURE, err
		}

		subChatSession.Chat(msg)

		keys := getWords(response_params)

		subflowParams := subChatSession.Runtime.GetParamMap()
		for k, v := range subflowParams {
			if k == "message" {
				continue
			}

			if utils.StringsContains(keys, k) >= 0 {
				s.SetParam(k, v)
			}
		}
	}

	return andflow.RESULT_SUCCESS, nil
}
