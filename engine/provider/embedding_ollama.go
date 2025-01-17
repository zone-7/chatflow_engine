package provider

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/zone-7/chatflow_engine/engine/provider/ollama"
	"github.com/zone-7/chatflow_engine/engine/utils"
)

func init() {
	e := Embedding_ollama{}
	embeddings = append(embeddings, e.GetDict().Name)
}

// ollama向量服务
type Embedding_ollama struct {
	Url       string `json:"url" yaml:"url"`
	Model     string `json:"model" yaml:"model"`
	Timeout   int64  `json:"timeout" yaml:"timeout"`
	KeepAlive string `json:"keep_alive" yaml:"keep_alive"`
}

func (e *Embedding_ollama) GetDict() Dict {
	dict := Dict{}
	dict.Name = "ollama"

	dict.Fields = []Field{}
	dict.Fields = append(dict.Fields, Field{Name: "url"})
	dict.Fields = append(dict.Fields, Field{Name: "model"})
	dict.Fields = append(dict.Fields, Field{Name: "timeout"})

	return dict
}

func (e *Embedding_ollama) Embed(params map[string]string, contents []string) ([][]float64, error) {
	var result []float64

	for k, v := range params {
		if k == "url" {
			e.Url = v
		}
		if k == "model" {
			e.Model = v
		}
		if k == "keep_alive" {
			e.KeepAlive = v
		}
		if k == "timeout" {
			e.Timeout, _ = utils.StringToInt64(v)
		}
	}

	results := make([][]float64, 0)

	for index, content := range contents {
		log.Println(fmt.Sprintf("embedding:%v/%v", index+1, len(contents)))

		request := ollama.EmbeddingRequest{}

		//请求内容
		request.Prompt = content

		//Model
		request.Model = e.Model

		request.KeepAlive = e.KeepAlive

		header := make(map[string]string)
		header["Content-Type"] = "application/json"

		var response *ollama.EmbeddingResponse

		if e.Timeout <= 0 {
			e.Timeout = 30000
		}

		//请求

		url := e.Url
		if !strings.Contains(e.Url, "/api/embeddings") {
			url = e.Url + "/api/embeddings"
		}
		err := ollama.Embedding(url, request, header, e.Timeout, func(re ollama.EmbeddingResponse, finish bool) error {

			if response == nil {
				response = &re
			}
			return nil
		})

		if err != nil {
			msg := fmt.Sprintf("Ollama embedding执行异常:%v", err.Error())
			log.Println(msg)
			return nil, errors.New(msg)
		}

		if response == nil {
			msg := fmt.Sprintf("Ollama embedding执行异常:%v", "response empty")
			log.Println(msg)
			return nil, errors.New(msg)
		}
		if response.Error.Code != nil {
			msg := fmt.Sprintf("Ollama embedding执行异常:%v", response.Error.Message)
			log.Println(msg)
			return nil, errors.New(msg)
		}

		result = response.Embedding

		results = append(results, result)
	}

	return results, nil

}
