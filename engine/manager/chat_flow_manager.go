package manager

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/meta"
)

var model_file_name_flow = "model.json"

type ChatFlowManager struct {
	// FlowSpace string
	Opt meta.Option
}

func NewChatFlowManager(opt meta.Option) ChatFlowManager {
	return ChatFlowManager{Opt: opt}
}

func (c *ChatFlowManager) GetDir(space string, code string) string {
	var dir string
	if space == meta.FLOW_SPACE_DEVELOP {

		dir = path.Join(GetFlowDevelopPath(c.Opt), code)
	} else if space == meta.FLOW_SPACE_TEMPLATE {
		dir = path.Join(GetFlowTemplatePath(c.Opt), code)
	} else if space == meta.FLOW_SPACE_PRODUCT {
		dir = path.Join(GetFlowProductPath(c.Opt), code)
	} else {
		dir = path.Join(GetFlowDevelopPath(c.Opt), code)
	}

	os.MkdirAll(dir, os.ModePerm)

	return dir
}

func (c *ChatFlowManager) CreateChatFlow(name string) (*meta.ChatFlow, error) {
	chatflow := &meta.ChatFlow{}

	uid, _ := uuid.NewV4()
	code := strings.ReplaceAll(uid.String(), "-", "")
	chatflow.Code = code

	chatflow.Name = name

	chatflow.Active = "true"

	chatflow.FlowModel = andflow.CreateFlowModel(code, name)

	// chatflow.SysUserId = c.Opt.SysUserId

	return chatflow, nil
}

func (c *ChatFlowManager) ChatFlowCount(flow_space string) int {
	dir := c.GetDir(flow_space, "")

	fs, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, f := range fs {

		file := path.Join(dir, f.Name(), model_file_name_flow)

		_, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		count++
	}

	return count
}

func (c *ChatFlowManager) LoadChatFlowInfo(flow_space, code string) (*meta.ChatFlowInfo, error) {
	if len(code) == 0 {
		return nil, nil
	}

	dir := c.GetDir(flow_space, code)

	jsonFile := path.Join(dir, model_file_name_flow)

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}

	var flowInfo meta.ChatFlowInfo
	err = json.Unmarshal(data, &flowInfo)

	if err != nil {
		return nil, err
	}

	return &flowInfo, nil
}

// 加载一个流程
func (c *ChatFlowManager) LoadChatFlow(flow_space, code string) (*meta.ChatFlow, error) {
	if len(code) == 0 {
		return nil, nil
	}

	dir := c.GetDir(flow_space, code)

	jsonFile := path.Join(dir, model_file_name_flow)

	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, errors.New("对话流程不存在")
	}

	var flow meta.ChatFlow
	err = json.Unmarshal(data, &flow)

	if err != nil {
		return nil, err
	}

	return &flow, nil
}

// 查询加载流程信息
func (c *ChatFlowManager) QueryChatFlowInfo(flow_space string, flow_type string, active string, name string) ([]*meta.ChatFlowInfo, error) {
	name = strings.Trim(name, " ")
	name = strings.ToLower(name)

	dir := c.GetDir(flow_space, "")

	fs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	infos := make([]*meta.ChatFlowInfo, 0)

	for _, f := range fs {

		file := path.Join(dir, f.Name(), model_file_name_flow)

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var info meta.ChatFlowInfo
		err = json.Unmarshal(data, &info)

		if err != nil {
			continue
		}

		// if len(c.Opt.SysUserId) > 0 {
		// 	if info.SysUserId != c.Opt.SysUserId {
		// 		continue
		// 	}
		// }

		if len(flow_type) > 0 {
			if flow_type != info.FlowType {
				continue
			}
		}
		if len(active) > 0 {
			if active != info.Active {
				continue
			}
		}

		if len(name) > 0 {
			if !(strings.Contains(strings.ToLower(info.SubTitle), name) || strings.Contains(strings.ToLower(info.Name), name) || strings.Contains(strings.ToLower(info.Description), name)) {
				continue
			}
		}

		infos = append(infos, &info)

	}
	// 排序
	sort.SliceStable(infos, func(i, j int) bool {

		order1 := infos[i].OrderNum
		order2 := infos[j].OrderNum
		createTime1 := infos[i].CreateTime
		createTime2 := infos[j].CreateTime

		if order1 > 0 && order2 > 0 {
			if order1 == order2 {

				return createTime1 > createTime2
			} else {
				return order1 > order2
			}

		}

		return createTime1 > createTime2

	})
	return infos, nil
}

// 加载所有流程
func (c *ChatFlowManager) ListChatFlowInfo(flow_space string) ([]*meta.ChatFlowInfo, error) {
	return c.QueryChatFlowInfo(flow_space, "", "", "")

}

// 加载可用的流程
func (c *ChatFlowManager) ListChatFlowInfoActive(flow_space string) ([]*meta.ChatFlowInfo, error) {
	return c.QueryChatFlowInfo(flow_space, "", "true", "")
}

// 加载流程图片
func (c *ChatFlowManager) LoadChatFlowImage(flow_space string, code string) ([]byte, error) {
	dir := c.GetDir(flow_space, code)

	file := path.Join(dir, "snap.jpeg")

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err

	}

	return data, nil
}

func (c *ChatFlowManager) SetChatFlowPublished(flow_space string, flow_code string) error {
	oldchatflow, err := c.LoadChatFlow(flow_space, flow_code)
	if err != nil || oldchatflow == nil {
		return errors.New("流程不存在")
	}

	oldchatflow.Published = "true"

	data, err := json.MarshalIndent(oldchatflow, "", "\t")
	if err != nil {
		return err
	}

	dir := c.GetDir(flow_space, oldchatflow.Code)

	jsonFile := path.Join(dir, model_file_name_flow)

	err = os.WriteFile(jsonFile, data, fs.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (c *ChatFlowManager) SaveChatFlow(flow_space string, chatflow *meta.ChatFlow) error {

	//新增加
	if len(chatflow.Code) == 0 || len(chatflow.FlowModel.Code) == 0 {
		uid, _ := uuid.NewV4()
		code := strings.ReplaceAll(uid.String(), "-", "")
		chatflow.Code = code
		chatflow.FlowModel.Code = code
	}
	chatflow.FlowSpace = flow_space

	// chatflow.SysUserId = c.Opt.SysUserId

	chatflow.Published = "false"

	oldchatflow, _ := c.LoadChatFlow(flow_space, chatflow.Code)

	if oldchatflow != nil {
		chatflow.CreateTime = oldchatflow.CreateTime

		if flow_space == meta.FLOW_SPACE_DEVELOP {

			chatflow.Edition = oldchatflow.Edition + 1
		}

	}

	if chatflow.CreateTime == 0 {
		chatflow.CreateTime = time.Now().UnixNano() / 1e6
	}

	data, err := json.MarshalIndent(chatflow, "", "\t")
	if err != nil {
		return err
	}

	dir := c.GetDir(flow_space, chatflow.Code)

	jsonFile := path.Join(dir, model_file_name_flow)

	err = os.WriteFile(jsonFile, data, fs.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

// 删除流程
func (c *ChatFlowManager) RemoveChatFlow(flow_space string, code string) error {
	dir := c.GetDir(flow_space, code)
	return os.RemoveAll(dir)
}

// 保存流程图片
func (c *ChatFlowManager) SaveChatFlowImageBase64(flow_space string, code string, image string) {

	if len(image) > 0 {
		image = strings.Replace(image, "data:image/jpeg;base64,", "", 1)
		imageData, err := base64.StdEncoding.DecodeString(image)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		c.SaveChatFlowImage(flow_space, code, imageData)
	}
}

func (c *ChatFlowManager) SaveChatFlowImage(flow_space string, code string, imageData []byte) {

	dir := c.GetDir(flow_space, code)

	imageFile := path.Join(dir, "snap.jpeg")

	os.WriteFile(imageFile, imageData, fs.ModePerm)

}

func (c *ChatFlowManager) CopyChatFlow(sourceFlowSpace string, code string, targetFlowSpace string, targetCode string, targetName string) (*meta.ChatFlow, error) {

	if len(targetCode) == 0 {
		targetCode = code
	}

	target_chatflow, err := c.LoadChatFlow(targetFlowSpace, targetCode)
	//流程存在就先删除
	if err == nil && target_chatflow != nil {
		c.RemoveChatFlow(targetFlowSpace, targetCode)
	}

	//新的流程
	chatflow, err := c.LoadChatFlow(sourceFlowSpace, code)
	if err != nil {
		return nil, err
	}

	chatflow.Code = targetCode
	chatflow.FlowModel.Code = targetCode
	if len(targetName) > 0 {
		chatflow.Name = targetName
		chatflow.FlowModel.Name = targetName
	}

	err = c.SaveChatFlow(targetFlowSpace, chatflow)
	if err != nil {
		return nil, err
	}

	imageData, err := c.LoadChatFlowImage(sourceFlowSpace, code)
	if err == nil {
		c.SaveChatFlowImage(targetFlowSpace, targetCode, imageData)
	}

	return chatflow, nil
}

// 克隆
func (c *ChatFlowManager) CloneChatFlow(flow_space string, code string, newname string) (*meta.ChatFlow, error) {
	uid, _ := uuid.NewV4()
	newcode := strings.ReplaceAll(uid.String(), "-", "")

	chatflow, err := c.CopyChatFlow(flow_space, code, flow_space, newcode, newname)

	return chatflow, err
}

// 发布
func (c *ChatFlowManager) PublishToProduct(code string) (*meta.ChatFlow, error) {
	chatflow, err := c.CopyChatFlow(meta.FLOW_SPACE_DEVELOP, code, meta.FLOW_SPACE_PRODUCT, code, "")
	if err == nil {
		c.SetChatFlowPublished(meta.FLOW_SPACE_DEVELOP, code)
	}
	return chatflow, err
}

// 公开为模板
func (c *ChatFlowManager) PublishToTemplate(code string) (*meta.ChatFlow, error) {
	chatflow, err := c.CopyChatFlow(meta.FLOW_SPACE_DEVELOP, code, meta.FLOW_SPACE_TEMPLATE, code, "")
	return chatflow, err
}

// 从模板创建
func (c *ChatFlowManager) CreateFromTemplate(code string, newname string) (*meta.ChatFlow, error) {
	uid, _ := uuid.NewV4()
	newcode := strings.ReplaceAll(uid.String(), "-", "")

	chatflow, err := c.CopyChatFlow(meta.FLOW_SPACE_TEMPLATE, code, meta.FLOW_SPACE_DEVELOP, newcode, newname)
	return chatflow, err
}

func (c *ChatFlowManager) IsPublished(code string) bool {
	flowDevelop, err := c.LoadChatFlow(meta.FLOW_SPACE_DEVELOP, code)
	if err != nil {
		return false
	}

	if flowDevelop.Published == "true" {
		return true
	} else {
		return false
	}
}
