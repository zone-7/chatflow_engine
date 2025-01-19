package flow

import (
	"encoding/json"
	"log"
	"regexp"
	"time"

	"github.com/zone-7/andflow_go/andflow"
)

// 从聊天记录或者当前参数中提取JSON

func init() {
	andflow.RegistActionRunner("json_extract", &Json_extract_runner{})
}

type Json_extract_runner struct {
	BaseRunner
}

func (r *Json_extract_runner) Properties() []andflow.Prop {
	return []andflow.Prop{}
}

func (r *Json_extract_runner) getJson(s string) string {
	// 匹配可能嵌套的 JSON 对象或数组 (通过贪婪模式匹配)
	re := regexp.MustCompile(`(?s)\{(?:[^{}]|\{.*\}|$begin:math:display$.*$end:math:display$)*}\s*|$begin:math:display$(?:[^\\[$end:math:display$]|\{.*\}|$begin:math:display$.*$end:math:display$)*]\s*`)
	matches := re.FindString(s)
	if matches == "" {
		return ""
	}

	var result interface{}
	// 尝试解析 JSON
	if err := json.Unmarshal([]byte(matches), &result); err != nil {
		return ""
	}

	return matches
}

func (r *Json_extract_runner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {
	var err error
	action := s.GetFlow().GetAction(param.ActionId)
	chatSession := r.GetChatSession(s)

	log.Printf("Json_extract_runner begin: %v", time.Now())
	defer log.Printf("Json_extract_runner end: %v", time.Now())

	prop, err := r.GetActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	content_source := prop["content_source"] //信息来自参数还是输入
	content_temp := prop["content_temp"]     //信息来自哪个参数
	param_key := prop["param_key"]

	result := ""
	if content_source == "temp" && len(content_temp) > 0 {
		result = r.getJson(content_temp)
	} else if content_source == "history" {

		messages := chatSession.GetMessages()
		var i int
		for i = len(messages) - 1; i >= 0; i-- {
			text := messages[i].Content
			result = r.getJson(text)

			if len(result) > 0 {
				break
			}
		}

	} else {
		messages := chatSession.GetCurrentRequestMessages(1)
		if messages == nil || len(messages) == 0 {
			return andflow.RESULT_REJECT, nil
		}
		text := ""
		for _, msg := range messages {
			text += msg.Content
		}

		result = r.getJson(text)

	}

	if len(param_key) > 0 {
		s.SetParam(param_key, result)
	}

	return andflow.RESULT_SUCCESS, nil
}
