package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/provider/qdrant"
)

func init() {
	andflow.RegistActionRunner("qdrant_search", &Qdrant_search_Runner{})
}

type Qdrant_search_Runner struct {
	BaseRunner
}

func (r *Qdrant_search_Runner) Properties() []andflow.Prop {
	return []andflow.Prop{}
}
func (r *Qdrant_search_Runner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {

	var err error

	action := s.GetFlow().GetAction(param.ActionId)

	prop, err := r.getActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}
	param_key := prop["param_key"] //返回内容
	// param_obj_key := prop["param_obj_key"]                   //返回内容对象
	// param_payload_key := prop["param_payload_key"]           //返回第一个有效数据对象
	param_payload_text_key := prop["param_payload_text_key"] // 第一个有效数据文本
	address := prop["address"]                               //地址
	port := prop["port"]                                     //端口
	collection := prop["collection"]                         //集合名
	vector := prop["vector"]                                 //向量
	limit := prop["limit"]                                   //数量
	score := prop["score"]                                   //分数

	if len(address) == 0 {
		return andflow.RESULT_FAILURE, errors.New("数据库地址不能为空")
	}

	if len(port) == 0 {
		return andflow.RESULT_FAILURE, errors.New("数据库端口不能为空")
	}

	if len(collection) == 0 {
		return andflow.RESULT_FAILURE, errors.New("数据库集合名称不能为空")
	}

	//vector
	var vt []float64
	err = json.Unmarshal([]byte(vector), &vt)
	if err != nil {
		var vts [][]float64
		err = json.Unmarshal([]byte(vector), &vts)
		if err != nil {
			return andflow.RESULT_FAILURE, errors.New("Vector 数据格式错误")
		}
		if len(vts) > 0 {
			vt = vts[0]
		}
	}
	if vt == nil || len(vt) == 0 {
		return andflow.RESULT_FAILURE, errors.New("Vector 数据格式错误")
	}

	//score_threshold
	var score_threshold float64
	if len(score) > 0 {
		score_threshold, _ = strconv.ParseFloat(score, 64)
	}
	//limit
	var lm int
	if len(limit) > 0 {
		lm, _ = strconv.Atoi(limit)
	}
	if lm == 0 {
		lm = 1
	}

	results, err := qdrant.QdrantSearchPoints(address, port, collection, vt, score_threshold, lm)

	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	if len(param_key) > 0 {
		s.SetParam(param_key, results)
	}

	if len(param_payload_text_key) > 0 {
		if len(results.Result) > 0 {
			textArr := make([]string, 0)
			for _, result := range results.Result {
				if payload, ok := result.Payload.(map[string]interface{}); ok {
					txt_data := payload["text"]
					txt := fmt.Sprintf("%v", txt_data)
					textArr = append(textArr, txt)

				}
			}
			s.SetParam(param_payload_text_key, strings.Join(textArr, ""))
		}
	}

	return andflow.RESULT_SUCCESS, nil
}
