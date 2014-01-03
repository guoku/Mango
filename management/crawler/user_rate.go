package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/jason-zou/taobaosdk/rest"
	"github.com/qiniu/iconv"
	"github.com/qiniu/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//通过店铺链接，提取店铺详细数据
func FetchShopDetail(shoplink string) (*rest.Shop, error) {
	shop := new(rest.Shop)
	detail, err := GetShopInfo(shoplink)
	if err != nil {
		return shop, err
	}
	shop.Nick = detail.Nick
	shop.Title = detail.Title
	shop.PicPath = detail.PicPath
	shop.Sid = detail.Sid
	shop.ShopScore = detail.ShopScore

	userid, err := GetUserid(shoplink)
	if err != nil {
		log.Error(err)
		return shop, err
	}
	detail2, err := ParseShop(userid)
	if err != nil {
		log.Error(err)
		return shop, err
	}
	shop.ShopType = detail2.ShopType
	shop.UpdatedTime = time.Now()
	shop.Company = detail2.Company
	shop.Location = detail2.Location
	shop.MainProducts = detail2.MainProducts
	shop.ShopScore = detail2.ShopScore
	log.Infof("%+v", shop)
	return shop, nil
}

//通过这个函数，可以获取淘宝店的昵称，名称，图片，sid
func GetShopInfo(shoplink string) (*rest.Shop, error) {
	//re := regexp.MustCompile("http://[A-Za-z0-9]+\\.(taobao|tmall)\\.com")
	log.Info("raw", shoplink)
	//	shopurl := re.FindString(shoplink)
	//	log.Info("shopurl", shopurl)
	link := ""
	if strings.Contains(shoplink, ".m.") {
		link = shoplink
	} else {
		link = strings.Replace(shoplink, ".", ".m.", 1)
	}
	log.Info("shop link", link)
	doc, e := goquery.NewDocument(link)
	if e != nil {
		log.Println(e.Error())
		return nil, e
	}

	titletag := doc.Find("p.box").Text()
	title := titletag[7:]
	title = strings.TrimSpace(title)
	log.Println(title)

	derate := doc.Find("td[valign=top]")
	red := regexp.MustCompile("好评率：([0-9\\.]+)")
	tmp := red.FindStringSubmatch(derate.Text())
	praiseRate := "0.0"
	if len(tmp) >= 2 {
		praiseRate = tmp[1]
	}
	log.Info(praiseRate)
	pictag := doc.Find("td.pic img")
	piclink, _ := pictag.Attr("src")
	wwimg := doc.Find("img[alt=ww]")
	src, _ := wwimg.Attr("src")
	srcparse, _ := url.Parse(src)
	srcquery := srcparse.Query()
	nick := srcquery.Get("nick")
	log.Println(nick)
	sidtag, exists := wwimg.Parent().Attr("href")
	var sid int
	if exists {
		URL, _ := url.Parse(sidtag)
		query := URL.Query()
		sid, _ = strconv.Atoi(query.Get("shopId"))
	}
	var shopinfo *rest.Shop = new(rest.Shop)
	var shopscore *rest.ShopScore = new(rest.ShopScore)
	log.Info("praiseRate ", praiseRate)
	praiseRate = strings.TrimSpace(praiseRate)
	prate, err := strconv.ParseFloat(praiseRate, 32)
	if err != nil {
		log.Error(err)
	}
	log.Info(prate)
	shopscore.PraiseRate = float32(prate)
	shopinfo = &rest.Shop{Nick: nick, Title: title, PicPath: piclink, Sid: sid, ShopScore: shopscore}
	return shopinfo, nil
}

//获取店主的旺旺id，通过这个id，可以看到其评分页面
func GetUserid(shoplink string) (string, error) {
	transport := &http.Transport{ResponseHeaderTimeout: time.Duration(30) * time.Second}
	var redirectFunc = func(req *http.Request, via []*http.Request) error {
		redirectUrl := req.URL.String()
		log.Info("redirectUrl ", redirectUrl)
		return nil
	}
	if strings.Contains(shoplink, ".m.") {
		shoplink = strings.Replace(shoplink, ".m.", ".", 1)
	}
	log.Info(shoplink)
	client := &http.Client{Transport: transport, CheckRedirect: redirectFunc}
	req, err := http.NewRequest("GET", shoplink, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:24.0) Gecko/20100101 Firefox.24.0")
	if err != nil {
		log.Error(err)
		return "", err
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
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
	re := regexp.MustCompile("userId=(?P<id>\\d+)")
	//log.Info(string(html))
	log.Info("shop link", shoplink)
	userId := re.FindStringSubmatch(string(html))[1]
	return userId, nil
}

//解析评价数据页面的信息
//userid是店主的旺旺ID
func ParseShop(userid string) (*rest.Shop, error) {
	shop := new(rest.Shop)
	shop.ShopType = "taobao.com"
	link := fmt.Sprintf("http://rate.taobao.com/user-rate-%s.htm", userid)
	log.Info(link)
	transport := &http.Transport{ResponseHeaderTimeout: time.Duration(30) * time.Second}
	client := &http.Client{Transport: transport}
	req, err := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:24.0) Gecko/20100101 Firefox.24.0")
	if err != nil {
		log.Error(err)
		return shop, err
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Error(err)
		return shop, err
	}
	cd, err := iconv.Open("utf-8", "gbk")
	if err != nil {
		log.Fatal(err)
	}
	defer cd.Close()
	html, _ := ioutil.ReadAll(resp.Body)
	shtml := cd.ConvString(string(html))
	reader := strings.NewReader(shtml)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return shop, err
	}
	//提取店铺所属公司名称
	company := doc.Find("ul li.company div.fleft2")
	if company.Length() > 0 {
		companyName := company.Last().Text()
		companyName = strings.TrimSpace(companyName)
		log.Info(companyName)
		shop.Company = companyName
		shop.ShopType = "tmall.com"
	}
	//提取主营业务和所在地区
	dm := doc.Find("div#shop-rate-box.shop-rate-box div.personal-info div.col-sub div.left-box div.bd div.info-block ul li")
	re := regexp.MustCompile("当前主营：(.+)")
	dt := re.FindStringSubmatch(dm.Eq(1).Text())
	if len(dt) < 2 {
		dt = re.FindStringSubmatch(dm.Eq(0).Text())
	}
	domain := dt[1]
	log.Info(domain)
	domain = strings.TrimSpace(domain)
	shop.MainProducts = domain
	loc := dm.Eq(2).Text()
	re = regexp.MustCompile("所在地区：[\\pC ]+(\\pL+)")
	lc := re.FindStringSubmatch(loc)
	if len(lc) >= 2 {
		loc = lc[1]
	} else {
		loc = dm.Eq(1).Text()
		log.Info(loc)
		lc = re.FindStringSubmatch(loc)
		if len(lc) < 2 {
			loc = ""
		} else {
			loc = lc[1]

		}
	}
	loc = strings.TrimSpace(loc)
	shop.Location = loc
	log.Info(loc)

	//提取30天内服务情况
	var semiScore []*rest.RateScore
	semiservice := doc.Find("div.personal-info div.personal-rating div.main-wrap div.rate-box div.bd ul#dsr.dsr-info li.J_RateInfoTrigger div.item-scrib")
	semiservice.Each(func(i int, st *goquery.Selection) {
		ratescore := new(rest.RateScore)
		st.Find("em").Each(func(i int, em *goquery.Selection) {
			//log.Info(em.Text())
			if i == 0 {
				log.Info("得分", em.Text())
				f, err := strconv.ParseFloat(em.Text(), 32)
				if err != nil {
					log.Error(err)
				} else {
					ratescore.Score = float32(f)
				}

			} else {
				percent, exist := em.Find("strong").Attr("class")
				if exist {
					if percent == "percent over" {
						t := em.Text()
						t = strings.TrimSpace(t)
						log.Info("高于:", t)
						if t == "----" {
							//持平
							ratescore.Rate = 0.0
						} else {
							f, err := strconv.ParseFloat(t[:len(t)-1], 32)
							if err != nil {
								log.Error(err)
							} else {
								ratescore.Rate = float32(f)
								log.Info(ratescore.Rate)
							}
						}
					} else {
						t := em.Text()
						t = strings.TrimSpace(t)
						log.Info("低于：", em.Text())
						if t == "----" {
							ratescore.Rate = 0.0
						} else {
							f, err := strconv.ParseFloat(t[:len(t)-1], 32)
							if err != nil {
								log.Error(err)
							} else {
								ratescore.Rate = 0.0 - float32(f)
								log.Info(ratescore.Rate)
							}
						}
					}
				}
			}
		})
		semiScore = append(semiScore, ratescore)
	})
	shopkps := rest.ShopKPS{DescScore: semiScore[0], ServiceScore: semiScore[1], DeliveryScore: semiScore[2]}
	//提取30天内服务情况
	serviceScore := new(rest.ShopService)
	serviceLink := fmt.Sprintf("http://rate.taobao.com/ShopService4C.htm?userNumId=%s", userid)
	resp, err = http.Get(serviceLink)
	defer resp.Body.Close()
	if err != nil {
		log.Error(err)
	} else {
		data, _ := ioutil.ReadAll(resp.Body)
		jdata := string(data)
		re = regexp.MustCompile("\\\"([0-9\\.]+)\\\"")
		var f = func(repl string) string {
			return repl[1 : len(repl)-1]
		}
		jdata = re.ReplaceAllStringFunc(jdata, f)
		log.Info(jdata)
		err = json.Unmarshal([]byte(jdata), serviceScore)
		if err != nil {
			log.Error(err)
		}
	}

	shopscore := rest.ShopScore{SemiScore: &shopkps, ServiceScore: serviceScore}
	shop.ShopScore = &shopscore
	log.Infof("%+v", shop)
	return shop, nil
}
