package flow

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/provider/qdrant"
)

func init() {
	andflow.RegistActionRunner("qdrant_upsert", &Qdrant_upsert_Runner{})
}

type Qdrant_upsert_Runner struct {
	BaseRunner
}

func (r *Qdrant_upsert_Runner) Properties() []andflow.Prop {
	return []andflow.Prop{}
}
func (r *Qdrant_upsert_Runner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {

	var err error

	action := s.GetFlow().GetAction(param.ActionId)

	prop, err := r.getActionParams(action, s.GetParamMap())

	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	address := prop["address"]       //地址
	port := prop["port"]             //端口
	collection := prop["collection"] //集合名
	autocreate := prop["autocreate"] //自动创建集合
	size := prop["size"]             //openai is 1536
	distance := prop["distance"]     //default Cosine

	vector := prop["vector"]   //向量
	payload := prop["payload"] //数据
	id := prop["id"]           //ID

	if len(address) == 0 {
		return andflow.RESULT_FAILURE, errors.New("数据库地址不能为空")
	}

	if len(port) == 0 {
		return andflow.RESULT_FAILURE, errors.New("数据库端口不能为空")
	}

	if len(collection) == 0 {
		return andflow.RESULT_FAILURE, errors.New("数据库集合名称不能为空")
	}
	if len(vector) == 0 {
		return andflow.RESULT_FAILURE, errors.New("向量数据不能为空")
	}
	if len(id) == 0 {
		return andflow.RESULT_FAILURE, errors.New("向量ID不能为空")
	}

	//自动创建
	if autocreate == "true" || autocreate == "1" {

		var s int
		if len(size) > 0 {
			s, err = strconv.Atoi(size)
		}
		if s <= 0 {
			s = 1536
		}

		d := "Cosine"
		if len(distance) > 0 {
			d = distance
		}

		_, err := qdrant.QdrantPutCollection(address, port, collection, s, d)
		if err != nil {
			return andflow.RESULT_FAILURE, errors.New("数据库集合" + collection + "创建失败")
		}

	}

	//ID
	// var index int64
	// if len(id) > 0 {
	// 	index, _ = strconv.ParseInt(id, 0, 64)

	// }
	// if index < 1 {
	// 	index = 1
	// }

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

	points := make([]qdrant.QdrantPoint, 0)
	points = append(points, qdrant.QdrantPoint{Id: id, Payload: payload, Vector: vt})

	re, err := qdrant.QdrantPutPoints(address, port, collection, points)

	if err != nil {
		return andflow.RESULT_FAILURE, err
	}
	if re.Status != "ok" {
		return andflow.RESULT_FAILURE, errors.New("存储向量数据失败")
	}

	return andflow.RESULT_SUCCESS, nil
}
