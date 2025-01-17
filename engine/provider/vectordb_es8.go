package provider

import (
	"errors"
	"strconv"
	"strings"

	"github.com/zone-7/chatflow_engine/engine/provider/es8"
)

func init() {
	v := VectorDB_es8{}
	vectordbs = append(vectordbs, v.GetDict().Name)
}

// 获取词语列表，逗号或者换行分隔
func getWords(keywords string) []string {
	words := make([]string, 0)
	kws1 := strings.Split(keywords, "\n")
	for _, kw1 := range kws1 {
		kws2 := strings.Split(kw1, ",")
		for _, kw2 := range kws2 {
			kws3 := strings.Split(kw2, "，")
			for _, kw3 := range kws3 {
				kws4 := strings.Split(kw3, " ")
				for _, kw4 := range kws4 {
					if len(kw4) > 0 {
						words = append(words, kw4)
					}
				}
			}
		}
	}

	return words
}

type VectorDB_es8 struct {
	Urls     []string `json:"urls" yaml:"urls"`
	Username string   `json:"username" yaml:"username"`
	Password string   `json:"password" yaml:"password"`
	ApiKey   string   `json:"api_key" yaml:"api_key"`
	Index    string   `json:"index" yaml:"index"`       //集合
	Size     int      `json:"size" yaml:"size"`         //向量维度大小
	Distance string   `json:"distance" yaml:"distance"` //比对距离公式
}

func (v *VectorDB_es8) GetDict() Dict {
	dict := Dict{}
	dict.Name = "es8"
	dict.Fields = []Field{}

	return dict
}

func (v *VectorDB_es8) setParams(params map[string]string) error {
	urls := params["urls"]         //地址
	username := params["username"] //用户名
	password := params["password"] //密码
	api_key := params["api_key"]   //密码

	index := params["collection"]  //集合名
	size := params["size"]         //openai is 1536
	distance := params["distance"] //算法 default Cosine

	if len(urls) == 0 {
		return errors.New("地址不能为空")
	}
	if len(index) == 0 {
		return errors.New("集合不能为空")
	}

	v.Urls = getWords(urls)
	v.Username = username
	v.Password = password
	v.ApiKey = api_key
	v.Index = index
	v.Distance = distance
	v.Size, _ = strconv.Atoi(size)

	if len(v.Distance) == 0 {
		v.Distance = "Cosine"
	}

	return nil
}

func (v *VectorDB_es8) Search(params map[string]string, vector []float64, score float64, limit int) ([]*VectorData, error) {
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

	hits, err := es8.Es8SearchVectors(v.Urls, v.Username, v.Password, v.ApiKey, v.Index, vector, v.Distance, limit)
	if err != nil {
		return nil, err
	}

	datas := make([]*VectorData, 0)
	if len(hits.Hits.Hits) > 0 {
		for _, h := range hits.Hits.Hits {

			v := VectorData{}
			v.Id = h.Source.Id
			v.Payload = h.Source.Payload
			v.Vector = h.Source.Vector
			v.Score = h.Score

			datas = append(datas, &v)
		}
	}

	return datas, nil

}

func (v *VectorDB_es8) Save(params map[string]string, datas []*VectorData) error {
	err := v.setParams(params)
	if err != nil {
		return err
	}

	// 判断集合是否存在，否则创建
	_, err = es8.Es8GetIndex(v.Urls, v.Username, v.Password, v.ApiKey, v.Index)

	if err != nil {
		_, err = es8.Es8CreateVectorIndex(v.Urls, v.Username, v.Password, v.ApiKey, v.Index, v.Size)
		if err != nil {
			return errors.New("数据库集合" + v.Index + "创建失败")
		}
	}

	//保存
	docs := make([]es8.Es8VectorDocument, 0)
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

		docs = append(docs, es8.Es8VectorDocument{Id: data.Id, Payload: data.Payload, Vector: vector})
	}

	re, err := es8.Es8StoreVectors(v.Urls, v.Username, v.Password, v.ApiKey, v.Index, docs)

	if err != nil {
		return err
	}

	if re.IsError() {
		return errors.New("存储向量数据失败")
	}

	return nil
}

func (v *VectorDB_es8) Get(params map[string]string, id string) (*VectorData, error) {
	err := v.setParams(params)
	if err != nil {
		return nil, err
	}

	hit, err := es8.Es8GetVector(v.Urls, v.Username, v.Password, v.ApiKey, v.Index, id)
	if err != nil {
		return nil, err
	}

	data := &VectorData{}
	data.Id = hit.Id
	data.Payload = hit.Source.Payload
	data.Vector = hit.Source.Vector
	data.Score = hit.Score

	return data, nil
}

func (v *VectorDB_es8) Remove(params map[string]string, id string) error {
	err := v.setParams(params)
	if err != nil {
		return err
	}

	response, err := es8.Es8Delete(v.Urls, v.Username, v.Password, v.ApiKey, v.Index, id)
	if err != nil {
		return err
	}
	if response.IsError() {
		return errors.New("删除向量数据失败")
	}

	return nil
}

func (v *VectorDB_es8) Clear(params map[string]string) error {
	err := v.setParams(params)
	if err != nil {
		return err
	}
	response, err := es8.Es8Clear(v.Urls, v.Username, v.Password, v.ApiKey, v.Index)
	if err != nil {
		return err
	}
	if response.IsError() {
		return errors.New("清除向量数据失败")
	}

	return nil

}
