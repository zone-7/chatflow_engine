package flow

import (
	"errors"
	"strings"

	"github.com/zone-7/andflow_go/andflow"
)

func init() {
	andflow.RegistActionRunner("db_sql", &Db_sql_Runner{})
}

type Db_sql_Runner struct {
	BaseRunner
}

func (r *Db_sql_Runner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {
	var err error

	actionId := param.ActionId

	action := s.GetFlow().GetAction(param.ActionId)

	prop, err := r.getActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	param_key := prop["param_key"]
	drivername := prop["drivername"]
	datasource := prop["datasource"]
	sql := prop["sql"]

	if len(param_key) == 0 {
		param_key = actionId
	}

	alias := drivername + "_" + actionId

	if len(datasource) == 0 {
		return 0, errors.New("数据库地址不能为空")
	}

	if len(sql) == 0 {
		return 0, errors.New("SQL不能为空")
	}

	s.AddLog_action_info(action.Name, action.Title, "开始执行查询SQL")

	if strings.Index(strings.Trim(strings.ToLower(sql), " "), "select") == 0 {
		data, err := dbQuery(drivername, alias, datasource, sql)
		if err != nil {
			return andflow.RESULT_FAILURE, err
		}

		s.SetParam(param_key, data)
		return andflow.RESULT_SUCCESS, nil
	} else {
		data, err := dbExec(drivername, alias, datasource, sql)
		if err != nil {
			return andflow.RESULT_FAILURE, err
		}
		s.SetParam(param_key, data)
		return andflow.RESULT_SUCCESS, nil
	}

}
