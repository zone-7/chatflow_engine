package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	CHAT_MESSAGE_ROLE_USER      = "user"
	CHAT_MESSAGE_ROLE_ASSISTANT = "assistant"
	CHAT_MESSAGE_ROLE_SYSTEM    = "system"
)

type ChatMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images"`
	Partial bool     `json:"partial"`
}

type ChatRequest struct {
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

type ChatResponse struct {
	Model     string      `json:"model"`
	CreatedAt string      `json:"created_at"`
	Message   ChatMessage `json:"message"`
	Done      bool        `json:"done"`
	Error     struct {
		Message string      `json:"message"`
		Code    interface{} `json:"code"`
	} `json:"error"`
	Code int `json:"code"`
}

// 请求
func Chat(url string, gptReq ChatRequest, headers map[string]string, timeout int64, callback func(gptRes ChatResponse, finish bool) error, isStop func() bool) error {
	// http://localhost:11434/api/chat
	if len(gptReq.KeepAlive) == 0 {
		gptReq.KeepAlive = "5m"
	}

	request, err := json.Marshal(gptReq)
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

	stream := gptReq.Stream

	// 读取响应
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	//流输出
	if stream {
		reader := bufio.NewReader(resp.Body)

		for {
			var gptRes ChatResponse
			isFinish := false

			if isStop != nil && isStop() {
				isFinish = true
			}

			line, _, err := reader.ReadLine()
			if err != nil { //结束
				break
			}

			lineStr := string(line)

			lineStr = strings.Trim(lineStr, " ")

			//正常流消息，应该以 data: 开始
			if strings.Index(lineStr, "data:") >= 0 {
				lineStr = strings.Replace(lineStr, "data:", "", 1)
				lineStr = strings.Trim(lineStr, " ")
			}
			if len(lineStr) == 0 {
				continue
			}

			//错误消息，结束标示
			if strings.Index(lineStr, "\"message\":\"EOF\"") >= 0 {
				continue
			}

			//正常消息，结束标示
			if strings.Index(lineStr, "[DONE]") >= 0 {
				break
			}
			//是否最后一条
			if strings.Index(lineStr, "\"finish_reason\":\"stop\"") >= 0 {
				isFinish = true
			}

			// 如果可以直接解析
			err = json.Unmarshal([]byte(lineStr), &gptRes)
			if err == nil {
				if gptRes.Done {
					isFinish = true
				}

				err = callback(gptRes, isFinish)
				if err != nil {
					return err
				}
				if isFinish {
					break
				}

			}

		}

	} else {

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var gptRes ChatResponse

		err = json.Unmarshal(data, &gptRes)
		if err != nil {
			return err
		}
		err = callback(gptRes, true)
		if err != nil {
			return err
		}

	}

	return nil
}
