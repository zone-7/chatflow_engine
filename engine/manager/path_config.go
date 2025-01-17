package manager

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/beego/beego"
	"github.com/zone-7/chatflow_engine/engine/meta"
)

// 系统运行路径
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		beego.Debug(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

// func GetWorkspacePath() string {
// 	p := beego.AppConfig.String("admin::workspace_path")
// 	return p
// }

// func GetSystemDatabase() string {
// 	p := beego.AppConfig.String("admin::system_database")
// 	return p
// }

// 流程设计路径
func GetFlowDevelopPath(opt meta.Option) string {
	p := path.Join(opt.WorkspacePath, "flow/develop")
	return p
}

// 流程模版路径
func GetFlowTemplatePath(opt meta.Option) string {
	p := path.Join(opt.WorkspacePath, "flow/template")

	return p
}

// 流程发布路径
func GetFlowProductPath(opt meta.Option) string {
	p := path.Join(opt.WorkspacePath, "flow/product")
	return p
}

// 知识库路径
func GetKnowledgePath(opt meta.Option) string {
	p := path.Join(opt.WorkspacePath, "knowledge")
	return p
}

// 用户会话状态
func GetSessionPath(opt meta.Option) string {

	p := path.Join(opt.WorkspacePath, "session")
	return p
}

// 用户参数状态
func GetParamPath(opt meta.Option) string {
	p := path.Join(opt.WorkspacePath, "param")
	return p
}
