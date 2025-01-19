package flow

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/meta"
	"github.com/zone-7/chatflow_engine/engine/provider"
	"github.com/zone-7/chatflow_engine/engine/utils"
)

func init() {
	andflow.RegistActionRunner("ollama_chat", &OllamaChatRunner{})
}

type OllamaChatRunner struct {
	BaseRunner
}

func (r *OllamaChatRunner) Properties() []andflow.Prop {
	return []andflow.Prop{}
}
func (r *OllamaChatRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {
	action := s.GetFlow().GetAction(param.ActionId)
	chatSession := r.GetChatSession(s)

	log.Printf("ollama begin: %v", time.Now())
	defer log.Printf("ollama end: %v", time.Now())

	prop, err := r.GetActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	content_source := prop["content_source"] //信息来自参数还是输入
	content_temp := prop["content_temp"]     //信息来自哪个参数

	image_source := prop["image_source"] //图片信息来自参数还是输入
	image_temp := prop["image_temp"]     //图片信息来自哪个参数

	res_chat := prop["res_chat"] //输出到对话

	url := prop["url"]
	param_key := prop["param_key"] //返回内容
	his_count := prop["his_count"]
	his_time := prop["his_time"]

	req_cos := prop["req_cosplay"]
	req_api_key := prop["req_api_key"]
	req_stream := prop["req_stream"] //流式传输

	req_max_tokens := prop["req_max_tokens"]
	req_keep_alive := prop["req_keep_alive"]

	req_top_p := prop["req_top_p"]
	req_temperature := prop["req_temperature"]

	req_model := prop["req_model"]
	req_model_other := prop["req_model_other"]
	if len(req_model) == 0 {
		req_model = req_model_other
	}

	req_user_contract := prop["req_user_contract"]

	//地址
	if len(url) == 0 {
		return andflow.RESULT_FAILURE, errors.New("参数 URL 地址不能为空")
	}

	requestContent := ""
	if content_source == "temp" && len(content_temp) > 0 {
		requestContent = content_temp
	} else {
		msgs := chatSession.GetCurrentRequestMessages(1)
		if msgs != nil && len(msgs) > 0 {
			for _, msg := range msgs {
				requestContent += msg.Content
			}
		}

	}

	attachImages := make([]string, 0)
	if image_source == "temp" && len(image_temp) > 0 {
		err = json.Unmarshal([]byte(image_temp), &attachImages)
		if err != nil {
			return andflow.RESULT_FAILURE, errors.New("图片格式错误")
		}
	} else {
		msgs := chatSession.GetCurrentRequestMessages(1)
		if msgs != nil && len(msgs) > 0 {
			for _, msg := range msgs {
				attachImages = append(attachImages, msg.Images...)
			}
		}
	}
	requestImages := make([]string, 0)
	for _, img := range attachImages {
		start := strings.Index(img, "data:image/")
		end := strings.Index(img, "base64,")

		if start >= 0 && end > 0 {
			img = img[end+len("base64,"):]
		}

		requestImages = append(requestImages, img)
	}

	//max token
	max_leng := 0
	if len(requestContent) > 0 {
		max_leng, _ = strconv.Atoi(req_max_tokens)
	}

	if max_leng > 0 && len(requestContent) > max_leng {
		return andflow.RESULT_FAILURE, errors.New("内容太长，超出" + req_max_tokens + "限制")
	}

	//请求内容
	messages := make([]provider.ChatMessage, 0)

	//设置系统面具定义
	cosplay := ""
	if len(req_cos) > 0 {
		cosplay = req_cos
	}
	messages = append(messages, provider.ChatMessage{Role: provider.MESSAGE_ROLE_SYSTEM, Content: cosplay})

	//设置用户背景要求
	if len(req_user_contract) > 0 {
		messages = append(messages, provider.ChatMessage{Role: provider.MESSAGE_ROLE_USER, Content: req_user_contract})
	}

	// 历史消息数量限制
	var history_count int
	if len(his_count) > 0 {
		history_count, _ = strconv.Atoi(his_count)
	} else {
		history_count = 4
	}

	//历史消息的时间限制
	var history_time int64
	if len(his_time) > 0 {
		history_time, _ = utils.StringToInt64(his_time)
	} else {
		history_time = 0
	}

	//加载历史消息
	historyChatMessages := chatSession.GetMessages()

	// 历史消息
	history_msgs := make([]provider.ChatMessage, 0)

	for i := len(historyChatMessages) - 1; i >= 0 && len(history_msgs) < history_count; i-- {
		m := historyChatMessages[i]

		//时间限制
		if history_time > 0 {
			if time.Now().UnixNano()-m.SendTime > history_time {
				continue
			}
		}

		item := provider.ChatMessage{Role: m.Role, Content: m.Content, Images: m.Images}

		history_msgs = append([]provider.ChatMessage{item}, history_msgs...)
	}

	if len(history_msgs) > 0 {
		messages = append(messages, history_msgs...)
	}

	//消息
	requestMsg := provider.ChatMessage{Role: provider.MESSAGE_ROLE_USER, Content: requestContent, Images: requestImages}
	messages = append(messages, requestMsg)

	// 参数
	params := map[string]string{}
	params["timeout"] = s.GetFlow().Timeout
	params["url"] = url
	params["api_key"] = req_api_key
	params["model"] = req_model
	params["stream"] = req_stream
	params["temperature"] = req_temperature
	params["top_p"] = req_top_p
	params["max_tokens"] = req_max_tokens
	params["keep_alive"] = req_keep_alive

	if !strings.Contains(params["url"], "/api/chat") {
		params["url"] = url + "/api/chat"
	}

	uid, _ := uuid.NewV4()
	mid := strings.ReplaceAll(uid.String(), "-", "")

	responseContent := ""
	responseImages := make([]string, 0)
	// 对话
	chatting := provider.CreateChatting("ollama")

	err = chatting.Chat(params, messages, func(msg []provider.ChatMessage, is_done bool) error {
		content := ""
		images := make([]string, 0)
		for _, m := range msg {
			content += m.Content
			if m.Images != nil {
				images = append(images, m.Images...)
			}

		}

		responseContent += content
		responseImages = append(responseImages, images...)

		//实时返回到对话
		if res_chat == "true" || res_chat == "1" {

			if len(content) > 0 || len(images) > 0 {
				chatSession.Response(meta.ChatFlowMessage{MessageId: mid, MessageType: meta.CHAT_MESSAGE_TYPE_MESSAGE, Format: meta.CHAT_MESSAGE_FORMAT_TEXT, Role: meta.CHAT_MESSAGE_ROLE_ASSISTANT, Content: content, Images: images, Finish: "no"}, true)
			}

			if is_done {
				chatSession.Response(meta.ChatFlowMessage{MessageId: mid, MessageType: meta.CHAT_MESSAGE_TYPE_MESSAGE, Format: meta.CHAT_MESSAGE_FORMAT_TEXT, Role: meta.CHAT_MESSAGE_ROLE_ASSISTANT, Content: "", Images: nil, Finish: "yes"}, true)
			}
		}

		return nil

	}, func() bool {
		return s.Operation.GetCmd() == andflow.CMD_STOP
	})

	if err != nil {
		log.Printf("Ollama执行异常:%v", err)
		return andflow.RESULT_FAILURE, err
	}

	//保存返回内容到参数
	if len(param_key) > 0 {
		s.SetParam(param_key, responseContent)
	}

	return andflow.RESULT_SUCCESS, nil
}
