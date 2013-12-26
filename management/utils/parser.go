package utils

import (
	//"Mango/management/crawler"
	//"Mango/management/filter"
	"Mango/management/models"
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/qiniu/log"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
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
	Brand      string            `json:"brand"`
	ShopType   string            `json:"shop_type"`
	DetailUrl  string            `json:"detail_url"`
}
type Loc struct {
	State string
	City  string
}

func Save(item *Info, mgocol *mgo.Collection) error {
	tItem := models.TaobaoItemStd{}
	change := bson.M{
		"detail_url":        item.DetailUrl,
		"title":             item.Title,
		"nick":              item.Nick,
		"desc":              item.Desc,
		"sid":               item.Sid,
		"cid":               item.Cid,
		"price":             item.Price,
		"location":          item.Location,
		"promotion_price":   item.Promprice,
		"shop_type":         item.ShopType,
		"reviews_count":     item.Reviews,
		"monthly_sales_num": item.Count,
		"props":             item.Attr,
		"item_imgs":         item.Imgs,
		"in_stock":          item.InStock,
	}
	err := mgocol.Find(bson.M{"num_iid": int(item.ItemId)}).One(&tItem)
	if err != nil {
		return err
	}
	if tItem.Title == "" {
		t := time.Now()
		change["data_updated_time"] = t
		change["data_last_revised_time"] = time.Now()

	} else {
		change["data_last_revised_time"] = time.Now()
	}
	err = mgocol.Update(bson.M{"num_iid": int(item.ItemId)}, bson.M{"$set": change})
	if err != nil {
		return err
	}
	return nil
}
func Post(info *Info) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	posturl := "http://10.0.1.23:8080/scheduler/api/send_item_detail?token=d61995660774083ccb8b533024f9b8bb"
	reader := strings.NewReader(string(data))
	log.Info(string(data))
	transport := &http.Transport{ResponseHeaderTimeout: time.Duration(30) * time.Second, DisableKeepAlives: true}
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

	log.Info(string(st))
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
		log.Info(err.Error())
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
	if shopid == itemid || shopid == "" {
		//由于之前代码错误导致部分商品的shopid为itemid
		link := GetShopLink(fontpage)
		id, e := GetShopid(link)
		if e != nil {
			log.Info(e)
		}
		shopid = id
	}
	sid, _ := strconv.Atoi(shopid)
	info.Sid = sid
	iid, _ := strconv.Atoi(itemid)
	info.ShopType = shoptype
	info.ItemId = iid
	info.Attr = detail.Attr
	info.UpdateTime = time.Now().Unix()
	return

}

//通过wap版某件商品的页面，获取其所属店铺的wap超链接
func GetShopLink(html string) string {
	//	re := regexp.MustCompile("\\<a href=\\\"(.+com).+进入店铺")
	if html == "" {
		return ""
	}
	reader := strings.NewReader(html)
	doc, e := goquery.NewDocumentFromReader(reader)
	if e != nil {
		log.Info(e.Error())
		return ""
	}
	shoptag := doc.Find("html body div.bd div.left-margin-5 p strong a")
	shoplink, exists := shoptag.Attr("href")
	if exists {
		return shoplink
	}
	return ""
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
		log.Info(e.Error())
		return info, e
	}

	titletag := doc.Find("title").Text()
	if strings.Contains(titletag, "商品屏蔽") {
		log.Info("商品屏蔽")
		err := errors.New("missing")
		return info, err
	}
	if titletag == "宝贝详情" {
		err := errors.New("missing")
		return info, err
	}
	if len(titletag) < 18 {
		err := errors.New("index out of range")
		return info, err
	}
	desc := titletag[0 : len(titletag)-18]
	info.Desc = desc
	info.Title = desc
	log.Info(desc)
	cattag := doc.Find("p.box")
	if len(cattag.Nodes) == 0 {
		//错误爬取了一些天猫触屏版的页面导致的
		err := errors.New("cattag")
		return info, err
	}
	cidtag := cattag.Find("a").Last()
	cidurl, exists := cidtag.Attr("href")
	if exists {
		if !strings.Contains(cidurl, "cat") {
			err := errors.New("聚划算")
			return info, err
		}
		re := regexp.MustCompile("\\d+$")
		cid := re.FindAllString(cidurl, -1)[0]
		c, err := strconv.Atoi(cid)
		if err != nil {
			return info, err
		}
		info.Cid = c
		log.Info(cid)

	}
	atags := cattag.Find("a")
	size := atags.Size()
	var catory []string
	for i := 1; i < size/2; i++ {
		catg := atags.Eq(i).Text()
		catory = append(catory, catg)
	}
	log.Info(catory)

	//details := doc.Find("div.detail")
	imgurltag := doc.Find("div.bd div.box div.detail p img")
	var imgs []string
	tmp, _ := imgurltag.Attr("alt")
	//可能会没有照片
	if tmp != "联系卖家" {
		imgurl, exists := imgurltag.Attr("src")
		if exists {
			re := regexp.MustCompile("_\\d+x\\d+\\.jpg|_b\\.jpg")
			imgurl := re.ReplaceAllString(imgurl, "")
			imgs = append(imgs, imgurl)
			log.Info(imgurl)
		}
		doc.Find("div.bd div.box div.detail table.mt tbody tr td a img").Each(func(i int, s *goquery.Selection) {
			re := regexp.MustCompile("_\\d+x\\d+\\.jpg|_b\\.jpg")
			src, exists := s.Attr("src")
			if exists {
				src = re.ReplaceAllString(src, "")
				imgs = append(imgs, src)
				log.Info(src)
			}
		})
	}
	info.Imgs = imgs
	instock := true
	detail := doc.Find("div.detail").Eq(1).Find("table")
	if detail.Size() == 1 {
		instock = false

		log.Info("已下架")
	}
	log.Info(instock)
	info.InStock = instock
	judgeindex := 1
	log.Info(imgs)
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
	log.Info("是不是二手", secondhand)
	if hasprom {
		log.Info("可能有促销价")
		prom := doc.Find("div.bd div.box div.detail p").Eq(1).Text()
		if prom != "" {
			re := regexp.MustCompile("\\d+\\.\\d+")
			proms := re.FindAllString(prom, -1)
			if len(proms) != 0 {
				log.Info("促销")
				prom := proms[0]
				p, err := strconv.ParseFloat(prom, 64)
				if err != nil {
					return info, err
				}
				info.Promprice = p
				log.Info("促销价", prom)

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
	log.Info(startindex)
	pricetag := doc.Find("body div.bd div.box div.detail p").Eq(startindex).Text()
	if pricetag != "" {
		rep := regexp.MustCompile("\\d+\\.\\d+")
		price := rep.FindAllString(pricetag, -1)
		if len(price) > 0 {
			log.Info("价格")
			p, err := strconv.ParseFloat(price[0], 64)
			if err != nil {
				return info, err
			}
			info.Price = p
			log.Info(price[0])
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
		log.Info("销量")
		log.Info(count[0])
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
		log.Info(location)
		loc := new(Loc)
		if len(location) == 6 || len(location) == 9 {
			loc.State = location
			loc.City = location
			info.Location = loc
			log.Info(location) //直辖市
		} else {
			if len(location) > 9 {
				if _, ok := states[location[0:6]]; ok {
					log.Info(location[0:6])
					loc.State = location[0:6]
					loc.City = location[6:]
					info.Location = loc
					log.Info(location[6:])

				} else if _, ok := states[location[0:9]]; ok {
					loc.State = location[0:9]
					loc.City = location[9:]
					log.Info(location[0:9])
					log.Info(location[9:])
					info.Location = loc
				}
			}

		}

	}
	if secondhand {
		reviewtag := doc.Find("div.bd div.box div.detail table.rate_desc tbody tr td.link_btn a").Text()
		log.Info(reviewtag)
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
			log.Info("评论")
			log.Info(r)

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
		log.Info("评论")
		log.Info(reviews)
	}

	nicktag, exists := doc.Find("body div.bd div.box div.detail p a img").Attr("src")
	log.Info(nicktag)
	p, _ := url.Parse(nicktag)
	q := p.Query()
	log.Info(q.Get("nick"))
	info.Nick = q.Get("nick")

	//由于代码错误导致有部分商品的店铺id与商品id是一样的

	return info, nil
}

func parsedetail(html string) (*Info, error) {
	reader := strings.NewReader(html)
	doc, e := goquery.NewDocumentFromReader(reader)
	detail := new(Info)
	if e != nil {
		log.Info(e.Error())
		return detail, e
	}
	reviewtag := doc.Find("body div.bd div.box p a.red strong").Text()
	if reviewtag == "" {
		err := errors.New("reviewtag is null")
		return detail, err
	}
	re := regexp.MustCompile("\\d+")
	reve := re.FindAllString(reviewtag, -1)
	log.Info(reve)
	if len(reve) != 0 {
		revie := reve[0]
		reviews, _ := strconv.Atoi(revie)
		log.Info(reviews)
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
