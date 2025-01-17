package qdrant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type QdrantCollectionInfo struct {
	Status        string `json:"status"`
	VectorsCount  int    `json:"vectors_count"`
	SegmentsCount int    `json:"segments_count"`
	DiskDataSize  int    `json:"disk_data_size"`
	RamDataSize   int    `json:"ram_data_size"`
}
type QdrantResponse[T any] struct {
	Time   float64 `json:"time"`
	Status any     `json:"status"`
	Result T       `json:"result"`
}

type QdrantPoint struct {
	Id      string    `json:"id"`
	Payload any       `json:"payload"`
	Vector  []float64 `json:"vector"`
	Score   float64   `json:"score"`
}

// 创建集合
func QdrantPutCollection(ip string, port string, collectionName string, size int, distance string) (*QdrantResponse[bool], error) {
	if size == 0 {
		size = 1536
	}
	if len(distance) == 0 {
		distance = "Cosine"
	}
	url := fmt.Sprintf("http://%s:%s/collections/%s", ip, port, collectionName)
	requestBody, err := json.Marshal(map[string]interface{}{
		"name": collectionName,
		"vectors": map[string]interface{}{
			"size":     size,
			"distance": distance,
		},
	})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	var response QdrantResponse[bool]
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// 删除集合
func QdrantDeleteCollection(ip string, port string, collectionName string) (*QdrantResponse[bool], error) {
	url := fmt.Sprintf("http://%s:%s/collections/%s", ip, port, collectionName)

	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	var response QdrantResponse[bool]
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// 查询集合信息
func QdrantGetCollection(ip string, port string, collectionName string) (*QdrantResponse[QdrantCollectionInfo], error) {
	url := fmt.Sprintf("http://%s:%s/collections/%s", ip, port, collectionName)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)

	var response QdrantResponse[QdrantCollectionInfo]
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// 查询集合信息
func QdrantGetPoint(ip string, port string, collectionName string, id string) (*QdrantResponse[QdrantPoint], error) {
	url := fmt.Sprintf("http://%s:%s/collections/%s", ip, port, collectionName)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)

	var response QdrantResponse[QdrantPoint]
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// 查询集合信息
func QdrantDeletePoint(ip string, port string, collectionName string, id string) (*QdrantResponse[any], error) {
	url := fmt.Sprintf("http://%s:%s/collections/%s/points/%s", ip, port, collectionName, id)

	// 发送请求
	b := []byte{}
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to PUT points to collection %s, status code: %d", collectionName, resp.StatusCode)
	}

	result, _ := io.ReadAll(resp.Body)

	var response QdrantResponse[any]
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// 增加向量数据
func QdrantPutPoints(ip string, port string, collectionName string, points []QdrantPoint) (*QdrantResponse[any], error) {
	url := fmt.Sprintf("http://%s:%s/collections/%s/points?wait=true", ip, port, collectionName)

	// 构造请求体
	requestBody := map[string]interface{}{
		"points": points,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// 发送请求
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(requestBodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	var response QdrantResponse[any]
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// 搜索向量数据
func QdrantSearchPoints(ip string, port string, collectionName string, vector []float64, score_threshold float64, limit int) (*QdrantResponse[[]QdrantPoint], error) {
	// 构造请求体
	requestBody := map[string]interface{}{
		"params": map[string]interface{}{
			"hnsw_ef":      0,
			"exact":        false,
			"quantization": nil,
			"indexed_only": false,
		},
		"vector":          vector,
		"limit":           limit,
		"with_payload":    true,
		"score_threshold": score_threshold,
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// 构造请求
	url := fmt.Sprintf("http://%s:%s/collections/%s/points/search", ip, port, collectionName)
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := http.DefaultClient
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 处理响应
	result, err := io.ReadAll(resp.Body)

	var response QdrantResponse[[]QdrantPoint]
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
