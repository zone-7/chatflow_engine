package flow

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/zone-7/chatflow_engine/engine/provider"

	"github.com/zone-7/andflow_go/andflow"
)

func init() {
	andflow.RegistActionRunner("openai_embedding", &OpenAIEmbeddingRunner{})
}

type OpenAIEmbeddingRunner struct {
	BaseRunner
}

func (r *OpenAIEmbeddingRunner) Properties() []andflow.Prop {
	return []andflow.Prop{}
}
func (r *OpenAIEmbeddingRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {

	action := s.GetFlow().GetAction(param.ActionId)
	chatSession := r.getChatSession(s)

	prop, err := r.getActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	content_source := prop["content_source"] //信息来自参数还是输入
	content_temp := prop["content_temp"]     //信息来自哪个参数

	url := prop["url"]
	param_key := prop["param_key"] //返回内容

	req_api_key := prop["req_api_key"]

	req_model := prop["req_model"]
	req_model_other := prop["req_model_other"]

	//返回消息存放变量
	if len(param_key) == 0 {
		return andflow.RESULT_FAILURE, errors.New("返回消息存放地址参数KEY不能为空")
	}

	//地址
	if len(url) == 0 {
		return andflow.RESULT_FAILURE, errors.New("参数 URL 地址不能为空")
	}

	//api key
	if len(req_api_key) == 0 {
		return andflow.RESULT_FAILURE, errors.New("参数 API KEY 不能为空")
	}

	//Model
	if len(req_model) == 0 {
		req_model = req_model_other
	}

	var requestContent string
	if content_source == "temp" && len(content_temp) > 0 {

		requestContent = content_temp

	} else {
		messages := chatSession.GetCurrentRequestMessages(1)
		if messages == nil || len(messages) == 0 {
			return andflow.RESULT_REJECT, nil
		}
		text := ""
		for _, msg := range messages {
			text += msg.Content
		}
		text = strings.ReplaceAll(text, "\n", "")

		requestContent = text

	}

	params := map[string]string{}
	params["url"] = url

	if !strings.Contains(params["url"], "/embeddings") {
		params["url"] = url + "/embeddings"
	}

	params["model"] = req_model
	params["api_key"] = req_api_key
	params["timeout"] = s.GetFlow().Timeout

	embedding := provider.CreateEmbedding("openai")
	results, err := embedding.Embed(params, []string{requestContent})

	if err != nil {
		msg := fmt.Sprintf("Openai embedding执行异常:%v", err.Error())
		log.Println(msg)
		return andflow.RESULT_FAILURE, errors.New(msg)
	}
	if len(results) == 0 {
		msg := fmt.Sprintf("Openai embedding执行异常:%v", "返回空")
		log.Println(msg)
		return andflow.RESULT_FAILURE, errors.New(msg)
	}

	//保存返回内容到参数
	if len(param_key) > 0 && len(results) > 0 {

		s.SetParam(param_key, results[0])

	}

	return andflow.RESULT_SUCCESS, nil
}
