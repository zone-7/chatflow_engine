package provider

import (
	"github.com/zone-7/chatflow_engine/engine/provider/ollama"
	"github.com/zone-7/chatflow_engine/engine/utils"
)

func init() {
	c := Chatting_ollama{}
	chattings = append(chattings, c.GetDict().Name)
}

type Chatting_ollama struct {
	Url         string  `json:"url" yaml:"url"`
	Model       string  `json:"model" yaml:"model"`
	Stream      bool    `json:"stream" yaml:"stream"`
	Seed        int     `json:"seed" yaml:"seed"`
	Temperature float64 `json:"temperature" yaml:"temperature"`
	TopP        int     `json:"top_p" yaml:"top_p"`
	Timeout     int64   `json:"timeout" yaml:"timeout"`
	KeepAlive   string  `json:"keep_alive" yaml:"keep_alive"`
}

func (c *Chatting_ollama) GetDict() Dict {
	dict := Dict{}
	dict.Name = "ollama"
	return dict
}

func (c *Chatting_ollama) Chat(params map[string]string, messages []ChatMessage, callback func(msg []ChatMessage, is_done bool) error, is_suspend func() bool) error {
	var err error

	for k, v := range params {

		if k == "url" {
			c.Url = v
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
		if k == "keep_alive" {
			c.KeepAlive = v
		}

		if k == "temperature" {
			c.Temperature, _ = utils.StringToFloat64(v)
		}

		if k == "top_p" {
			c.TopP, _ = utils.StringToInt(v)
		}
		if k == "seed" {
			c.Seed, _ = utils.StringToInt(v)
		}

		if k == "timeout" {
			c.Timeout, _ = utils.StringToInt64(v)
		}

	}

	request := ollama.ChatRequest{}
	request.Messages = make([]ollama.ChatMessage, 0)

	for _, m := range messages {
		request.Messages = append(request.Messages, ollama.ChatMessage{Role: m.Role, Content: m.Content, Images: m.Images, Partial: m.Partial})
	}

	request.Model = c.Model
	request.Stream = c.Stream
	request.KeepAlive = c.KeepAlive
	request.Options.Seed = c.Seed
	request.Options.Temperature = c.Temperature
	request.Options.TopP = c.TopP

	header := make(map[string]string)
	header["Content-Type"] = "application/json"

	err = ollama.Chat(c.Url, request, header, c.Timeout, func(gptRes ollama.ChatResponse, finish bool) error {

		msg := ChatMessage{}
		msg.Content = gptRes.Message.Content
		msg.Images = gptRes.Message.Images
		msg.Role = gptRes.Message.Role

		suberr := callback([]ChatMessage{msg}, finish)

		return suberr

	}, func() bool {
		return is_suspend()
	})

	return err
}
