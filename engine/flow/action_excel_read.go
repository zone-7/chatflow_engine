package flow

import (
	"errors"
	"strconv"

	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/utils"
)

func init() {
	andflow.RegistActionRunner("excel_read", &Excel_readRunner{})
}

type Excel_readRunner struct {
	BaseRunner
}

func (r *Excel_readRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {

	var err error

	actionId := param.ActionId
	action := s.GetFlow().GetAction(param.ActionId)

	prop, err := r.getActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	file := prop["file"]
	fromRow := prop["from_row"]
	toRow := prop["to_row"]
	fromCel := prop["from_col"]
	toCel := prop["to_col"]
	param_key := prop["param_key"]

	if len(file) == 0 {
		s.AddLog_action_error(action.Name, action.Title, "Excel文件路径不能为空")
		return andflow.RESULT_FAILURE, errors.New("Excel文件路径不能为空")
	}

	fr := 1
	if len(fromRow) > 0 {
		fr, err = strconv.Atoi(fromRow)
	}

	tr := 0
	if len(toRow) > 0 {
		tr, err = strconv.Atoi(toRow)
	}

	fc := 1
	if len(fromCel) > 0 {
		fr, err = strconv.Atoi(fromCel)
	}

	tc := 0
	if len(toCel) > 0 {
		tc, err = strconv.Atoi(toCel)
	}

	if len(param_key) == 0 {
		param_key = actionId
	}

	datas, err := utils.ExcelImport(file, fr, tr, fc, tc)
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	s.SetParam(param_key, datas)

	return andflow.RESULT_SUCCESS, err
}
