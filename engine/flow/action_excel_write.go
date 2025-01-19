package flow

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/zone-7/andflow_go/andflow"
	"github.com/zone-7/chatflow_engine/engine/utils"
)

func init() {
	andflow.RegistActionRunner("excel_write", &Excel_writeRunner{})
}

type Excel_writeRunner struct {
	BaseRunner
}

func (r *Excel_writeRunner) Properties() []andflow.Prop {
	return []andflow.Prop{}
}
func (r *Excel_writeRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {

	var err error
	actionId := param.ActionId

	action := s.GetFlow().GetAction(param.ActionId)

	prop, err := r.GetActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	file := prop["file"]
	isappend := prop["isappend"] //是否追加
	fromRow := prop["from_row"]  //从第几行开始
	fromCel := prop["from_col"]  //从第几列开始
	param_key := prop["param_key"]

	if len(file) == 0 {
		s.AddLog_action_error(action.Name, action.Title, "Excel文件路径不能为空")

		return andflow.RESULT_FAILURE, errors.New("Excel文件路径不能为空")
	}

	fr := 1
	if len(fromRow) > 0 {
		fr, err = strconv.Atoi(fromRow)
	}

	//获取行号索引
	if len(isappend) > 0 && (isappend == "1" || isappend == "true" || isappend == "是") {
		ts := s.GetParam("current_excel_index_" + s.GetRuntime().Id + "_" + actionId)

		if ts != nil {
			ti, err := strconv.Atoi(fmt.Sprintf("%v", ts))
			if err != nil {
				fr = ti
			}
		}
	}

	fc := 1
	if len(fromCel) > 0 {
		fc, err = strconv.Atoi(fromCel)
	}

	if len(param_key) == 0 {
		param_key = actionId
	}

	dataObj := s.GetParam(param_key)

	datas := make([][]interface{}, 0)
	switch dataObj.(type) {
	case [][]string: //json格式
		datas = dataObj.([][]interface{})
		break
	case []string: //json格式
		data := dataObj.([]interface{})
		datas = append(datas, data)
		break

	case []map[string]interface{}:
		for _, item := range dataObj.([]map[string]interface{}) {
			data := make([]interface{}, 0)
			for _, v := range item {
				data = append(data, v)
			}
			datas = append(datas, data)
		}
		break
	case [][]interface{}:
		datas = dataObj.([][]interface{})
		break
	case []interface{}:
		rows := dataObj.([]interface{})
		for _, row := range rows {
			switch row.(type) {
			case string:
				data := make([]interface{}, 0)
				data = append(data, row.(string))
				datas = append(datas, data)
				break
			case []string:
				datas = append(datas, row.([]interface{}))
				break
			case []interface{}:
				datas = append(datas, row.([]interface{}))
				break
			}
		}

		break
	case string:
		data := make([]interface{}, 0)
		data = append(data, dataObj.(interface{}))
		datas = append(datas, data)
		break
	}

	if len(datas) == 0 {
		return andflow.RESULT_SUCCESS, nil
	}

	fr, err = utils.ExcelExport(file, datas, fr, fc)
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	s.SetParam("current_excel_index_"+s.GetRuntime().Id+"_"+actionId, fmt.Sprintf("%d", fr))

	return andflow.RESULT_SUCCESS, err
}
