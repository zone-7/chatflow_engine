package flow

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/provider/es8"
)

func init() {
	andflow.RegistActionRunner("es8_content_search", &Es8_content_search_Runner{})
}

type Es8_content_search_Runner struct {
	BaseRunner
}

func (r *Es8_content_search_Runner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {

	var err error
	action := s.GetFlow().GetAction(param.ActionId)

	prop, err := r.getActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	param_obj_key := prop["param_obj_key"]                         //返回内容对象
	param_hits_key := prop["param_hits_key"]                       //返回命中内容列表
	param_source_key := prop["param_source_key"]                   //返回命中内容原始数据列表
	param_highlight_key := prop["param_highlight_key"]             //返回命中内容高亮数据列表
	param_source_first_key := prop["param_source_first_key"]       //返回命中内容原始数据列表
	param_highlight_first_key := prop["param_highlight_first_key"] //返回命中内容高亮数据列表
	es_url := prop["es_url"]                                       //localhost:9200
	es_index := prop["es_index"]                                   //索引 index
	es_username := prop["es_username"]                             //username
	es_password := prop["es_password"]                             //password
	es_api_key := prop["es_api_key"]                               //api key
	es_query := prop["es_query"]                                   //query
	es_size := prop["es_size"]                                     //top

	if len(es_url) == 0 {
		return andflow.RESULT_FAILURE, errors.New("elasticsearch URL地址不能为空")
	}
	if len(es_index) == 0 {
		return andflow.RESULT_FAILURE, errors.New("elasticsearch 索引不能为空")
	}
	if len(es_query) == 0 {
		return andflow.RESULT_FAILURE, errors.New("elasticsearch 查询内容不能为空")
	}

	urls := getWords(es_url)

	// query := fmt.Sprintf(`{
	// 	"query":{
	// 		"match":{
	// 			"%s": %s
	// 		}
	// 	},
	// 	"highlight":{
	// 		"pre_tags": "<font color='red'>",
	// 		"post_tags": "</font>",
	// 		"fields": {
	// 			"%s": {}
	// 		}
	// 	}
	// }
	// `, es_field, es_query, es_field)

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

	res, err := es8.Es8Search(urls, es_username, es_password, es_api_key, es_index, es_query, size)

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

	sources := make([]map[string]interface{}, 0)
	highlights := make([]map[string]interface{}, 0)

	for _, hit := range res.Hits.Hits {
		// 处理每个hit
		source := hit.Source
		highlight := hit.Highlight

		if source != nil {
			sources = append(sources, source)
		}

		if highlight != nil {
			highlights = append(highlights, highlight)
		}
	}

	if len(param_source_key) > 0 {
		s.SetParam(param_source_key, sources)
	}
	if len(param_highlight_key) > 0 {
		s.SetParam(param_highlight_key, highlights)
	}

	if len(param_source_first_key) > 0 {
		if len(sources) > 0 {
			s.SetParam(param_source_first_key, sources[0])

		} else {
			s.SetParam(param_source_first_key, "")
		}
	}

	if len(param_highlight_first_key) > 0 {
		if len(highlights) > 0 {
			s.SetParam(param_highlight_first_key, highlights[0])
		} else {
			s.SetParam(param_highlight_first_key, "")
		}
	}

	return andflow.RESULT_SUCCESS, err
}
