package flow

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path"

	"github.com/zone-7/andflow_go/andflow"
)

func init() {
	andflow.RegistActionRunner("file_write", &File_writeRunner{})
}

type File_writeRunner struct {
	BaseRunner
}

func (r *File_writeRunner) Properties() []andflow.Prop {
	return []andflow.Prop{}
}
func (r *File_writeRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {
	var err error

	action := s.GetFlow().GetAction(param.ActionId)

	prop, err := r.GetActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	filepath := prop["file"]
	content := prop["content"]

	if len(filepath) == 0 {
		s.AddLog_action_error(action.Name, action.Title, "文件路径不能为空")
		return andflow.RESULT_FAILURE, errors.New("文件路径不能为空")
	}

	filepath, err = replaceTemplate(filepath, "file_read_filepath", s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	//分隔字符类型
	split := action.GetParam("split")

	if !isFileExist(filepath) {
		dir := path.Dir(filepath)
		if !isFileExist(dir) {
			err = os.MkdirAll(dir, fs.ModePerm)
			if err != nil {
				return andflow.RESULT_FAILURE, err
			}
		}
	}
	// 追加模式
	isappend := action.GetParam("isappend")
	flat := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if isappend == "true" || isappend == "1" || isappend == "是" {
		flat = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}

	file, err := os.OpenFile(filepath, flat, 0666)
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}
	//及时关闭file句柄
	defer file.Close()

	//文件大小
	fi, err := os.Stat(filepath)
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}
	filesize := fi.Size()

	//写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)

	if filesize > 0 {
		if split == "1" || split == "\n" || split == "\\n" {
			write.WriteString("\n")
		}
		if split == "2" || split == "," {
			write.WriteString(",")
		}
		if split == "3" || split == "\t" {
			write.WriteString("\t")
		}
	}

	write.WriteString(content)

	//Flush将缓存的文件真正写入到文件中
	write.Flush()

	return andflow.RESULT_SUCCESS, nil
}
