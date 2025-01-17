package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/meta"
)

type ChatSessionInfoManager struct {
	Opt meta.Option
}

func NewChatSessionInfoManager(opt meta.Option) ChatSessionInfoManager {
	return ChatSessionInfoManager{Opt: opt}
}

func (s *ChatSessionInfoManager) GetSessionDir() string {
	return GetSessionPath(s.Opt)

}
func (s *ChatSessionInfoManager) GetParamDir() string {
	return GetParamPath(s.Opt)

}

func (s *ChatSessionInfoManager) LoadUserCount() int {
	dir := s.GetSessionDir()

	dirEntrys, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	return len(dirEntrys)
}

func (s *ChatSessionInfoManager) LoadSessionCount() int {
	dir := s.GetSessionDir()

	userDirEntrys, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}

	count := 0

	// 最终用户列表
	for _, userDirEntry := range userDirEntrys {
		userDir := path.Join(dir, userDirEntry.Name())
		userFlowDirEntrys, err := os.ReadDir(userDir)
		if err != nil {
			continue
		}

		for _, userFlowDirEntry := range userFlowDirEntrys {

			userFlowDir := path.Join(userDir, userFlowDirEntry.Name())

			userFlowSessionDirEntrys, err := os.ReadDir(userFlowDir)
			if err != nil {
				continue
			}
			count = count + len(userFlowSessionDirEntrys)
		}

	}

	return count

}

// 加载用户参数
func (s *ChatSessionInfoManager) LoadUserChatFlowParams(user_id string, flow_code string) ([]*meta.ChatFlowParam, error) {
	params := make([]*meta.ChatFlowParam, 0)
	dir := path.Join(s.GetParamDir(), user_id, flow_code)

	file := path.Join(dir, "params.json")

	data, err := os.ReadFile(file)

	if err != nil {
		return params, err
	}

	err = json.Unmarshal(data, &params)
	if err != nil {
		return nil, err
	}

	return params, nil
}

// 保存用户参数
func (s *ChatSessionInfoManager) StoreUserChatFlowParams(userparam meta.UserChatFlowParams) error {
	if userparam.Params == nil {
		return errors.New("params empty")
	}
	if len(userparam.UserId) == 0 {
		return errors.New("UserId empty")
	}
	if len(userparam.FlowCode) == 0 {
		return errors.New("FlowCode empty")
	}

	data, err := json.MarshalIndent(userparam.Params, "", "\t")
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	dir := path.Join(s.GetParamDir(), userparam.UserId, userparam.FlowCode)

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	file := path.Join(dir, "params.json")

	err = os.WriteFile(file, data, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	return nil
}

func (s *ChatSessionInfoManager) RemoveSession(user_id string, flow_code string, session_id string) error {
	dir := path.Join(s.GetSessionDir(), user_id, flow_code, session_id)

	err := os.RemoveAll(dir)
	return err
}

func (s *ChatSessionInfoManager) RemoveAllSessions(user_id string, flow_code string) error {
	dir := path.Join(s.GetSessionDir(), user_id, flow_code)

	err := os.RemoveAll(dir)
	return err
}

// 加载会话列表
func (s *ChatSessionInfoManager) LoadSessionInfos(user_id string, flow_code string) ([]*meta.ChatSessionInfo, error) {
	infos := make([]*meta.ChatSessionInfo, 0)
	dir := path.Join(s.GetSessionDir(), user_id, flow_code)
	fs, err := os.ReadDir(dir)

	if err != nil {
		return infos, err
	}

	for _, f := range fs {
		file := path.Join(dir, f.Name(), "info.json")

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var info meta.ChatSessionInfo
		err = json.Unmarshal(data, &info)

		if err == nil {
			infos = append(infos, &info)
		}
	}

	sort.SliceStable(infos, func(i, j int) bool {
		createTime1 := infos[i].CreateTime
		createTime2 := infos[j].CreateTime
		return createTime1 > createTime2

	})

	return infos, nil
}

// 加载会话
func (s *ChatSessionInfoManager) LoadSessionInfo(user_id string, flow_code string, session_id string) (*meta.ChatSessionInfo, error) {

	dir := path.Join(s.GetSessionDir(), user_id, flow_code, session_id)
	file := path.Join(dir, "info.json")
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var info meta.ChatSessionInfo
	err = json.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// 存储会话信息
func (s *ChatSessionInfoManager) StoreSessionInfo(info *meta.ChatSessionInfo) error {
	if info == nil {
		return nil
	}
	user_id := info.UserId
	flow_code := info.FlowCode
	session_id := info.Id

	data, err := json.MarshalIndent(info, "", "\t")
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	dir := path.Join(s.GetSessionDir(), user_id, flow_code, session_id)

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	file := path.Join(dir, "info.json")

	err = os.WriteFile(file, data, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	return nil
}

func (s *ChatSessionInfoManager) LoadSessionMessages(user_id string, flow_code string, session_id string, start int, size int) ([]*meta.ChatFlowMessage, error) {

	msgs := make([]*meta.ChatFlowMessage, 0)

	dir := path.Join(s.GetSessionDir(), user_id, flow_code, session_id)
	file := path.Join(dir, "messages.json")
	data, err := os.ReadFile(file)
	if err != nil {
		return msgs, err
	}

	his := make([]*meta.ChatFlowMessage, 0)
	err = json.Unmarshal(data, &his)
	if err != nil {
		return msgs, err
	}

	//排序

	sort.SliceStable(his, func(i, j int) bool {
		createTime1 := his[i].SendTime
		createTime2 := his[j].SendTime
		return createTime1 > createTime2

	})

	//分页
	if len(his) <= start {
		return msgs, nil
	}
	if size == 0 || start+size > len(his) {
		size = len(his) - start
	}

	msgs = his[start : start+size]

	return msgs, nil
}

// 存储历史对话记录
func (s *ChatSessionInfoManager) StoreSessionMessages(info *meta.ChatSessionInfo, msgs []*meta.ChatFlowMessage) error {
	if info == nil || msgs == nil {
		return nil
	}
	user_id := info.UserId
	flow_code := info.FlowCode
	session_id := info.Id

	data, err := json.MarshalIndent(msgs, "", "\t")
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	dir := path.Join(s.GetSessionDir(), user_id, flow_code, session_id)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	file := path.Join(dir, "messages.json")

	err = os.WriteFile(file, data, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	return nil
}

func (s *ChatSessionInfoManager) LoadSessionRuntime(user_id string, flow_code string, session_id string) (*andflow.RuntimeModel, error) {
	dir := path.Join(s.GetSessionDir(), user_id, flow_code, session_id)

	os.MkdirAll(dir, os.ModePerm)

	file := path.Join(dir, "runtime.json")
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var rt andflow.RuntimeModel

	err = json.Unmarshal(data, &rt)
	if err != nil {
		return nil, err
	}

	return &rt, nil
}

// 存储运行记录
func (s *ChatSessionInfoManager) StoreSessionRuntime(info *meta.ChatSessionInfo, runtime *andflow.RuntimeModel) error {
	if info == nil || runtime == nil {
		return nil
	}
	user_id := info.UserId
	flow_code := info.FlowCode
	session_id := info.Id
	data, err := json.MarshalIndent(runtime, "", "\t")
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	dir := path.Join(s.GetSessionDir(), user_id, flow_code, session_id)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}

	file := path.Join(dir, "runtime.json")

	err = os.WriteFile(file, data, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}
	return nil
}
