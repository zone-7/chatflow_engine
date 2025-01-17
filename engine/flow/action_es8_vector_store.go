package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/provider/es8"
)

func init() {
	andflow.RegistActionRunner("es8_vector_store", &Es8_vector_store_Runner{})
}

type Es8_vector_store_Runner struct {
	BaseRunner
}

func (r *Es8_vector_store_Runner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {

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
	es_id := prop["es_id"]             //vector
	es_vector := prop["es_vector"]     //vector
	es_payload := prop["es_payload"]   //vector

	if len(es_url) == 0 {
		return andflow.RESULT_FAILURE, errors.New("elasticsearch URL地址不能为空")
	}
	if len(es_index) == 0 {
		return andflow.RESULT_FAILURE, errors.New("elasticsearch 索引不能为空")
	}
	if len(es_vector) == 0 {
		return andflow.RESULT_FAILURE, errors.New("elasticsearch 查询向量不能为空")
	}

	var id string
	if len(strings.TrimSpace(es_id)) == 0 {
		uid, _ := uuid.NewV4()
		id = strings.ReplaceAll(uid.String(), "-", "")
	}

	var vector []float64
	err = json.Unmarshal([]byte(es_vector), &vector)
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	var payload map[string]interface{}
	if len(strings.TrimSpace(es_payload)) > 0 {
		err = json.Unmarshal([]byte(es_payload), &payload)
		if err != nil {
			return andflow.RESULT_FAILURE, err
		}
	}

	es_doc := es8.Es8VectorDocument{}
	es_doc.Id = id
	es_doc.Vector = vector
	es_doc.Payload = payload

	urls := getWords(es_url)

	_, err = es8.Es8StoreVectors(urls, es_username, es_password, es_api_key, es_index, []es8.Es8VectorDocument{es_doc})

	if err != nil {
		fmt.Println("ES存储失败: ", err)
		return andflow.RESULT_FAILURE, err
	}

	if len(param_obj_key) > 0 {
		s.SetParam(param_obj_key, id)
	}

	return andflow.RESULT_SUCCESS, err
}
