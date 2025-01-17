package es8

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gofrs/uuid"
)

type Es8VectorDocument struct {
	Vector  []float64 `json:"vector"`
	Id      string    `json:"id"`
	Payload any       `json:"payload"`
}

type Hit[T any] struct {
	Source    T                      `json:"_source"`
	Index     string                 `json:"_index"`
	Id        string                 `json:"_id"`
	Score     float64                `json:"_score"`
	Highlight map[string]interface{} `json:"highlight"`
}

// 定义与检索结果相对应的结构体
type Hits[T any] struct {
	Hits struct {
		Hits []Hit[T] `json:"hits"`
	} `json:"hits"`
}
type IndexInfo struct {
	Mappings struct {
		Properties struct {
			Id struct {
				Type   string                 `json:"type"`
				Fields map[string]interface{} `json:"fields"`
			} `json:"id"`
			Payload map[string]interface{} `json:"payload"`
			Vector  struct {
				Type       string  `json:"type"`
				Dims       float64 `json:"dims"`
				Index      bool    `json:"index"`
				Similarity string  `json:"similarity"`
			} `json:"vector"`
		} `json:"properties"`
	} `json:"mappings"`
	Settings struct {
		Index struct {
			NumberOfShards   string `json:"number_of_shards"`
			NumberOfReplicas string `json:"number_of_replicas"`
			ProvidedName     string `json:"provided_name"`
		} `json:"index"`
	} `json:"settings"`
}

func Es8GetIndex(urls []string, username string, password string, apiKey string, index string) (*IndexInfo, error) {
	// 创建配置
	cfg := elasticsearch.Config{
		Addresses: urls,
		Username:  username,
		Password:  password,
		APIKey:    apiKey,
	}

	// 创建客户端
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("创建elasticsearch客户端失败: ", err)
		return nil, err
	}

	// 获取索引信息
	res, err := es.Indices.Get([]string{index})
	if err != nil {
		fmt.Println("Error getting index information: ", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	var data map[string]interface{}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	if data[index] == nil {
		return nil, err
	}

	infoByte, err := json.Marshal(data[index])
	if err != nil {
		return nil, err
	}

	var info IndexInfo

	err = json.Unmarshal(infoByte, &info)
	if err != nil {
		return nil, err
	}

	return &info, err
}

// 创建索引
func Es8CreateIndex(urls []string, username string, password string, apiKey string, index string, mapping string) (*esapi.Response, error) {
	// 创建配置
	cfg := elasticsearch.Config{
		Addresses: urls,
		Username:  username,
		Password:  password,
		APIKey:    apiKey,
	}

	// 创建客户端
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("创建elasticsearch客户端失败: ", err)
		return nil, err
	}

	// 创建索引请求
	req := esapi.IndicesCreateRequest{
		Index: index, // 索引名称
		Body:  strings.NewReader(mapping),
	}

	// 执行请求
	res, err := req.Do(context.Background(), es)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	return res, nil
}

// 构建查询
func Es8Search(urls []string, username string, password string, apiKey string, index string, query string, limit int) (*Hits[map[string]interface{}], error) {
	// 创建配置
	cfg := elasticsearch.Config{
		Addresses: urls,
		Username:  username,
		Password:  password,
		APIKey:    apiKey,
	}

	// 创建客户端
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("创建elasticsearch客户端失败: ", err)
		return nil, err
	}

	// 创建查询
	// query := fmt.Sprintf(`{
	// 	"query": {
	// 		"script_score": {
	// 			"query": {
	// 				"match_all": {}
	// 			},
	// 			"script": {
	// 				"source": "%s",
	// 				"params": {
	// 					"query_vector": %v
	// 				}
	// 			}
	// 		}
	// 	}
	// }`, distance_source, string(data))
	// query := `
	// {
	// 	"query": {
	// 		"match": {
	// 			"title": "example" // 替换为你想要搜索的关键字
	// 		}
	// 	}
	// }`

	// 执行搜索请求
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(index),
		es.Search.WithBody(strings.NewReader(query)),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithSize(limit),
		es.Search.WithPretty(),
	)

	if err != nil {
		fmt.Println("Error getting response: ", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		fmt.Println("Error response:", res.String())
		return nil, errors.New(res.String())
	}

	b, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error getting response: ", err)
		return nil, err
	}

	var hits Hits[map[string]interface{}]

	if err := json.Unmarshal(b, &hits); err != nil {
		fmt.Println("Error parsing the response body: ", err)
		return nil, err
	}

	return &hits, nil
}

func Es8Store(urls []string, username string, password string, apiKey string, index string, docs []map[string]interface{}) (*esapi.Response, error) {
	// 创建配置
	cfg := elasticsearch.Config{
		Addresses: urls,
		Username:  username,
		Password:  password,
		APIKey:    apiKey,
	}

	// 创建客户端
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("创建elasticsearch客户端失败: ", err)
		return nil, err
	}

	// 准备批量数据
	var buf bytes.Buffer
	for _, doc := range docs {

		var id string
		if doc["id"] != nil {
			id = fmt.Sprintf("%v", doc["id"])
		}
		if len(id) == 0 {
			uid, _ := uuid.NewV4()
			id = strings.ReplaceAll(uid.String(), "-", "")
		}

		meta := fmt.Sprintf(`{ "index" : { "_index" : "%s", "_id" : "%s" } }%s`, index, id, "\n")
		data, err := json.Marshal(doc)
		if err != nil {
			fmt.Println("Error marshaling document: ", err)
			continue
		}

		buf.WriteString(meta)
		buf.Write(data)
		buf.WriteString("\n")
	}

	// 执行批量操作
	res, err := es.Bulk(bytes.NewReader(buf.Bytes()), es.Bulk.WithIndex(index), es.Bulk.WithContext(context.Background()))

	if err != nil {
		fmt.Println("Error performing bulk request: ", err)
		return nil, err
	}

	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}
	return res, err
}

func Es8CreateVectorIndex(urls []string, username string, password string, apiKey string, index string, dim_size int) (*esapi.Response, error) {
	// 创建配置
	cfg := elasticsearch.Config{
		Addresses: urls,
		Username:  username,
		Password:  password,
		APIKey:    apiKey,
	}

	// 创建客户端
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("创建elasticsearch客户端失败: ", err)
		return nil, err
	}

	// 创建索引并定义映射
	mapping := `{
		"mappings": {
			"properties": {
				"vector": {
					"type": "dense_vector",
					"dims": ` + fmt.Sprintf("%d", dim_size) + `
				}
			}
		}
	}`

	res, err := es.Indices.Create(
		index,
		es.Indices.Create.WithBody(strings.NewReader(mapping)),
	)

	if err != nil {
		fmt.Println("Error creating index: ", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return res, errors.New(res.String())
	}

	return res, nil
}

func Es8StoreVectors(urls []string, username string, password string, apiKey string, index string, docs []Es8VectorDocument) (*esapi.Response, error) {
	// 创建配置
	cfg := elasticsearch.Config{
		Addresses: urls,
		Username:  username,
		Password:  password,
		APIKey:    apiKey,
	}

	// 创建客户端
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("创建elasticsearch客户端失败: ", err)
		return nil, err
	}

	// 准备批量数据
	var buf bytes.Buffer
	for _, doc := range docs {

		meta := fmt.Sprintf(`{ "index" : { "_index" : "%s", "_id" : "%s" } }%s`, index, doc.Id, "\n")
		data, err := json.Marshal(doc)
		if err != nil {
			fmt.Println("Error marshaling document: ", err)
			continue
		}

		buf.WriteString(meta)
		buf.Write(data)
		buf.WriteString("\n")
	}

	// 执行批量操作
	res, err := es.Bulk(bytes.NewReader(buf.Bytes()), es.Bulk.WithIndex(index), es.Bulk.WithContext(context.Background()))

	if err != nil {
		fmt.Println("Error performing bulk request: ", err)
		return nil, err
	}

	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}
	return res, err
}

func Es8Delete(urls []string, username string, password string, apiKey string, index string, id string) (*esapi.Response, error) {

	// 创建配置
	cfg := elasticsearch.Config{
		Addresses: urls,
		Username:  username,
		Password:  password,
		APIKey:    apiKey,
	}

	// 创建客户端
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("创建elasticsearch客户端失败: ", err)
		return nil, err
	}

	res, err := es.Delete(index, id, es.Delete.WithContext(context.Background()))

	return res, err
}

func Es8Clear(urls []string, username string, password string, apiKey string, index string) (*esapi.Response, error) {

	// 创建配置
	cfg := elasticsearch.Config{
		Addresses: urls,
		Username:  username,
		Password:  password,
		APIKey:    apiKey,
	}

	// 创建客户端
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("创建elasticsearch客户端失败: ", err)
		return nil, err
	}

	// 构建删除查询
	query := `{
		"query": {
			"match_all": {}
		}
	}`

	// 执行 _delete_by_query API
	res, err := es.DeleteByQuery(
		[]string{index},
		strings.NewReader(query),
		es.DeleteByQuery.WithContext(context.Background()),
	)

	if err != nil {
		fmt.Println("Error deleting documents by query: ", err)
		return nil, err
	}
	defer res.Body.Close()

	// 打印删除结果
	if res.IsError() {
		fmt.Println("Error response: ", res.String())
		return nil, errors.New(res.String())
	}

	return res, err

}

func Es8GetVector(urls []string, username string, password string, apiKey string, index string, id string) (*Hit[Es8VectorDocument], error) {
	// 创建配置
	cfg := elasticsearch.Config{
		Addresses: urls,
		Username:  username,
		Password:  password,
		APIKey:    apiKey,
	}

	// 创建客户端
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("创建elasticsearch客户端失败: ", err)
		return nil, err
	}

	//读取文档
	res, err := es.Get(index, id)
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		fmt.Println("Error response:", res.String())
		return nil, errors.New("Error response:" + res.String())
	}

	b, err := io.ReadAll(res.Body)

	fmt.Println(string(b))

	var hit Hit[Es8VectorDocument]

	if err := json.Unmarshal(b, &hit); err != nil {
		fmt.Println("Error parsing the response body: ", err)
		return nil, err
	}

	return &hit, nil
}

func Es8SearchVectors(urls []string, username string, password string, apiKey string, index string, vector []float64, distance string, limit int) (*Hits[Es8VectorDocument], error) {
	var dim_size int

	indexInfo, err := Es8GetIndex(urls, username, password, apiKey, index)
	if err != nil {
		return nil, err
	}

	dim_size = int(indexInfo.Mappings.Properties.Vector.Dims)

	// 创建配置
	cfg := elasticsearch.Config{
		Addresses: urls,
		Username:  username,
		Password:  password,
		APIKey:    apiKey,
	}

	// 创建客户端
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("创建elasticsearch客户端失败: ", err)
		return nil, err
	}

	data, err := json.Marshal(vector)

	if err != nil {
		fmt.Println("向量格式错误: ", err)
		return nil, err
	}

	// 余弦距离, 归一化[0,1]
	distance_source := `(cosineSimilarity(params.query_vector, 'vector') + 1.0) / 2.0`

	// 欧式距离, 归一化[0,1]
	if distance == "Euclid" {
		distance_source = `1 / (1 + l2norm(params.query_vector, 'vector'))`

	}
	// 点式距离, 归一化[0,1]
	if distance == "Dot" {
		max := dim_size
		min := 0
		distance_source = fmt.Sprintf(`(dotProduct(params.query_vector, 'vector') - %v) / (%v - %v) `, min, max, min)
	}

	// 创建查询
	query := fmt.Sprintf(`{
		"query": {
			"script_score": {
				"query": {
					"match_all": {}
				},
				"script": {
					"source": "%s",
					"params": {
						"query_vector": %v
					}
				}
			}
		}
	}`, distance_source, string(data))

	// 执行搜索请求
	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex(index),
		es.Search.WithBody(strings.NewReader(query)),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)

	if err != nil {
		fmt.Println("Error getting response: ", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		fmt.Println("Error response:", res.String())
		return nil, errors.New(res.String())
	}

	b, err := io.ReadAll(res.Body)

	if err != nil {
		fmt.Println("Error getting response: ", err)
		return nil, err
	}

	var hits Hits[Es8VectorDocument]

	if err := json.Unmarshal(b, &hits); err != nil {
		fmt.Println("Error parsing the response body: ", err)
		return nil, err
	}

	return &hits, nil
}
