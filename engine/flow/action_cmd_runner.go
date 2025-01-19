package flow

import (
	"context"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/zone-7/andflow_go/andflow"
)

func init() {
	andflow.RegistActionRunner("cmd", &CmdRunner{})
}

type CmdRunner struct {
	BaseRunner
}

func (r *CmdRunner) Properties() []andflow.Prop {
	return []andflow.Prop{}
}
func (r *CmdRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {

	var err error
	var res string

	action := s.GetFlow().GetAction(param.ActionId)

	actionId := param.ActionId

	prop, err := r.GetActionParams(action, s.GetParamMap())

	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	command := prop["cmd_content"]
	timeoutStr := prop["cmd_timeout"]
	param_key := prop["param_key"]

	if len(param_key) == 0 {
		param_key = actionId
	}

	timeout := 10000
	if len(timeoutStr) > 0 {
		timeoutInt, err := strconv.Atoi(timeoutStr)

		if err == nil && timeoutInt > 100 && timeoutInt < 60000 {
			timeout = timeoutInt
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	command, err = replaceTemplate(command, actionId, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	lines := strings.Split(command, "\n")

	for _, line := range lines {

		if len(strings.Trim(line, " ")) == 0 {
			continue
		}

		commandArr := strings.Split(line, " ")

		name := commandArr[0]
		attr := commandArr[1:]

		cmd := exec.CommandContext(ctx, name, attr...)

		var stdout io.ReadCloser

		if stdout, err = cmd.StdoutPipe(); err != nil { //获取输出对象，可以从该对象中读取输出结果
			return andflow.RESULT_FAILURE, err
		}

		defer stdout.Close() // 保证关闭输出流

		if err = cmd.Start(); err != nil { // 运行命令
			return andflow.RESULT_FAILURE, err
		}

		var opBytes []byte
		if opBytes, err = ioutil.ReadAll(stdout); err != nil { // 读取输出结果
			return andflow.RESULT_FAILURE, err
		}

		res = string(opBytes)
	}

	s.SetParam(param_key, res)
	return andflow.RESULT_SUCCESS, nil
}
