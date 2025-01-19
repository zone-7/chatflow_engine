package flow

import (
	"crypto/md5"
	"encoding/hex"
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
	"github.com/zone-7/chatflow_engine/engine/provider/baidu"
	"github.com/zone-7/chatflow_engine/engine/utils"
)

func init() {
	andflow.RegistActionRunner("baidu_ernie", &BaiduErnieRunner{})
}

type BaiduErnieRunner struct {
	BaseRunner
}

func (r *BaiduErnieRunner) Properties() []andflow.Prop {
	return []andflow.Prop{}
}
func (r *BaiduErnieRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {
	action := s.GetFlow().GetAction(param.ActionId)

	log.Printf("baidu ernie begin: %v", time.Now())
	defer log.Printf("baidu ernie end: %v", time.Now())

	prop, err := r.GetActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	content_source := prop["content_source"] //信息来自参数还是输入
	content_temp := prop["content_temp"]     //信息来自哪个参数

	image_source := prop["image_source"] //图片信息来自参数还是输入
	image_temp := prop["image_temp"]     //图片信息来自哪个参数

	res_chat := prop["res_chat"]   //输出到对话
	param_key := prop["param_key"] //返回内容存储到参数
	his_count := prop["his_count"] //携带历史记录个数
	his_time := prop["his_time"]

	req_service_other := prop["req_service_other"]
	req_service := prop["req_service"]

	req_cos := prop["req_cosplay"] //角色扮演
	req_api_key := prop["req_api_key"]
	req_secret_key := prop["req_secret_key"]
	req_penalty_score := prop["req_penalty_score"]
	req_top_p := prop["req_top_p"]
	req_temperature := prop["req_temperature"]
	req_user_contract := prop["req_user_contract"]
	req_stream := prop["req_stream"] //实时返回

	if len(req_service) == 0 {
		req_service = req_service_other
	}
	if len(req_service) == 0 {
		req_service = "completions"
	}

	//返回消息存放变量
	if len(param_key) == 0 {
		return andflow.RESULT_FAILURE, errors.New("返回消息存放地址参数KEY不能为空")
	}

	//api key
	if len(req_api_key) == 0 {
		return andflow.RESULT_FAILURE, errors.New("参数 api_key 不能为空")
	}

	//secret key
	if len(req_secret_key) == 0 {
		return andflow.RESULT_FAILURE, errors.New("参数 secret_key 不能为空")
	}

	chatSession := r.GetChatSession(s)

	requestContent := ""
	if content_source == "temp" && len(content_temp) > 0 {
		requestContent = content_temp

	} else {
		messages := chatSession.GetCurrentRequestMessages(1)
		if messages == nil || len(messages) == 0 {
			return andflow.RESULT_REJECT, nil
		}

		for _, msg := range messages {
			requestContent += msg.Content
		}
	}

	requestImages := make([]string, 0)
	if image_source == "temp" && len(image_temp) > 0 {
		err = json.Unmarshal([]byte(image_temp), &requestImages)
		if err != nil {
			return andflow.RESULT_FAILURE, errors.New("图片格式错误")
		}
	} else {
		msgs := chatSession.GetCurrentRequestMessages(1)
		if msgs != nil && len(msgs) > 0 {
			for _, msg := range msgs {
				requestImages = append(requestImages, msg.Images...)
			}
		}
	}

	//参数
	params := map[string]string{}
	params["service"] = req_service
	params["api_key"] = req_api_key
	params["secret_key"] = req_secret_key
	params["stream"] = req_stream
	params["temperature"] = req_temperature
	params["top_p"] = req_top_p
	params["penalty_score"] = req_penalty_score

	//设置用户ID
	hash := md5.Sum([]byte(s.GetRuntime().Id))
	params["user_id"] = hex.EncodeToString(hash[:])
	params["timeout"] = s.GetFlow().Timeout

	//消息列表

	messages := []provider.ChatMessage{}

	//设置系统面具定义\用户背景要求
	if len(req_cos) == 0 {
		req_cos = "？"
	}

	messages = append(messages, provider.ChatMessage{Role: provider.MESSAGE_ROLE_USER, Content: req_cos + "\n" + req_user_contract})

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
		for _, m := range history_msgs {

			if len(messages)%2 == 0 && m.Role == provider.MESSAGE_ROLE_ASSISTANT {
				messages = append(messages, provider.ChatMessage{Role: provider.MESSAGE_ROLE_USER, Content: "？"})
			}
			if len(messages)%2 != 0 && m.Role == provider.MESSAGE_ROLE_USER {
				messages = append(messages, provider.ChatMessage{Role: provider.MESSAGE_ROLE_ASSISTANT, Content: "？"})
			}

			messages = append(messages, provider.ChatMessage{Role: m.Role, Content: m.Content})
		}
	}

	//确保总的消息数是奇数
	if len(messages)%2 != 0 {
		messages = append(messages, provider.ChatMessage{Role: provider.MESSAGE_ROLE_ASSISTANT, Content: "？"})
	}

	//消息
	requestMsg := provider.ChatMessage{Role: provider.MESSAGE_ROLE_USER, Content: requestContent, Images: requestImages}
	messages = append(messages, requestMsg)

	header := make(map[string]string)
	header["Authorization"] = "Bearer " + req_api_key
	header["Content-Type"] = "application/json"

	uid, _ := uuid.NewV4()
	mid := strings.ReplaceAll(uid.String(), "-", "")

	//获取token
	accessToken, err := baidu.GetErnieAccessToken(req_api_key, req_secret_key)
	if err != nil {
		log.Println(err)
		return andflow.RESULT_FAILURE, errors.New("获取百度令牌失败")
	}
	if accessToken == nil {
		return andflow.RESULT_FAILURE, errors.New("获取百度令牌失败")
	}

	//请求文心一言

	responseContent := ""
	chatting := provider.Chatting_baidu{}
	err = chatting.Chat(params, messages, func(msg []provider.ChatMessage, is_done bool) error {

		content := ""
		for _, m := range msg {
			content += m.Content
		}

		responseContent += content

		//实时返回
		if res_chat == "true" || res_chat == "1" {
			finish := "no"
			if is_done {
				finish = "yes"
			}

			chatSession.Response(meta.ChatFlowMessage{MessageId: mid, MessageType: meta.CHAT_MESSAGE_TYPE_MESSAGE, Format: meta.CHAT_MESSAGE_FORMAT_TEXT, Role: meta.CHAT_MESSAGE_ROLE_ASSISTANT, Content: content, Finish: finish}, true)

		}

		return nil
	}, func() bool {
		return s.Operation.GetCmd() == andflow.CMD_STOP
	})

	if err != nil {
		log.Printf("baidu ernie execute error:%v", err)
		return andflow.RESULT_FAILURE, err
	}

	//保存返回内容到参数
	if len(param_key) > 0 {
		s.SetParam(param_key, responseContent)
	}

	return andflow.RESULT_SUCCESS, nil
}
