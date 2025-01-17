package baidu

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
	ERNIE_MESSAGE_ROLE_USER      = "user"
	ERNIE_MESSAGE_ROLE_ASSISTANT = "assistant"
	ERNIE_MESSAGE_ROLE_SYSTEM    = "system"
)

type ErnieAccessToken struct {
	AccessToken   string `json:"access_token"`
	ExpiresIn     int64  `json:"expires_in"`
	RefreshToken  string `json:"refresh_token"`
	SessionKey    string `json:"session_key"`
	SessionSecret string `json:"session_secret"`
	Scope         string `json:"scope"`
}
type ErnieMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ErnieRequest struct {
	Messages []ErnieMessage `json:"messages"`
	/*
		较高的数值会使输出更加随机，而较低的数值会使其更加集中和确定
		（2）默认0.95，范围 (0, 1.0]，不能为0
		（3）建议该参数和top_p只设置1个
		（4）建议top_p和temperature不要同时更改
	*/
	Temperature float64 `json:"temperature"`

	/*
		（1）影响输出文本的多样性，取值越大，生成文本的多样性越强
		（2）默认0.8，取值范围 [0, 1.0]
		（3）建议该参数和temperature只设置1个
		（4）建议top_p和temperature不要同时更改
	*/
	TopP float64 `json:"top_p"`

	/*
			通过对已生成的token增加惩罚，减少重复生成的现象。说明：
		（1）值越大表示惩罚越大
		（2）默认1.0，取值范围：[1.0, 2.0]
	*/
	PenaltyScore float64 `json:"penalty_score"`

	/*
		是否以流式接口的形式返回数据，默认false
	*/
	Stream bool   `json:"stream"`
	UserId string `json:"user_id"`
}

type ErnieResponse struct {
	Id               string `json:"id"`
	Object           string `json:"object"`
	Created          int64  `json:"created"`
	SentenceId       int64  `json:"sentence_id"`
	IsEnd            bool   `json:"is_end"`
	IsTruncated      bool   `json:"is_truncated"`
	NeedClearHistory bool   `json:"need_clear_history"` //表示用户输入是否存在安全
	Result           string `json:"result"`

	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`

	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_msg"`
}

func GetErnieAccessToken(api_key string, secret_key string) (*ErnieAccessToken, error) {
	url := "https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=" + api_key + "&client_secret=" + secret_key
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}
	// 发送请求
	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Body == nil {
		return nil, errors.New("获取百度API令牌失败")
	}

	// 读取响应
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	var token ErnieAccessToken
	err = json.Unmarshal(data, &token)

	if err != nil {
		return nil, err
	}
	return &token, nil

}

func Chat(req_service string, accessToken string, request ErnieRequest, headers map[string]string, timeout int64, callback func(res *ErnieResponse) error, isStop func() bool) error {
	if len(req_service) == 0 {
		req_service = "completions"
	}
	url := "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/" + req_service + "?access_token=" + accessToken

	data, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	// 设置请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("x-bce-date", time.Now().String())

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

	//流输出
	if request.Stream {
		reader := bufio.NewReader(resp.Body)
		for {
			line, _, err := reader.ReadLine()
			if err != nil { //结束
				break
			}

			lineStr := string(line)
			lineStr = strings.Trim(lineStr, " ")

			if len(lineStr) == 0 {
				continue
			}
			lineStr = strings.Replace(lineStr, "data:", "", 1)
			lineStr = strings.Trim(lineStr, " ")

			if strings.Index(lineStr, "[DONE]") >= 0 { //最后一条
				break
			}

			if len(lineStr) == 0 {
				continue
			}

			var response ErnieResponse

			err = json.Unmarshal([]byte(lineStr), &response)
			if err != nil { //结束
				continue
			}

			if isStop != nil && isStop() {
				response.IsEnd = true
			}

			if callback != nil {
				err = callback(&response)
				if err != nil {
					return err
				}
			}

			if response.IsEnd {
				break
			}

		}

	} else {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		var response ErnieResponse

		err = json.Unmarshal(data, &response)
		if err != nil {
			return err
		}

		if callback != nil {
			err = callback(&response)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
