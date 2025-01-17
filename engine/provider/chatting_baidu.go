package provider

import (
	"errors"
	"log"

	"github.com/zone-7/chatflow_engine/engine/provider/baidu"
	"github.com/zone-7/chatflow_engine/engine/utils"
)

func init() {
	c := Chatting_baidu{}
	chattings = append(chattings, c.GetDict().Name)
}

type Chatting_baidu struct {
	Url       string `json:"url" yaml:"url"`
	Service   string `json:"service" yaml:"service"`
	ApiKey    string `json:"api_key" yaml:"api_key"`
	SecretKey string `json:"secret_key" yaml:"secret_key"`

	Model string `json:"model" yaml:"model"`

	Temperature float64 `json:"temperature" yaml:"temperature"`
	/*
		（1）影响输出文本的多样性，取值越大，生成文本的多样性越强
		（2）默认0.8，取值范围 [0, 1.0]
		（3）建议该参数和temperature只设置1个
		（4）建议top_p和temperature不要同时更改
	*/
	TopP float64 `json:"top_p" yaml:"top_p"`

	/*
			通过对已生成的token增加惩罚，减少重复生成的现象。说明：
		（1）值越大表示惩罚越大
		（2）默认1.0，取值范围：[1.0, 2.0]
	*/
	PenaltyScore float64 `json:"penalty_score" yaml:"penalty_score"`

	/*
		是否以流式接口的形式返回数据，默认false
	*/
	Stream bool   `json:"stream" yaml:"stream"`
	UserId string `json:"user_id" yaml:"user_id"`

	Timeout int64 `json:"timeout" yaml:"timeout"`
}

func (c *Chatting_baidu) GetDict() Dict {
	dict := Dict{}
	dict.Name = "baidu"
	return dict
}

func (c *Chatting_baidu) Chat(params map[string]string, messages []ChatMessage, callback func(msg []ChatMessage, is_done bool) error, is_suspend func() bool) error {
	var err error

	for k, v := range params {

		if k == "url" {
			c.Url = v
		}
		if k == "service" {
			c.Service = v
		}

		if k == "api_key" {
			c.ApiKey = v
		}
		if k == "secret_key" {
			c.SecretKey = v
		}

		if k == "model" {
			c.Model = v
		}
		if k == "stream" {
			if v == "true" || v == "1" {
				c.Stream = true
			} else {
				c.Stream = false
			}
		}

		if k == "temperature" {
			c.Temperature, _ = utils.StringToFloat64(v)
		}
		if c.Temperature <= 0 {
			c.Temperature = 0.95
		}
		if c.Temperature > 1 {
			c.Temperature = 1
		}

		if k == "penalty_score" {
			c.PenaltyScore, _ = utils.StringToFloat64(v)
		}
		if c.PenaltyScore < 1 {
			c.PenaltyScore = 1
		}
		if c.PenaltyScore > 2 {
			c.PenaltyScore = 2
		}

		if k == "top_p" {
			c.TopP, _ = utils.StringToFloat64(v)
		}
		if c.TopP <= 0 {
			c.TopP = 0.8
		}
		if c.TopP > 1 {
			c.TopP = 1
		}

		if k == "user_id" {
			c.UserId = v
		}

		if k == "timeout" {
			c.Timeout, _ = utils.StringToInt64(v)
		}

	}

	request := baidu.ErnieRequest{}
	request.Messages = make([]baidu.ErnieMessage, 0)

	for _, m := range messages {
		request.Messages = append(request.Messages, baidu.ErnieMessage{Role: m.Role, Content: m.Content})
	}

	request.Stream = c.Stream

	request.UserId = c.UserId
	request.Temperature = c.Temperature
	request.PenaltyScore = c.PenaltyScore
	request.TopP = c.TopP

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = "Bearer " + c.ApiKey

	accessToken, err := baidu.GetErnieAccessToken(c.ApiKey, c.SecretKey)

	if err != nil {
		log.Println(err)
		return errors.New("获取百度令牌失败")
	}
	if accessToken == nil {
		return errors.New("获取百度令牌失败")
	}

	err = baidu.Chat(c.Service, accessToken.AccessToken, request, header, c.Timeout, func(re *baidu.ErnieResponse) error {

		if re.ErrorCode > 0 {
			log.Println(re.ErrorMessage)
			return errors.New(re.ErrorMessage)
		}

		msg := ChatMessage{}

		msg.Content = re.Result
		msg.Role = "assistant"

		suberr := callback([]ChatMessage{msg}, re.IsEnd)

		return suberr

	}, func() bool {
		return is_suspend()
	})

	return err
}
