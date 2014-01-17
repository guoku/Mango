package crawler

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/qiniu/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
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
	if !strings.Contains(shoplink, ".m.") {
		shoplink = strings.Replace(shoplink, ".taobao.com", ".m.taobao.com", 1)
		shoplink = strings.Replace(shoplink, ".tmall.com", ".m.tmall.com", 1)
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
		if resp == nil {
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

	shopIds := re.FindStringSubmatch(string(html))
	if len(shopIds) < 2 {
		err = errors.New("no shopid")
		log.Error(err, shoplink)
	}
	shopId := shopIds[1]
	if shopId == "" {
		err = errors.New("no shopid")
	}
	return shopId, nil

}

//通过wap版某件商品的页面，获取其所属店铺的wap超链接
//parse的时候，会提取出这个链接的
func GetShopLink(html string) (string, error) {
	//	re := regexp.MustCompile("\\<a href=\\\"(.+com).+进入店铺")
	if html == "" {
		e := errors.New("no content")
		return "", e
	}
	reader := strings.NewReader(html)
	doc, e := goquery.NewDocumentFromReader(reader)
	if e != nil {
		log.Info(e.Error())
		return "", e
	}
	shoptag := doc.Find("html body div.bd div.left-margin-5 p strong a")
	shoplink, exists := shoptag.Attr("href")
	if exists {
		return shoplink, nil
	} else {
		e := errors.New("parse no shoplink")
		return "", e
	}
}

//通过wap版某件商品的页面，获取其所属店铺的店主nick
func GetShopNick(fontpage string) (nick string, err error) {
	reader := strings.NewReader(fontpage)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Error(err)
		return
	}
	nicktag, _ := doc.Find("body div.bd div.box div.detail p a img").Attr("src")
	log.Info(nicktag)
	p, _ := url.Parse(nicktag)
	q := p.Query()
	nick = q.Get("nick")
	return

}
