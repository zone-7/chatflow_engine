package ollama

// http://localhost:11434/api/embeddings
// 使用Ollama 生成 Embedding

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type EmbeddingRequest struct {
	Model    string        `json:"model"`
	Prompt   string        `json:"prompt"`
	Messages []ChatMessage `json:"messages"`
	Options  struct {
		Seed        int     `json:"seed"`
		Temperature float64 `json:"temperature"`
		TopP        int     `json:"top_p"`
	} `json:"options"`

	KeepAlive string `json:"keep_alive"`
	Stream    bool   `json:"stream"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`

	Error struct {
		Message string      `json:"message"`
		Code    interface{} `json:"code"`
	} `json:"error"`
	Code int `json:"code"`
}

// 请求
func Embedding(url string, emb_req EmbeddingRequest, headers map[string]string, timeout int64, callback func(emb_res EmbeddingResponse, finish bool) error) error {
	// http://localhost:11434/api/embedding
	if len(emb_req.KeepAlive) == 0 {
		emb_req.KeepAlive = "5m"
	}

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

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var emb_res EmbeddingResponse

	err = json.Unmarshal(data, &emb_res)
	if err != nil {
		return err
	}
	err = callback(emb_res, true)
	if err != nil {
		return err
	}

	return nil
}
