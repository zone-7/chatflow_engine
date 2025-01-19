package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/provider/es8"
)

func init() {
	andflow.RegistActionRunner("es8_vector_search", &Es8_vector_search_Runner{})
}

type Es8_vector_search_Runner struct {
	BaseRunner
}

func (r *Es8_vector_search_Runner) Properties() []andflow.Prop {
	return []andflow.Prop{}
}
func (r *Es8_vector_search_Runner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {

	var err error
	action := s.GetFlow().GetAction(param.ActionId)

	prop, err := r.GetActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	param_obj_key := prop["param_obj_key"]                   //返回内容对象
	param_hits_key := prop["param_hits_key"]                 //返回命中内容列表
	param_source_key := prop["param_source_key"]             //返回命中内容原始数据列表
	param_source_first_key := prop["param_source_first_key"] //返回命中内容原始数据列表
	es_url := prop["es_url"]                                 //localhost:9200
	es_index := prop["es_index"]                             //索引 index
	es_username := prop["es_username"]                       //username
	es_password := prop["es_password"]                       //password
	es_api_key := prop["es_api_key"]                         //api key
	es_vector := prop["es_vector"]                           //vector

	es_distance := prop["distance"] //匹配算法

	es_size := prop["es_size"] //top

	es_timeout := prop["es_timeout"] //es_timeout

	if len(es_url) == 0 {
		return andflow.RESULT_FAILURE, errors.New("elasticsearch URL地址不能为空")
	}
	if len(es_index) == 0 {
		return andflow.RESULT_FAILURE, errors.New("elasticsearch 索引不能为空")
	}
	if len(es_vector) == 0 {
		return andflow.RESULT_FAILURE, errors.New("elasticsearch 查询向量不能为空")
	}

	var vector []float64
	err = json.Unmarshal([]byte(es_vector), &vector)
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	urls := getWords(es_url)

	var size int
	if len(es_size) > 0 {
		size, err = strconv.Atoi(es_size)
	}
	if size <= 0 {
		size = 10
	}
	if size > 1000 {
		size = 1000
	}

	var timeout int
	if len(es_timeout) > 0 {
		timeout, _ = strconv.Atoi(es_timeout)
	}
	if timeout < 1000 {
		timeout = 1000
	}
	if timeout > 1000*60 {
		timeout = 1000 * 60
	}

	if len(es_distance) == 0 {
		es_distance = "Cosine"
	}

	res, err := es8.Es8SearchVectors(urls, es_username, es_password, es_api_key, es_index, vector, es_distance, size)

	if err != nil {
		fmt.Println("ES检索失败: ", err)
		return andflow.RESULT_FAILURE, err
	}

	if len(param_obj_key) > 0 {
		s.SetParam(param_obj_key, res)
	}

	if res.Hits.Hits == nil || len(res.Hits.Hits) == 0 {
		return andflow.RESULT_SUCCESS, err
	}

	// 处理返回的结果
	if len(param_hits_key) > 0 {
		s.SetParam(param_hits_key, res.Hits.Hits)
	}

	sources := make([]es8.Es8VectorDocument, 0)

	for _, hit := range res.Hits.Hits {
		// 处理每个hit
		source := hit.Source
		sources = append(sources, source)
	}

	if len(param_source_key) > 0 {
		s.SetParam(param_source_key, sources)
	}

	if len(param_source_first_key) > 0 {
		if len(sources) > 0 {
			s.SetParam(param_source_first_key, sources[0])

		} else {
			s.SetParam(param_source_first_key, "")
		}
	}

	return andflow.RESULT_SUCCESS, err
}
