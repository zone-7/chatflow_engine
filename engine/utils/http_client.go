package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func SendHttpGet(address string, content string) (string, error) {
	if len(content) > 0 {
		if strings.Index(address, "?") >= 0 {
			address = address + "&" + content
		} else {
			address = address + "?" + content
		}
	}

	address = url.QueryEscape(address)

	resp, err := http.Get(address)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
func SendHttpGetForm(address string, values map[string]interface{}) (string, error) {
	if len(values) > 0 {
		query := ""
		for k, v := range values {
			query += fmt.Sprintf("%s=%v&", k, v)
		}
		if strings.Index(address, "?") >= 0 {
			address = address + "&" + query
		} else {
			address = address + "?" + query
		}

	}

	address = url.QueryEscape(address)

	resp, err := http.Get(address)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func SendHttpPost(address string, content string) (string, error) {

	resp, err := http.Post(address, "application/json", strings.NewReader(content))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func SendHttpPostForm(address string, values map[string]interface{}) (string, error) {
	data := url.Values{}

	for k, v := range values {
		data[k] = []string{fmt.Sprintf("%v", v)}
	}

	resp, err := http.PostForm(address, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
