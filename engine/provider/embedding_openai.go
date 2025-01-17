package provider

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/zone-7/chatflow_engine/engine/provider/openai"
	"github.com/zone-7/chatflow_engine/engine/utils"
)

func init() {
	e := Embedding_openai{}
	embeddings = append(embeddings, e.GetDict().Name)
}

// ollama向量服务
type Embedding_openai struct {
	Url    string `json:"url" yaml:"url"`
	Model  string `json:"model" yaml:"model"`
	ApiKey string `json:"api_key" yaml:"api_key"`

	Timeout int64 `json:"timeout" yaml:"timeout"`
}

func (e *Embedding_openai) GetDict() Dict {
	dict := Dict{}
	dict.Name = "openai"

	dict.Fields = []Field{}
	dict.Fields = append(dict.Fields, Field{Name: "url"})
	dict.Fields = append(dict.Fields, Field{Name: "model"})
	dict.Fields = append(dict.Fields, Field{Name: "api_key"})

	dict.Fields = append(dict.Fields, Field{Name: "timeout"})

	return dict
}

func (e *Embedding_openai) Embed(params map[string]string, contents []string) ([][]float64, error) {
	var results [][]float64

	for k, v := range params {
		if k == "url" {
			e.Url = v
		}
		if k == "model" {
			e.Model = v
		}
		if k == "api_key" {
			e.ApiKey = v
		}

		if k == "timeout" {
			e.Timeout, _ = utils.StringToInt64(v)
		}

	}

	request := openai.EmbeddingRequest{}

	//请求内容
	request.Input = contents

	//Model
	request.Model = e.Model

	header := make(map[string]string)
	header["Authorization"] = "Bearer " + e.ApiKey
	header["Content-Type"] = "application/json"

	var response *openai.EmbeddingResponse

	if e.Timeout <= 0 {
		e.Timeout = 30000
	}

	url := e.Url
	if !strings.Contains(e.Url, "/embeddings") {
		url = e.Url + "/embeddings"
	}

	//请求
	err := openai.Embedding(url, request, header, e.Timeout, func(re openai.EmbeddingResponse, finish bool) error {

		if response == nil {
			response = &re
		}
		return nil
	})

	if err != nil {
		msg := fmt.Sprintf("Ollama embedding执行异常:%v", err.Error())
		log.Println(msg)
		return results, errors.New(msg)
	}

	if response == nil {
		msg := fmt.Sprintf("Ollama embedding执行异常:%v", "response empty")
		log.Println(msg)
		return results, errors.New(msg)
	}
	if response.Error.Code != nil || len(response.Error.Message) > 0 {
		msg := fmt.Sprintf("Ollama embedding执行异常:%v", response.Error.Message)
		log.Println(msg)
		return results, errors.New(msg)
	}

	for _, data := range response.Data {
		results = append(results, data.Embedding)
	}

	return results, nil

}
