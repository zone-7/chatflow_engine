package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	CHATGPT_MESSAGE_ROLE_USER      = "user"
	CHATGPT_MESSAGE_ROLE_ASSISTANT = "assistant"
	CHATGPT_MESSAGE_ROLE_SYSTEM    = "system"
)

type ChatChoice struct {
	Message      *ChatMessage `json:"message"`
	Delta        *ChatMessage `json:"delta"`
	Index        int          `json:"index"`
	FinishReason string       `json:"finish_reason"`
}

type ChatContent struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	ImageUrl struct {
		Url    string `json:"url"`
		Detail string `json:"detail"`
	} `json:"image_url"`
}

type ChatMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images"`
	Partial bool     `json:"partial"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`

	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	N           int     `json:"n"`
	Stream      bool    `json:"stream"`
	User        string  `json:"user"`
}
type ChatError struct {
	Message string      `json:"message"`
	Code    interface{} `json:"code"`
}
type ChatResponse struct {
	Id      string       `json:"id"`
	Object  string       `json:"object"`
	Model   string       `json:"model"`
	Created int64        `json:"created"`
	Choices []ChatChoice `json:"choices"`

	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`

	Error *ChatError `json:"error"`

	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

// 请求openai
func Chat(url string, gptReq ChatRequest, headers map[string]string, timeout int64, callback func(gptRes ChatResponse, finish bool) error, isStop func() bool) error {

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

	//流输出
	if stream {
		reader := bufio.NewReader(resp.Body)
		lines := make([]string, 0)

		for {
			var gptRes ChatResponse
			isFinish := false

			if isStop != nil && isStop() {
				isFinish = true
			}

			line, _, err := reader.ReadLine()
			if err != nil { //结束
				fmt.Println(err)
				break
			}

			lineStr := string(line)
			fmt.Println(lineStr)

			lineStr = strings.Trim(lineStr, " ")

			if len(lineStr) == 0 {
				continue
			}

			// 如果可以直接解析
			err = json.Unmarshal([]byte(lineStr), &gptRes)
			if err == nil {
				err = callback(gptRes, isFinish)
				if err != nil {
					return err
				}

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

			lines = append(lines, lineStr)

			//正常流消息，应该以 data: 开始
			if strings.Index(lineStr, "data:") < 0 {
				continue
			}

			lineStr = strings.Replace(lineStr, "data:", "", 1)
			lineStr = strings.Trim(lineStr, " ")
			if len(lineStr) == 0 {
				continue
			}

			//是否最后一条
			if strings.Index(lineStr, "\"finish_reason\":\"stop\"") >= 0 {
				isFinish = true
			}

			err = json.Unmarshal([]byte(lineStr), &gptRes)
			if err != nil {
				continue
			}

			err = callback(gptRes, isFinish)
			if err != nil {
				return err
			}

			if isFinish {
				break
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
			var errorInfo ChatError
			err = json.Unmarshal(data, &errorInfo)
			if err != nil {
				return err
			}

			gptRes.Error = &errorInfo
		}

		err = callback(gptRes, true)
		if err != nil {
			return err
		}

	}

	return nil
}
