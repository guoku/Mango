package utils

import (
	"github.com/qiniu/log"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetShopLink(t *testing.T) {
	resp, _ := http.Get("http://a.m.taobao.com/i35551329262.htm")
	data, _ := ioutil.ReadAll(resp.Body)
	link := GetShopLink(string(data))
	log.Info(link)
}
