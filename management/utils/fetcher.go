package utils

import (
	"errors"
	"github.com/qiniu/log"
	"io/ioutil"
	"labix.org/v2/mgo"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Pages struct {
	ShopId     string
	ItemId     string
	FontPage   string
	DetailPage string
	ShopType   string
	UpdateTime int64
	Parsed     bool
	InStock    bool //是否下架了
}

type FailedPages struct {
	ShopId     string
	ItemId     string
	ShopType   string
	UpdateTime int64
	InStock    bool
}

type ShopItem struct {
	Date       time.Time
	Items_list []string
	Items_num  int
	Shop_id    int
	State      string
}

var proxys []string = []string{
	"http://127.0.0.1:30048",
	"http://127.0.0.1:30049",
	"http://127.0.0.1:30050",
	"http://127.0.0.1:30051",
	"http://127.0.0.1:30052",
	"http://127.0.0.1:30053",
	"http://127.0.0.1:30054",
	"http://127.0.0.1:30055",
	"http://127.0.0.1:30056",
	"http://127.0.0.1:30057",
	"http://127.0.0.1:30058",
	"http://127.0.0.1:30059",
	"http://127.0.0.1:30060",
	"http://127.0.0.1:30061",
	"http://127.0.0.1:30062",
	"http://127.0.0.1:30063",
	"http://127.0.0.1:30064",
	"http://127.0.0.1:30065",
	"http://127.0.0.1:30066",
	"http://127.0.0.1:30067",
	"http://127.0.0.1:30068",
	"http://127.0.0.1:30069",
	"http://127.0.0.1:30070",
	"http://127.0.0.1:30071",
	"http://127.0.0.1:30072",
	"http://127.0.0.1:30073",
	"http://127.0.0.1:30074",
	"http://127.0.0.1:30075",
	"http://127.0.0.1:30076",
	"http://127.0.0.1:30077",
}

func getTransport() (transport *http.Transport) {
	length := len(proxys)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	proxy := proxys[r.Intn(length)]
	url_i := url.URL{}
	url_proxy, _ := url_i.Parse(proxy)
	transport = &http.Transport{Proxy: http.ProxyURL(url_proxy), ResponseHeaderTimeout: time.Duration(30) * time.Second, DisableKeepAlives: true}
	return
}

func Fetch(itemid string, shoptype string) (html string, err error, detail string) {
	url := ""
	detailurl := ""
	if shoptype == "tmall.com" {
		url = "http://a.m.tmall.com/i" + itemid + ".htm"
		detailurl = "http://a.m.tmall.com/da" + itemid + ".htm"
	} else {
		url = "http://a.m.taobao.com/i" + itemid + ".htm"
		detailurl = "http://a.m.taobao.com/da" + itemid + ".htm"
	}
	transport := getTransport()
	client := &http.Client{Transport: transport}
	req, err := http.NewRequest("GET", url, nil)
	useragent := userAgentGen()
	req.Header.Set("User-Agent", useragent)
	if err != nil {
		log.Print(err.Error())
		return "", err, ""
	}
	log.Info("start to do request")
	resp, err := client.Do(req)
	log.Info("request has been done")
	if err != nil {
		if resp == nil {
			log.Info("当proxy不可达时，resp为空")
		}
		time.Sleep(1 * time.Second)
		log.Info(err.Error())
		return "", err, ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		//fmt.Println(resp.Request.URL.String())
		resplink := resp.Request.URL.String()
		if strings.Contains(resplink, "h5") {
			html = ""
			detail = ""
			err = errors.New("taobao forbidden")
			log.Info("taobao forbidden")
			return
		}
		robots, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Info(err.Error())
			return "", err, ""
		}
		html = string(robots)
	} else {
		log.Info(resp.Status)
		html = ""
		err = errors.New("404")
		return html, err, ""
	}
	resp.Body.Close()
	req, err = http.NewRequest("GET", detailurl, nil)
	req.Header.Set("User-Agent", useragent)
	if err != nil {
		log.Print(err.Error())
		return "", err, ""
	}
	resp, err = client.Do(req)
	if err != nil {
		log.Info(err.Error())
		return "", err, ""
	}
	if resp.StatusCode == 200 {
		//fmt.Println(resp.Request.URL.String())
		resplink := resp.Request.URL.String()
		if strings.Contains(resplink, "h5") {
			html = ""
			detail = ""
			err = errors.New("taobao forbidden")
			log.Info("taobao forbidden")
			return
		}
		robots, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Info(err.Error())
			return "", err, ""
		}
		detail = string(robots)
	} else {
		log.Info(resp.StatusCode)
		html = ""
		err = errors.New(resp.Status)
		return html, err, ""
	}
	resp.Body.Close()
	re := regexp.MustCompile("\\<style[\\S\\s]+?\\</style\\>")
	html = re.ReplaceAllString(html, "")
	re = regexp.MustCompile("\\<script[\\S\\s]+?\\</script\\>")
	detail = re.ReplaceAllString(detail, "")
	return
}

func IsTmall(itemid string) (bool, error) {
	url := "http://a.m.taobao.com/i" + itemid + ".htm"
	request, _ := http.NewRequest("GET", url, nil)
	transport := getTransport()
	client := &http.Client{Transport: transport}
	resp, err := client.Do(request)
	if err != nil {
		return false, err
	} else {
		finalURL := resp.Request.URL.String()
		if finalURL == url {
			return false, nil
		} else {
			return true, nil
		}
	}
	resp.Body.Close()
	return true, nil
}
func MongoInit(host, db, collection string) *mgo.Collection {
	session, err := mgo.Dial(host)
	if err != nil {
		log.Info("严重错误")
		panic(err)
	}
	return session.DB(db).C(collection)
}

var UserAgents []string = []string{
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:24.0) Gecko/20100101 Firefox.24.0",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:22.0) Gecko/20100101 Firefox/22.0",
}

func userAgentGen() string {
	length := len(UserAgents)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return UserAgents[r.Intn(length)]
}
