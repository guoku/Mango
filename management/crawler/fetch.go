package crawler

import (
	"errors"
	"fmt"
	"github.com/qiniu/log"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

//返回值font是商品的页面，detail是商品的详情页面，instock表示下架与否
//这里下架与否的判断设计得比较不好,如果抓取正常，instock是未知的，只有进行解析后才知道结果
//而如果出现了err，则应该看instock与否，如果下架了，这个itemid就不需要保存了
//这个是对Fetch的封装，因为返回的错误类型需要用来判断是否要保存这个item
func FetchItem(itemid string, shoptype string) (font, detail string, instock bool, err error) {
	log.Infof("start to fetch %s", itemid)
	font, err, detail = Fetch(itemid, shoptype)
	if err != nil {
		log.Infof("%s failed", itemid)
		if err.Error() != "404" {
			//说明不是因为商品下架而导致的失败
			instock = true
			return
		} else {
			//商品的页面已经找不到了
			instock = false
			return
		}
	}
	log.Infof("%s fetched successed!", itemid)
	return
}

//这种方式在天猫商品里会多一次访问，所以建议少用
func FetchWithOutType(itemid string) (html, detail, shoptype string, instock bool, err error) {
	shoplink := fmt.Sprintf("http://a.m.taobao.com/i%s.htm", itemid)
	instock = true
	//transport := getTransport()
	//client := &http.Client{Transport: transport}
	client := &http.Client{}
	req, err := http.NewRequest("GET", shoplink, nil)
	if err != nil {
		log.Error(err)
		return
	}
	useragent := userAgentGen()
	req.Header.Set("User-Agent", useragent)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	if resp == nil {
		err = errors.New("response is nil")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		resplink := resp.Request.URL.String()
		if strings.Contains(resplink, "h5") {
			err = errors.New("taobao forbidden")
			log.Error(err)
			return
		}
		if resplink != shoplink {
			shoptype = "tmall.com"
		} else {
			shoptype = "taobao.com"
		}
		bytedata, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			return "", "", "", true, err
		}
		html = string(bytedata)

	} else {
		instock = false
		log.Info(resp.Status)
		err = errors.New("404")
		return
	}
	resp.Body.Close()
	detailurl := ""
	if shoptype == "taobao.com" {
		detailurl = fmt.Sprintf("http://a.m.taobao.com/da%s.htm", itemid)
	} else {
		detailurl = fmt.Sprintf("http://a.m.tmall.com/da%s.htm", itemid)
	}

	req, err = http.NewRequest("GET", detailurl, nil)
	if err != nil {
		log.Error(err)
		return
	}
	req.Header.Set("User-Agent", useragent)
	resp, err = client.Do(req)
	if err != nil || resp == nil {
		log.Error(err)
		return
	}
	if resp.StatusCode == 200 {
		resplink := resp.Request.URL.String()
		if strings.Contains(resplink, "h5") {
			err = errors.New("taobao forbidden")
			log.Error(err)
			return
		}
		bytedata, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			return "", "", "", true, err
		}
		detail = string(bytedata)
	} else {
		log.Info(resp.StatusCode)
		err = errors.New(resp.Status)
		instock = false
		return
	}

	resp.Body.Close()
	re := regexp.MustCompile("\\<style[\\S\\s]+?\\</style\\>")
	re2 := regexp.MustCompile("\\<script[\\S\\s]+?\\</script\\>")
	html = re.ReplaceAllString(html, "")
	detail = re.ReplaceAllString(detail, "")
	html = re2.ReplaceAllString(html, "")
	detail = re2.ReplaceAllString(detail, "")
	err = nil
	return
}

//根据商品id和店铺类型抓取页面
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
	re2 := regexp.MustCompile("\\<script[\\S\\s]+?\\</script\\>")
	html = re.ReplaceAllString(html, "")
	detail = re.ReplaceAllString(detail, "")
	html = re2.ReplaceAllString(html, "")
	detail = re2.ReplaceAllString(detail, "")
	err = nil
	return
}
