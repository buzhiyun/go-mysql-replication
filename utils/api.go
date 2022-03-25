package utils

import (
	"bytes"
	jsoniter "github.com/json-iterator/go"
	"github.com/kataras/golog"
	"io/ioutil"
	"net/http"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type ApiJson struct {
	Status int         `json:"status"`
	Msg    interface{} `json:"msg"`
	Data   interface{} `json:"data"`
}

var client = &http.Client{}

func HttpPostJson(url string, body interface{}) (responseBody []byte, err error) {
	requestJson, err := json.Marshal(body)
	if err != nil {
		golog.Errorf("http 发送初始化失败，无法json参数, %v", body)
		return
	}
	golog.Debugf("发送接口: %s ，body: %s", requestJson)

	req, err := http.NewRequest("POST", url, bytes.NewReader(requestJson))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Cookie", "name=anny")

	client.Timeout = 5 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	responseBody, err = ioutil.ReadAll(resp.Body)

	return
}
