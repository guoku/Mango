package crawler

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/qiniu/log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func GetShopItems(shoplink string) ([]string, error) {
	shoplink = transUrl(shoplink)
	//transport := getTransport()
	client := &http.Client{}
	req, err := http.NewRequest("GET", shoplink, nil)
	ua := userAgentGen()
	req.Header.Set("User-Agent", ua)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		if resp == nil {
			log.Info("proxy 此时不可达")

		}
		time.Sleep(2 * time.Second)
		log.Error(err)
		return nil, err
	}
	if resp.StatusCode == 200 {
		doc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			return nil, err
		}

		alllink := doc.Find("html body div.bd div.box div.box div.detail table tbody tr td a")
		linktag := alllink.Eq(1)
		allitemsurl, exist := linktag.Attr("href")
		log.Info(allitemsurl)
		if !exist {
			err = errors.New("所有宝贝的链接提取不到")
			log.Error(err)
			return nil, err
		}
		allitemsurl, err = wrapUrl(allitemsurl)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		resp.Body.Close()
		items, err := getitems(allitemsurl)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		log.Info(items)
	}
	return nil, nil
}

func getitems(alllink string) (items []string, err error) {
	//该url是wap版店铺首页上的所有宝贝链接
	//transport := getTransport()
	client := &http.Client{}
	req, _ := http.NewRequest("GET", alllink, nil)
	ua := userAgentGen()
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile("\\d+")
	tds := doc.Find("html body div.bd div.box div.detail table tbody tr td.pic")
	tds.Each(func(i int, sq *goquery.Selection) {
		linktag := sq.Find("a")
		link, exist := linktag.Attr("href")
		if !exist {
			log.Error("not exist")
			return
		}
		log.Info(link)
		ids := re.FindAllString(link, 1)
		if len(ids) == 1 {
			log.Info(ids[0])
			items = append(items, ids[0])
		}
	})
	nexttag := doc.Find("div.pager div a")
	if nexttag.Length() == 0 {
		return
	}
	next := nexttag.Eq(0)
	if next.Text() == "下页" {
		nextlink, exist := next.Attr("href")
		if !exist {
			return
		}
		log.Info(nextlink)
		nextitems, nerr := getitems(nextlink)
		if nerr != nil {
			log.Error(err)
			err = nerr
			return
		}
		if len(nextitems) > 0 {
			items = append(items, nextitems...)
		}
	}

	return
}
func transUrl(shoplink string) string {
	//这里对URL进行转换，保证转换出来的url是wap版的店铺url
	if strings.Contains(shoplink, ".m.t") {
		return shoplink
	}
	shoplink = strings.Replace(shoplink, ".taobao.com", ".m.taobao.com", 1)
	shoplink = strings.Replace(shoplink, ".tmall.com", ".m.tmall.com", 1)
	log.Info("transform url is ", shoplink)
	return shoplink
}

func wrapUrl(itemslink string) (string, error) {
	//对所有宝贝链接解析和重新拼凑，形成一个
	URL, err := url.Parse(itemslink)
	if err != nil {
		return "", err
	}
	query := URL.Query()
	uid := query.Get("suid")
	index := strings.Index(itemslink, "/shop/")
	prefix := itemslink[:index+6]
	wrap := fmt.Sprintf("%sa-5-40-42-%s.htm", prefix, uid)
	log.Info(wrap)
	return wrap, nil
}
