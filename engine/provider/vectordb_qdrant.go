package provider

import (
	"errors"
	"strconv"

	"github.com/zone-7/chatflow_engine/engine/provider/qdrant"
	"github.com/zone-7/chatflow_engine/engine/utils"
)

func init() {
	v := VectorDB_qdrant{}
	vectordbs = append(vectordbs, v.GetDict().Name)
}

type VectorDB_qdrant struct {
	Address    string `json:"address" yaml:"address"`
	Port       string `json:"port" yaml:"port"`
	Collection string `json:"collection" yaml:"collection"` //集合
	Size       int    `json:"size" yaml:"size"`             //向量维度大小
	Distance   string `json:"distance" yaml:"distance"`     //比对距离公式
}

func (v *VectorDB_qdrant) GetDict() Dict {
	dict := Dict{}
	dict.Name = "qdrant"
	dict.Fields = []Field{}
	dict.Fields = append(dict.Fields, Field{Name: "address"})
	dict.Fields = append(dict.Fields, Field{Name: "port"})
	dict.Fields = append(dict.Fields, Field{Name: "collection"})
	dict.Fields = append(dict.Fields, Field{Name: "size"})
	dict.Fields = append(dict.Fields, Field{Name: "limit"})
	dict.Fields = append(dict.Fields, Field{Name: "score"})
	dict.Fields = append(dict.Fields, Field{Name: "distance"})

	return dict
}

func (v *VectorDB_qdrant) setParams(params map[string]string) error {
	address := params["address"]       //地址
	port := params["port"]             //端口
	collection := params["collection"] //集合名
	size := params["size"]             //openai is 1536
	distance := params["distance"]     //default Cosine

	if len(address) == 0 {
		return errors.New("地址不能为空")
	}
	if len(collection) == 0 {
		return errors.New("集合不能为空")
	}

	if len(port) == 0 {
		port = "6333"
	}

	v.Address = address
	v.Port = port
	v.Collection = collection
	v.Distance = distance
	v.Size, _ = strconv.Atoi(size)

	if len(v.Distance) == 0 {
		v.Distance = "Cosine"
	}

	return nil
}

func (v *VectorDB_qdrant) Search(params map[string]string, vector []float64, score float64, limit int) ([]*VectorData, error) {
	err := v.setParams(params)
	if err != nil {
		return nil, err
	}

	// 向量大小如果不等于数据库维度，就进行截断或者补充
	if len(vector) > v.Size {
		vector = vector[:v.Size]
	}
	if len(vector) < v.Size {
		for len(vector) < v.Size {
			vector = append(vector, 0)
		}
	}

	var results []*VectorData

	response, err := qdrant.QdrantSearchPoints(v.Address, v.Port, v.Collection, vector, score, limit)

	if err != nil {
		return results, err
	}

	if response.Status != "ok" {
		return results, errors.New("检索失败:" + utils.ToString(response.Status))
	}

	for _, r := range response.Result {
		data := &VectorData{}
		data.Id = r.Id
		data.Vector = r.Vector
		data.Payload = r.Payload
		data.Score = r.Score
		results = append(results, data)
	}

	return results, nil
}

func (v *VectorDB_qdrant) Save(params map[string]string, datas []*VectorData) error {
	err := v.setParams(params)
	if err != nil {
		return err
	}

	// 判断集合是否存在，否则创建
	response, err := qdrant.QdrantGetCollection(v.Address, v.Port, v.Collection)

	if err != nil || response == nil || response.Status != "ok" {
		_, err = qdrant.QdrantPutCollection(v.Address, v.Port, v.Collection, v.Size, v.Distance)
		if err != nil {
			return errors.New("数据库集合" + v.Collection + "创建失败")
		}
	}

	//保存
	points := make([]qdrant.QdrantPoint, 0)
	for _, data := range datas {
		vector := data.Vector
		if len(vector) > v.Size {
			vector = vector[:v.Size]
		}
		if len(vector) < v.Size {
			for len(vector) < v.Size {
				vector = append(vector, 0)
			}
		}

		points = append(points, qdrant.QdrantPoint{Id: data.Id, Payload: data.Payload, Vector: vector})
	}

	re, err := qdrant.QdrantPutPoints(v.Address, v.Port, v.Collection, points)

	if err != nil {
		return err
	}
	if re.Status != "ok" {
		return errors.New("存储向量数据失败:" + utils.ToString(re.Status))
	}

	return nil
}

func (v *VectorDB_qdrant) Get(params map[string]string, id string) (*VectorData, error) {
	err := v.setParams(params)
	if err != nil {
		return nil, err
	}

	response, err := qdrant.QdrantGetPoint(v.Address, v.Port, v.Collection, id)
	if err != nil {
		return nil, err
	}
	if response.Status != "ok" {
		return nil, errors.New("读取向量数据失败")
	}

	data := &VectorData{}
	data.Id = response.Result.Id
	data.Payload = response.Result.Payload
	data.Vector = response.Result.Vector

	return data, nil
}

func (v *VectorDB_qdrant) Remove(params map[string]string, id string) error {
	err := v.setParams(params)
	if err != nil {
		return err
	}

	response, err := qdrant.QdrantDeletePoint(v.Address, v.Port, v.Collection, id)
	if err != nil {
		return err
	}
	if response.Status != "ok" {
		return errors.New("删除向量数据失败")
	}

	return nil
}

func (v *VectorDB_qdrant) Clear(params map[string]string) error {
	err := v.setParams(params)
	if err != nil {
		return err
	}
	response, err := qdrant.QdrantDeleteCollection(v.Address, v.Port, v.Collection)
	if err != nil {
		return err
	}
	if response.Status != "ok" {
		return errors.New("清空向量数据失败")
	}

	return nil

}
