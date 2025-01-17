package flow

// 网络请求
import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/zone-7/andflow_go/andflow"
)

func init() {
	andflow.RegistActionRunner("net_request", &Net_requestRunner{})
}

type Net_requestRunner struct {
	BaseRunner
}

func (r *Net_requestRunner) Execute(s *andflow.Session, param *andflow.ActionParam, state *andflow.ActionStateModel) (andflow.Result, error) {

	var err error

	actionId := param.ActionId
	action := s.GetFlow().GetAction(param.ActionId)

	prop, err := r.getActionParams(action, s.GetParamMap())
	if err != nil {
		return andflow.RESULT_FAILURE, err
	}

	var result string

	var headers = make(map[string]string)
	var cookies = make(map[string]string)

	dist_url := prop["url"]
	proxy := prop["proxy"]
	useragent := prop["user_agent"]
	param_key := prop["param_key"]
	method := prop["method"]
	body := prop["body_template"]
	headersJson := prop["headers"]
	cookiesJson := prop["cookies"]

	if len(dist_url) == 0 {
		s.AddLog_action_error(action.Name, action.Title, "地址不能为空"+err.Error())
		return andflow.RESULT_FAILURE, err
	}

	domain := getDomain(dist_url)

	if len(method) == 0 {
		method = "post"
	}

	if len(headersJson) > 0 {

		json.Unmarshal([]byte(headersJson), &headers)
	}

	if len(cookiesJson) > 0 {

		json.Unmarshal([]byte(cookiesJson), &cookies)
	}

	//网络调用
	conn := colly.NewCollector(
		colly.Async(false),
	)
	log.Printf("request: %v \n", dist_url)
	//设置代理
	if len(proxy) > 0 {
		conn.SetProxy(proxy)
	}
	if len(useragent) > 0 {
		conn.UserAgent = useragent
	}

	if len(cookies) > 0 {
		cs := make([]*http.Cookie, 0)
		for k, v := range cookies {
			c := http.Cookie{Name: k, Value: v, Expires: time.Now().AddDate(0, 0, 1), Domain: domain}
			cs = append(cs, &c)
		}
		conn.SetCookies(dist_url, cs)
	}

	conn.OnRequest(func(r *colly.Request) {

		if len(headers) > 0 {
			for k, v := range headers {
				r.Headers.Set(k, v)
			}
		}

	})

	conn.OnError(func(res *colly.Response, e error) {
		s.AddLog_action_error(action.Name, action.Title, "网络请求异常"+e.Error())

		err = e
	})

	conn.OnResponse(func(res *colly.Response) {

		resetValue(&result, string(res.Body))

	})

	conn.OnHTML("", func(el *colly.HTMLElement) {

	})

	//onhtml 之后执行
	conn.OnScraped(func(_ *colly.Response) {

	})

	//发送
	if len(dist_url) > 0 {
		if method == "post" {
			//POST
			if len(body) > 0 {
				err = conn.PostRaw(dist_url, []byte(body))
			} else {
				err = conn.Post(dist_url, nil)
			}

		} else {
			//GET
			if len(body) > 0 {
				if strings.Index(dist_url, "?") >= 0 {
					dist_url += "&" + body
				} else {
					dist_url += "?" + body
				}
			}

			err = conn.Visit(dist_url)

		}
	}

	if err != nil {
		s.AddLog_action_error(action.Name, action.Title, "网络请求异常"+err.Error())
		return andflow.RESULT_FAILURE, err
	}

	if len(param_key) == 0 {
		param_key = actionId
	}
	s.SetParam(param_key, result)

	return andflow.RESULT_SUCCESS, nil
}

// 字符串覆盖
func resetValue(s *string, newValue string) {
	sByte := []byte(*s)
	for i := 0; i < len(sByte); i++ {
		sByte[i] = ' '
	}
	*s = newValue
}

// 获取域名
func getDomain(url string) string {
	if len(url) == 0 {
		return url
	}
	var domain string
	//获取Domain
	start := strings.Index(url, "//")
	if start >= 0 {
		domain = url[start+2:]
	} else {
		domain = url
	}

	end := strings.Index(domain, "/")
	if end >= 0 {
		domain = domain[0:end]
	}
	end = strings.Index(domain, ":")
	if end >= 0 {
		domain = domain[0:end]
	}

	return domain
}
