package flow

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/provider"
)

func init() {
	andflow.RegistActionRunner("ollama_embedding", &OllamaEmbeddingRunner{})
}

type OllamaEmbeddingRunner struct {
	BaseRunner
}

func (r *OllamaEmbeddingRunner) Properties() []andflow.Prop {
	return []andflow.Prop{}
}
func (r *OllamaEmbeddingRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {
	action := s.GetFlow().GetAction(param.ActionId)
	chatSession := r.getChatSession(s)

	log.Printf("ollama embedding begin: %v", time.Now())
	defer log.Printf("ollama embedding end: %v", time.Now())

	prop, err := r.getActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	content_source := prop["content_source"] //信息来自参数还是输入
	content_temp := prop["content_temp"]     //信息来自哪个参数

	url := prop["url"]
	param_key := prop["param_key"] //返回内容

	req_model := prop["req_model"]
	req_model_other := prop["req_model_other"]

	req_keep_alive := prop["req_keep_alive"]

	//地址
	if len(url) == 0 {
		return andflow.RESULT_FAILURE, errors.New("参数 URL 地址不能为空")
	}
	//Model
	if len(req_model) == 0 {
		req_model = req_model_other
	}
	if len(req_model) == 0 {
		return andflow.RESULT_FAILURE, errors.New("参数模型不能为空")
	}

	requestContent := ""
	if content_source == "temp" && len(content_temp) > 0 {
		requestContent = content_temp
	} else {
		messages := chatSession.GetCurrentRequestMessages(1)
		if messages == nil || len(messages) == 0 {
			return andflow.RESULT_REJECT, nil
		}

		for _, msg := range messages {
			requestContent += msg.Content
		}
	}

	params := map[string]string{}
	params["url"] = url
	params["model"] = req_model
	params["keep_alive"] = req_keep_alive
	params["timeout"] = s.GetFlow().Timeout

	if !strings.Contains(params["url"], "/api/embeddings") {
		params["url"] = url + "/api/embeddings"
	}

	embedding := provider.CreateEmbedding("ollama")

	results, err := embedding.Embed(params, []string{requestContent})

	if err != nil {
		msg := fmt.Sprintf("Ollama embedding执行异常:%v", err.Error())
		log.Println(msg)
		return andflow.RESULT_FAILURE, errors.New(msg)
	}

	if len(results) == 0 {
		msg := fmt.Sprintf("Ollama embedding执行异常:%v", "返回空")
		log.Println(msg)
		return andflow.RESULT_FAILURE, errors.New(msg)
	}

	//保存返回内容到参数
	if len(param_key) > 0 && len(results) > 0 {
		s.SetParam(param_key, results[0])
	}

	return andflow.RESULT_SUCCESS, nil
}
