package flow

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	"html/template"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/beego/beego/orm"
)

func isFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		log.Println(err)
		return false
	}
	return true
}
func dbRegist(driverName string, alias string, datasource string) error {

	var err error
	switch driverName {
	case "sqlite3", "sqlite":
		err = orm.RegisterDriver(driverName, orm.DRSqlite)
		if err != nil {
			return err
		}
		err = orm.RegisterDataBase(alias, driverName, datasource)
		if err != nil {
			return err
		}
	case "mysql":
		err = orm.RegisterDriver("mysql", orm.DRMySQL)
		if err != nil {
			return err
		}
		err = orm.RegisterDataBase(alias, driverName, datasource, 30)
		if err != nil {
			return err
		}

	}

	return err
}

// 执行sql
func dbExec(driverName string, alias string, datasource string, sql string, args ...interface{}) (int64, error) {

	db, err := orm.GetDB(alias)
	if err != nil || db == nil {
		dbRegist(driverName, alias, datasource)
		db, err = orm.GetDB(alias)
	}
	if err != nil || db == nil {
		return 0, err
	}

	res, err := db.Exec(sql, args...)

	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

// 执行sql查询
func dbQuery(driverName string, alias string, datasource string, query string, args ...interface{}) ([]map[string]interface{}, error) {

	db, err := orm.GetDB(alias)
	if err != nil || db == nil {
		dbRegist(driverName, alias, datasource)
		db, err = orm.GetDB(alias)
	}

	if err != nil {
		return nil, err
	}

	o, err := orm.NewOrmWithDB(driverName, alias, db)

	if err != nil {
		return nil, err
	}

	items := make([]orm.Params, 0)

	_, err = o.Raw(query, args...).Values(&items)

	if err != nil {
		return nil, err
	}

	list := make([]map[string]interface{}, 0)
	for _, item := range items {
		list = append(list, item)
	}

	return list, nil
}

// 获取词语列表，逗号或者换行分隔
func getWords(keywords string) []string {
	words := make([]string, 0)
	kws1 := strings.Split(keywords, "\n")
	for _, kw1 := range kws1 {
		kws2 := strings.Split(kw1, ",")
		for _, kw2 := range kws2 {
			kws3 := strings.Split(kw2, "，")
			for _, kw3 := range kws3 {
				kws4 := strings.Split(kw3, " ")
				for _, kw4 := range kws4 {
					if len(kw4) > 0 {
						words = append(words, kw4)
					}
				}
			}
		}
	}

	return words
}

func unescapeHTML(s any) template.HTML {

	var str string
	if v, ok := s.(string); ok {
		str = v
	} else {
		switch s.(type) {
		case string:
		case float32:
		case float64:
		case int:
		case int16:
		case int32:
		case int64:
		case bool:
			str = fmt.Sprintf("%v", s)
			break
		default:
			data, err := json.Marshal(s)
			if err == nil {
				str = string(data)
			} else {
				str = fmt.Sprintf("%v", s)
			}
		}
	}

	return template.HTML(str)
}

// 模版替换
func replaceTemplate(temp string, name string, params map[string]any) (string, error) {
	if strings.Index(temp, "{{") < 0 || strings.Index(temp, "}}") < 0 {
		return temp, nil
	}
	// 正则表达式,替换前面没有$.或. 的不合规
	need_replaces := make(map[string]string)
	reg := regexp.MustCompile(`\{\{[^{}]*\}\}`)

	items := reg.FindAll([]byte(temp), -1)
	for _, item := range items {

		str := string(item)

		key := strings.Replace(str, "{{", "", -1)
		key = strings.Replace(key, "}}", "", -1)
		key = strings.Trim(key, " ")

		newkey := key
		if strings.Index(key, "$.") != 0 && strings.Index(key, ".") != 0 {
			newkey = "." + key
		}

		need_replaces[str] = "{{ unescapeHTML " + newkey + "}}"

	}
	for o, n := range need_replaces {
		temp = strings.ReplaceAll(temp, o, n)
	}

	//解析模板
	t, err := template.New(name).Funcs(template.FuncMap{"unescapeHTML": unescapeHTML}).Parse(temp)

	if err != nil {
		return temp, err
	}

	b := bytes.NewBuffer(make([]byte, 0))
	bw := bufio.NewWriter(b)

	err = t.Execute(bw, params)
	bw.Flush()

	return string(b.Bytes()), err
}
