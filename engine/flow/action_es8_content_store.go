package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/provider/es8"
)

func init() {
	andflow.RegistActionRunner("es8_content_store", &Es8_content_store_Runner{})
}

type Es8_content_store_Runner struct {
	BaseRunner
}

func (r *Es8_content_store_Runner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {

	var err error
	action := s.GetFlow().GetAction(param.ActionId)

	prop, err := r.getActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	param_obj_key := prop["param_obj_key"] //返回内容对象

	es_url := prop["es_url"]           //localhost:9200
	es_index := prop["es_index"]       //索引 index
	es_username := prop["es_username"] //username
	es_password := prop["es_password"] //password
	es_api_key := prop["es_api_key"]   //api key
	es_doc := prop["es_doc"]           //doc

	if len(es_url) == 0 {
		return andflow.RESULT_FAILURE, errors.New("elasticsearch URL地址不能为空")
	}
	if len(es_index) == 0 {
		return andflow.RESULT_FAILURE, errors.New("elasticsearch 索引不能为空")
	}
	if len(es_doc) == 0 {
		return andflow.RESULT_FAILURE, errors.New("elasticsearch 存储内容不能为空")
	}

	var id string

	var docs []map[string]interface{}

	if strings.Index(strings.TrimSpace(es_doc), "[") == 0 {
		err = json.Unmarshal([]byte(es_doc), &docs)
		if err != nil {
			return andflow.RESULT_FAILURE, err
		}
	} else {
		var doc map[string]interface{}
		err = json.Unmarshal([]byte(es_doc), &doc)
		if err != nil {
			return andflow.RESULT_FAILURE, err
		}

		docs = append(docs, doc)
	}

	urls := getWords(es_url)

	_, err = es8.Es8Store(urls, es_username, es_password, es_api_key, es_index, docs)

	if err != nil {
		fmt.Println("ES存储失败: ", err)
		return andflow.RESULT_FAILURE, err
	}

	if len(param_obj_key) > 0 {
		s.SetParam(param_obj_key, id)
	}

	return andflow.RESULT_SUCCESS, err
}
