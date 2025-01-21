package flow

import (
	"errors"

	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/utils"
)

func init() {
	andflow.RegistActionRunner("file_read", &File_readRunner{})
}

type File_readRunner struct {
	BaseRunner
}

func (r *File_readRunner) Properties() []andflow.Prop {
	return []andflow.Prop{}
}
func (r *File_readRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {
	var err error

	action := s.GetFlow().GetAction(param.ActionId)

	prop, err := r.GetActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	filepath := prop["file"]
	param_key := prop["param_key"]

	if len(filepath) == 0 {
		s.AddLog_action_error(action.Name, action.Title, "文件路径不能为空")

		return andflow.RESULT_FAILURE, errors.New("文件路径不能为空")
	}

	filepath, err = replaceTemplate(filepath, "file_read_filepath", s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	if len(param_key) == 0 {
		param_key = action.Id
	}

	data, err := utils.ReadFile(filepath)

	if err != nil {
		s.AddLog_action_error(action.Name, action.Title, "读取文件失败"+err.Error())

		return andflow.RESULT_FAILURE, err
	}

	s.SetParam(param_key, string(data))

	return andflow.RESULT_SUCCESS, nil
}
