package provider

import (
	"errors"
	"log"

	"github.com/zone-7/chatflow_engine/engine/provider/openai"
	"github.com/zone-7/chatflow_engine/engine/utils"
)

func init() {
	c := Chatting_openai{}
	chattings = append(chattings, c.GetDict().Name)
}

type Chatting_openai struct {
	Url         string  `json:"url" yaml:"url"`
	ApiKey      string  `json:"api_key" yaml:"api_key"`
	Model       string  `json:"model" yaml:"model"`
	Stream      bool    `json:"stream" yaml:"stream"`
	Temperature float64 `json:"temperature" yaml:"temperature"`
	TopP        float64 `json:"top_p" yaml:"top_p"`
	MaxTokens   int     `json:"max_tokens" yaml:"max_tokens"`
	N           int     `json:"n" yaml:"n"`
	User        string  `json:"user" yaml:"user"`
	Timeout     int64   `json:"timeout" yaml:"timeout"`
}

func (c *Chatting_openai) GetDict() Dict {
	dict := Dict{}
	dict.Name = "openai"
	return dict
}

func (c *Chatting_openai) Chat(params map[string]string, messages []ChatMessage, callback func(msg []ChatMessage, is_done bool) error, is_suspend func() bool) error {
	var err error

	for k, v := range params {

		if k == "url" {
			c.Url = v
		}
		if k == "api_key" {
			c.ApiKey = v
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

		if k == "top_p" {
			c.TopP, _ = utils.StringToFloat64(v)
		}
		if k == "max_tokens" {
			c.MaxTokens, _ = utils.StringToInt(v)
		}
		if k == "n" {
			c.N, _ = utils.StringToInt(v)
		}
		if k == "user" {
			c.User = v
		}

		if k == "timeout" {
			c.Timeout, _ = utils.StringToInt64(v)
		}

	}

	request := openai.ChatRequest{}
	request.Messages = make([]openai.ChatMessage, 0)

	for _, m := range messages {
		request.Messages = append(request.Messages, openai.ChatMessage{Role: m.Role, Content: m.Content, Images: m.Images, Partial: m.Partial})
	}

	request.Model = c.Model
	request.Stream = c.Stream
	request.MaxTokens = c.MaxTokens
	request.TopP = c.TopP
	request.Temperature = c.Temperature
	request.N = c.N
	request.User = c.User

	header := make(map[string]string)
	header["Authorization"] = "Bearer " + c.ApiKey
	header["Content-Type"] = "application/json"

	err = openai.Chat(c.Url, request, header, c.Timeout, func(gptRes openai.ChatResponse, finish bool) error {

		if gptRes.Error != nil && len(gptRes.Error.Message) > 0 {
			log.Println(errors.New(gptRes.Error.Message))
			return errors.New(gptRes.Error.Message)
		}

		//实时返回到对话
		content := ""
		for _, cookie := range gptRes.Choices {
			if cookie.Message != nil && len(cookie.Message.Content) > 0 {
				content = content + cookie.Message.Content
			}
			if cookie.Delta != nil && len(cookie.Delta.Content) > 0 {
				content = content + cookie.Delta.Content
			}
		}

		if len(gptRes.Detail) > 0 {
			content += gptRes.Detail
		}
		if len(gptRes.Message) > 0 {
			content += gptRes.Message
		}

		msg := ChatMessage{}

		msg.Content = content
		msg.Role = "assistant"

		suberr := callback([]ChatMessage{msg}, finish)

		return suberr

	}, func() bool {
		return is_suspend()
	})

	return err
}
