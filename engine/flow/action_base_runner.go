package flow

import (
	"strings"

	"github.com/zone-7/andflow_go/andflow"
)

type BaseRunner struct {
}

func (r *BaseRunner) GetChatSession(s *andflow.Session) *ChatSession {
	runtimeId := s.GetRuntime().Id

	session := GetChatSession(runtimeId)

	return session
}

func (r *BaseRunner) GetActionParam(action *andflow.ActionModel, key string, ps map[string]interface{}) string {
	value := action.Params[key]
	if len(value) > 0 && ps != nil && len(ps) > 0 {

		vv, err := replaceTemplate(value, "temp_"+action.Id, ps)
		if err != nil {
			return value
		}
		value = vv

	}
	return value
}

func (r *BaseRunner) GetActionParams(action *andflow.ActionModel, ps map[string]interface{}) (map[string]string, error) {

	params := make(map[string]string)

	for k, v := range action.Params {
		var value string
		value = v
		if len(v) > 0 && ps != nil && len(ps) > 0 {

			vv, err := replaceTemplate(v, "temp_"+action.Id, ps)
			if err != nil {
				return nil, err
			}
			value = vv

		}

		params[k] = value
	}
	return params, nil
}

func (r *BaseRunner) GetNextActionsByName(s *andflow.Session, action *andflow.ActionModel, name string) []*andflow.ActionModel {
	//获取后续节点
	var nas []*andflow.ActionModel
	if strings.Trim(name, " ") == "" {
		return nas
	}
	nextLinks := s.GetFlow().GetLinkBySourceId(action.Id)

	if nextLinks != nil && len(nextLinks) > 0 {

		for _, link := range nextLinks {
			target_action := s.GetFlow().GetAction(link.TargetId)

			if (len(strings.Trim(target_action.Name, " ")) > 0 && strings.Index(name, target_action.Name) >= 0) || (len(strings.Trim(link.Name, " ")) > 0 && strings.Index(name, link.Name) >= 0) {
				nas = append(nas, target_action)
			}
		}
	}
	return nas
}

// 执行下一步的判断
func (r *BaseRunner) GetNextActionsByKeyword(s *andflow.Session, action *andflow.ActionModel, messageContent string) []*andflow.ActionModel {
	//获取后续节点
	var nas []*andflow.ActionModel

	nextLinks := s.GetFlow().GetLinkBySourceId(action.Id)

	if nextLinks != nil && len(nextLinks) > 0 {

		for _, link := range nextLinks {
			target_action := s.GetFlow().GetAction(link.TargetId)
			link_words := getWords(link.Keywords)
			action_words := getWords(target_action.Keywords)

			words := append(link_words, action_words...)

			// 如果没有配置路由关键字，就表示接受
			if len(words) > 0 {
				// by key word match
				for _, kw := range words {
					if strings.Index(messageContent, kw) >= 0 {
						nas = append(nas, target_action)
						break
					}
				}

			}

		}

	}

	return nas
}

// 执行下一步keyword空的路径
func (r *BaseRunner) GetNextActionsByEmptyKeyword(s *andflow.Session, action *andflow.ActionModel) []*andflow.ActionModel {
	//获取后续节点
	var nas []*andflow.ActionModel

	nextLinks := s.GetFlow().GetLinkBySourceId(action.Id)

	if nextLinks != nil && len(nextLinks) > 0 {

		for _, link := range nextLinks {
			target_action := s.GetFlow().GetAction(link.TargetId)
			link_words := getWords(link.Keywords)
			action_words := getWords(target_action.Keywords)

			words := append(link_words, action_words...)

			// 如果没有配置路由关键字，就表示接受
			if len(words) == 0 {
				nas = append(nas, target_action)
			}

		}

	}

	return nas
}
