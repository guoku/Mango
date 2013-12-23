package utils

import (
	"errors"
	"github.com/qiniu/log"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

//给出店铺的首页，获取其淘宝店的id
func GetShopid(shoplink string) (string, error) {
	if shoplink == "" {
		e := errors.New("link is null")
		return "", e
	}
	re := regexp.MustCompile("http://shop(\\d+)\\.")
	sids := re.FindStringSubmatch(shoplink)
	if len(sids) > 1 {
		sid := sids[1]
		log.Info(sid)
		return sid, nil
	}
	transport := &http.Transport{ResponseHeaderTimeout: time.Duration(30) * time.Second}
	var redirectFunc = func(req *http.Request, via []*http.Request) error {
		redirectUrl := req.URL.String()
		log.Info("redirectUrl ", redirectUrl)
		return nil
	}
	client := &http.Client{Transport: transport, CheckRedirect: redirectFunc}
	req, err := http.NewRequest("GET", shoplink, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:24.0) Gecko/20100101 Firefox.24.0")
	if err != nil {
		log.Error(err)
		return "", err
	}
	resp, err := client.Do(req)
	defer func() {
		if resp.Body == nil {
			return
		} else {
			resp.Body.Close()
			return
		}
	}()
	if err != nil {
		log.Info(err.Error())
		return "", err
	}
	if resp.StatusCode != 200 {
		log.Error(resp.Status)
		err = errors.New(resp.Status)
		return "", err
	}
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return "", err
	}
	re = regexp.MustCompile("shopId=(?P<id>\\d+)")
	shopId := re.FindStringSubmatch(string(html))[1]
	if shopId == "" {
		err = errors.New("no shopid")
	}
	return shopId, nil

}
