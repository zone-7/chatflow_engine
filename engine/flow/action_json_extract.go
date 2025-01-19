package flow

import (
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
	// 定义正则表达式匹配 JSON 对象或数组
	// 匹配以 { 开始并以 } 结束 或 以 [ 开始并以 ] 结束
	re := regexp.MustCompile(`\{[^{}]*\}|$begin:math:display$[^\\[$end:math:display$]*\]`)

	result := ""
	if content_source == "temp" && len(content_temp) > 0 {

		matches := re.FindAllString(content_temp, -1)
		if len(matches) > 0 {
			result = matches[0]
		}

	} else if content_source == "history" {

		messages := chatSession.GetMessages()
		var i int
		for i = len(messages) - 1; i >= 0; i-- {
			text := messages[i].Content
			// 正则表达式匹配

			// 提取匹配结果
			matches := re.FindAllString(text, -1)
			if len(matches) > 0 {
				result = matches[0]
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

		matches := re.FindAllString(text, -1)
		if len(matches) > 0 {
			result = matches[0]
		}

	}

	if len(param_key) > 0 {
		s.SetParam(param_key, result)
	}

	return andflow.RESULT_SUCCESS, nil
}
