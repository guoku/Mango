package utils

import (
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var states map[string]bool = map[string]bool{
	"北京":  true,
	"上海":  true,
	"天津":  true,
	"重庆":  true,
	"广东":  true,
	"江苏":  true,
	"山东":  true,
	"浙江":  true,
	"河北":  true,
	"山西":  true,
	"辽宁":  true,
	"吉林":  true,
	"河南":  true,
	"安徽":  true,
	"福建":  true,
	"江西":  true,
	"黑龙江": true,
	"湖南":  true,
	"湖北":  true,
	"海南":  true,
	"四川":  true,
	"贵州":  true,
	"云南":  true,
	"陕西":  true,
	"甘肃":  true,
	"青海":  true,
	"台湾":  true,
	"西藏":  true,
	"内蒙古": true,
	"广西":  true,
	"宁夏":  true,
	"新疆":  true,
	"香港":  true,
	"澳门":  true,
	"海外":  true,
}

type Info struct {
	Desc       string            `json:"desc"`
	Cid        int               `json:"cid"`
	Promprice  float64           `json:"promotion_price"`
	Price      float64           `json:"price"`
	Imgs       []string          `json:"item_imgs"`
	Count      int               `json:"monthly_sales_volume"`
	Reviews    int               `json:"reviews_count"`
	Nick       string            `json:"nick"`
	InStock    bool              `json:"in_stock"`
	Attr       map[string]string `json:"props"`
	Location   *Loc              `json:"location"`
	UpdateTime int64             `json:"data_updated_time"`
	ItemId     int               `json:"num_iid"`
	Sid        int               `json:"sid"`
	Title      string            `json:"title"`
	ShopType   string            `json:"shop_type"`
}
type Loc struct {
	State string
	City  string
}

func Post(info *Info) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	posturl := "http://10.0.1.23:8080/scheduler/api/send_item_detail?token=d61995660774083ccb8b533024f9b8bb"
	reader := strings.NewReader(string(data))
	log.Println(string(data))
	transport := &http.Transport{ResponseHeaderTimeout: time.Duration(30) * time.Second}
	var DefaultClinet = &http.Client{Transport: transport}
	resp, err := DefaultClinet.Post(posturl, "application/json", reader)
	if err != nil {
		return err
	}
	st, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println(string(st))
	return nil
}
func ParseWithoutID(fontpage, detailpage string) (info *Info, missing bool, err error) {
	return Parse(fontpage, detailpage, "", "", "")
}
func Parse(fontpage, detailpage, itemid, shopid, shoptype string) (info *Info, missing bool, err error) {
	font, err := parsefontpage(fontpage)
	info = new(Info)
	missing = false
	if err != nil {
		log.Println(err.Error())
		if err.Error() == "missing" {
			//抓取的页面属于屏蔽的页面
			missing = true
		}
		return
	}
	detail, err := parsedetail(detailpage)
	if err != nil {
		return
	}
	info = font
	sid, _ := strconv.Atoi(itemid)
	info.Sid = sid
	iid, _ := strconv.Atoi(itemid)
	info.ShopType = shoptype
	info.ItemId = iid
	info.Attr = detail.Attr
	info.UpdateTime = time.Now().Unix()
	return

}
func parsefontpage(html string) (*Info, error) {
	info := new(Info)
	if html == "" {
		err := errors.New("null")
		return info, err
	}
	reader := strings.NewReader(html)
	doc, e := goquery.NewDocumentFromReader(reader)
	if e != nil {
		log.Println(e.Error())
		return info, e
	}

	titletag := doc.Find("title").Text()
	if strings.Contains(titletag, "商品屏蔽") {
		log.Println("商品屏蔽")
		err := errors.New("missing")
		return info, err
	}
	if titletag == "宝贝详情" {
		err := errors.New("missing")
		return info, err
	}
	desc := titletag[0 : len(titletag)-18]
	info.Desc = desc
	info.Title = desc
	log.Println(desc)
	cattag := doc.Find("p.box")
	if len(cattag.Nodes) == 0 {
		//错误爬取了一些天猫触屏版的页面导致的
		err := errors.New("cattag")
		return info, err
	}
	cidtag := cattag.Find("a").Last()
	cidurl, exists := cidtag.Attr("href")
	if exists {
		re := regexp.MustCompile("\\d+$")
		cid := re.FindAllString(cidurl, -1)[0]
		c, err := strconv.Atoi(cid)
		if err != nil {
			return info, err
		}
		info.Cid = c
		log.Println(cid)

	}
	atags := cattag.Find("a")
	size := atags.Size()
	var catory []string
	for i := 1; i < size/2; i++ {
		catg := atags.Eq(i).Text()
		catory = append(catory, catg)
	}
	log.Println(catory)

	//details := doc.Find("div.detail")
	imgurltag := doc.Find("div.bd div.box div.detail p img")
	var imgs []string
	tmp, _ := imgurltag.Attr("alt")
	//可能会没有照片
	if tmp != "联系卖家" {
		imgurl, exists := imgurltag.Attr("src")
		if exists {
			re := regexp.MustCompile("_\\d+x\\d+\\.jpg")
			imgurl := re.ReplaceAllString(imgurl, "")
			imgs = append(imgs, imgurl)
			log.Println(imgurl)
		}
		doc.Find("div.bd div.box div.detail table.mt tbody tr td a img").Each(func(i int, s *goquery.Selection) {
			re := regexp.MustCompile("_\\d+x\\d+\\.jpg")
			src, exists := s.Attr("src")
			if exists {
				src = re.ReplaceAllString(src, "")
				imgs = append(imgs, src)
				log.Println(src)
			}
		})
	}
	info.Imgs = imgs
	instock := true
	detail := doc.Find("div.detail").Eq(1).Find("table")
	if detail.Size() == 1 {
		instock = false

		log.Println("已下架")
	}
	log.Println(instock)
	info.InStock = instock
	judgeindex := 1
	log.Println(imgs)
	if len(imgs) == 0 {
		judgeindex = 0
	}
	judge := doc.Find("div.detail p").Eq(judgeindex).Text()
	hasprom := true
	secondhand := false
	if strings.Contains(judge, "格：") {
		hasprom = false
	} else {
		if strings.Contains(judge, "价：") {
			secondhand = true
		}
	}
	log.Println("是不是二手", secondhand)
	if hasprom {
		log.Println("可能有促销价")
		prom := doc.Find("div.bd div.box div.detail p").Eq(1).Text()
		if prom != "" {
			re := regexp.MustCompile("\\d+\\.\\d+")
			proms := re.FindAllString(prom, -1)
			if len(proms) != 0 {
				log.Println("促销")
				prom := proms[0]
				p, err := strconv.ParseFloat(prom, 64)
				if err != nil {
					return info, err
				}
				info.Promprice = p
				log.Println("促销价", prom)

			}
		}
	}

	startindex := 2 //默认第二个开始才是价格，第一个是促销价
	if hasprom == false {
		startindex = 1
	}
	if len(imgs) == 0 {
		//没有图片的情况
		startindex = 0
	}
	log.Println(startindex)
	pricetag := doc.Find("body div.bd div.box div.detail p").Eq(startindex).Text()
	if pricetag != "" {
		rep := regexp.MustCompile("\\d+\\.\\d+")
		price := rep.FindAllString(pricetag, -1)
		if len(price) > 0 {
			log.Println("价格")
			p, err := strconv.ParseFloat(price[0], 64)
			if err != nil {
				return info, err
			}
			info.Price = p
			log.Println(price[0])
		}
	}

	counttag := doc.Find("div.detail p").Eq(startindex + 2).Text()
	countre := regexp.MustCompile("\\d+")
	count := countre.FindAllString(counttag, -1)
	if len(count) > 0 {
		c, err := strconv.Atoi(count[0])
		if err != nil {
			return info, err
		}
		info.Count = c
		log.Println("销量")
		log.Println(count[0])
	} else {
		if strings.Contains(counttag, "量") {
			err := errors.New("count is missing")
			return info, err
		}
	}

	loctag := doc.Find("div.detail p").Eq(startindex + 3).Text()
	if strings.Contains(loctag, "地") {
		location := strings.Split(loctag, "：")[1]
		location = strings.TrimSpace(location)
		log.Println(location)
		loc := new(Loc)
		if len(location) == 6 || len(location) == 9 {
			loc.State = location
			loc.City = location
			info.Location = loc
			log.Println(location) //直辖市
		} else {
			if len(location) > 9 {
				if _, ok := states[location[0:6]]; ok {
					log.Println(location[0:6])
					loc.State = location[0:6]
					loc.City = location[6:]
					info.Location = loc
					log.Println(location[6:])

				} else if _, ok := states[location[0:9]]; ok {
					loc.State = location[0:9]
					loc.City = location[9:]
					log.Println(location[0:9])
					log.Println(location[9:])
					info.Location = loc
				}
			}

		}

	}
	if secondhand {
		reviewtag := doc.Find("div.bd div.box div.detail table.rate_desc tbody tr td.link_btn a").Text()
		log.Println(reviewtag)
		rege := regexp.MustCompile("\\d+")
		reviewarray := rege.FindAllString(reviewtag, -1)
		if len(reviewarray) == 0 {
			info.Reviews = 0
		} else {
			r, err := strconv.Atoi(reviewarray[0])
			if err != nil {
				return info, err
			}
			info.Reviews = r
			log.Println("评论")
			log.Println(r)

		}
	} else {
		fixtag := doc.Find("div.bd div.box div.detail table.rate_desc tbody tr td.link_btn a span").Text()
		rege := regexp.MustCompile("\\d+")
		reviews := rege.FindAllString(fixtag, -1)[0]
		r, err := strconv.Atoi(reviews)
		if err != nil {
			return info, err
		}
		info.Reviews = r
		log.Println("评论")
		log.Println(reviews)
	}

	nicktag, exists := doc.Find("body div.bd div.box div.detail p a img").Attr("src")
	log.Println(nicktag)
	p, _ := url.Parse(nicktag)
	q := p.Query()
	log.Println(q.Get("nick"))
	info.Nick = q.Get("nick")
	return info, nil
}

func parsedetail(html string) (*Info, error) {
	reader := strings.NewReader(html)
	doc, e := goquery.NewDocumentFromReader(reader)
	detail := new(Info)
	if e != nil {
		log.Println(e.Error())
		return detail, e
	}
	reviewtag := doc.Find("body div.bd div.box p a.red strong").Text()
	if reviewtag == "" {
		err := errors.New("reviewtag is null")
		return detail, err
	}
	re := regexp.MustCompile("\\d+")
	reve := re.FindAllString(reviewtag, -1)
	log.Println(reve)
	if len(reve) != 0 {
		revie := reve[0]
		reviews, _ := strconv.Atoi(revie)
		log.Println(reviews)
		detail.Reviews = reviews
	}
	attr := doc.Find("body div.bd div#itemProp.box div.detail table.goods-property tbody tr")
	attrs := make(map[string]string)
	attr.Each(func(i int, s *goquery.Selection) {
		td := s.Find("td")
		key := strings.TrimSpace(td.Eq(0).Text())
		key = strings.Replace(key, "\r\n", " ", -1)
		key = strings.Replace(key, "\n", " ", -1)
		value := strings.TrimSpace(td.Eq(1).Text())
		value = strings.Replace(value, "\r\n", " ", -1)
		value = strings.Replace(value, "\n", " ", -1)
		attrs[key] = value
	})
	detail.Attr = attrs
	return detail, nil
}
