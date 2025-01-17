package openai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// 向量
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
		Object    string    `json:"object"`
	} `json:"data"`
	Model  string `json:"model"`
	Object string `json:"object"`
	Usage  struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error struct {
		Message string      `json:"message"`
		Code    interface{} `json:"code"`
	} `json:"error"`
}

// 请求openai
func Embedding(url string, emb_req EmbeddingRequest, headers map[string]string, timeout int64, callback func(emb_res EmbeddingResponse, finish bool) error) error {

	request, err := json.Marshal(emb_req)
	if err != nil {
		return err
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(request))
	if err != nil {
		return err
	}
	// 设置请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	//最大超时时间控制
	if timeout <= 0 || timeout > 1000*60*10 {
		timeout = 1000 * 60 * 10 //毫秒 10分钟
	}

	tout := time.Millisecond * time.Duration(timeout)
	// 发送请求
	client := &http.Client{Timeout: tout}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp == nil || resp.Body == nil {
		return nil
	}

	// 读取响应
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var eRes EmbeddingResponse

	err = json.Unmarshal(data, &eRes)
	if err != nil {
		return err
	}
	err = callback(eRes, true)
	if err != nil {
		return err
	}

	return nil
}
