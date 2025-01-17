package flow

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/manager"
)

func init() {
	andflow.RegistActionRunner("knowledge_search", &Knowledge_search_Runner{})
}

type Knowledge_search_Runner struct {
	BaseRunner
}

func (r *Knowledge_search_Runner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {

	var err error

	action := s.GetFlow().GetAction(param.ActionId)

	chatSession := r.getChatSession(s)

	prop, err := r.getActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	param_source := prop["param_source"]                     //参数信息来自参数还是输入
	param_source_temp := prop["param_source_temp"]           //参数信息来自哪个参数
	param_key := prop["param_key"]                           //返回内容
	param_payload_text_key := prop["param_payload_text_key"] // 第一个有效数据文本
	knowledge_id := prop["knowledge_id"]                     //知识库ID
	limit := prop["limit"]                                   //检索条数
	score := prop["score"]                                   //阈值

	if len(knowledge_id) == 0 {
		return andflow.RESULT_FAILURE, errors.New("知识库不能为空")
	}

	// 获取消息
	requestContent_param := ""
	if param_source == "temp" && len(param_source_temp) > 0 {
		requestContent_param = param_source_temp
	} else {
		chatSession := r.getChatSession(s)
		requestContent_param = chatSession.GetCurrentRequestMessagesContent(1)
	}

	//limit
	var lm int
	if len(limit) > 0 {
		lm, _ = strconv.Atoi(limit)
	}
	if lm == 0 {
		lm = 1
	}
	//score_threshold
	var sc float64
	if len(score) > 0 {
		sc, _ = strconv.ParseFloat(score, 64)
	}

	opt := chatSession.Opt

	kno := manager.KnowledgeManager{Opt: opt}
	results, err := kno.SearchKnowledge(knowledge_id, requestContent_param, sc, lm)

	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	if len(results) > 0 {
		textArr := make([]string, 0)
		for _, result := range results {
			if payload, ok := result.Payload.(map[string]interface{}); ok {
				txt_data := payload["text"]
				txt := fmt.Sprintf("%v", txt_data)
				textArr = append(textArr, txt)

			}
		}
		if len(param_key) > 0 {
			s.SetParam(param_key, textArr)
		}
		if len(param_payload_text_key) > 0 {
			s.SetParam(param_payload_text_key, strings.Join(textArr, ""))
		}

	}

	return andflow.RESULT_SUCCESS, nil
}
