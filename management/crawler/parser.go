package crawler

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/qiniu/iconv"
	"github.com/qiniu/log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	IMG_POSTFIX string = "_\\d+x\\d+.*\\.jpg|_b\\.jpg"
)

func ParseWithoutID(fontpage, detailpage string) (info *Info, missing bool, err error) {
	return Parse(fontpage, detailpage, "", "", "")
}

//如果返回错误，应该看是否下架，下架了就不保存了
//这是对Parse的封装，因为要根据Parse返回的error，判断是否要保存这个item
func ParsePage(font, detail, itemid, shopid, shoptype string) (info *Info, instock bool, err error) {
	info, missing, err := Parse(font, detail, itemid, shopid, shoptype)
	log.Info("解析完毕")
	instock = true
	if err != nil {
		if missing {
			instock = false
			return
		} else if err.Error() == "聚划算" || err.Error() == "cattag" {
			//商品来自聚划算或者找不着了
			//直接丢弃，不予以保存
			instock = false
			return
		} else {
			//只是解析错误，出现了新情况，暂时不管先
			log.Error("解析错误", itemid)
			log.Error(err)

			return
		}
	} else {
		instock = info.InStock
		//即使解析出来的结果还是下架的，但这个结果要保存
		//因为淘宝的这种下架方式还可以查看页面，就是不能购买
		//且有货了之后可能会重新上架
		return
	}
}

func ParseWeb(fontpage, detailpage, itemid, shopid, shoptype string) (info *Info, err error) {
	if shoptype == "taobao.com" {
		info, err = ParseWebFontTaobao(fontpage)
		if err != nil {
			log.Error(err)
			log.Error("解析错误", itemid)
			return
		}
	} else {
		//可能是天猫，也可能没有传人店铺类型
		info, err = ParseWebFontTmall(fontpage)
		if err != nil {
			info, err = ParseWebFontTaobao(fontpage)
			if err != nil {
				log.Error(err)
				log.Error("解析错误", itemid)
				return
			}

		}
	}
	/*
		detail, err := parsedetail(detailpage)
		if err != nil {
			log.Error(err)
			log.Error("解析detail页面出错", itemid)
			return
		}
	*/
	sid, _ := strconv.Atoi(shopid)
	info.Sid = sid
	iid, _ := strconv.Atoi(itemid)
	info.ItemId = iid
	info.ShopType = shoptype
	info.UpdateTime = time.Now().Unix()
	return
}

//解析页面数据
func Parse(fontpage, detailpage, itemid, shopid, shoptype string) (info *Info, missing bool, err error) {
	font, err := parsefontpage(fontpage)
	info = new(Info)
	missing = false
	if err != nil {
		if err.Error() == "missing" {
			//抓取的页面属于屏蔽的页面
			missing = true
		}
		log.Error(err, itemid)

		return
	}
	detail, err := parsedetail(detailpage)
	if err != nil {
		log.Error("解析详情页出错", itemid)
		log.Error(err)
		return
	}
	info = font
	if shopid == itemid || shopid == "" {
		//由于之前代码错误导致部分商品的shopid为itemid
		link, _ := GetShopLink(fontpage)
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
		log.Error(err)
		log.Error(titletag)
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
			log.Error(err)
			log.Error(desc)
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
					log.Error(err)
					log.Error(desc)
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
		reviews := rege.FindAllString(fixtag, -1)
		review := "0"
		if len(reviews) > 0 {
			review = reviews[0]
		}
		r, err := strconv.Atoi(review)
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
	/*
		shoptag := doc.Find("html body div.bd div.left-margin-5 p strong a")
		shoplink, exists := shoptag.Attr("href")
		if exists {
			info.DetailUrl = shoplink
		}
	*/
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

func ParseWebFontTmall(fonthtml string) (*Info, error) {
	reader := strings.NewReader(fonthtml)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	re := regexp.MustCompile("categoryId':'(\\d+)")
	cats := re.FindStringSubmatch(fonthtml)
	if len(cats) == 0 {
		err = errors.New("categoryId is not exist")
		return nil, err
	}
	cid, _ := strconv.Atoi(cats[1])
	log.Info("cid", cid)
	re = regexp.MustCompile("sellerNickName\" : '(.+)'")
	seller := re.FindStringSubmatch(fonthtml)
	if len(seller) == 0 {
		err = errors.New("seller nick is not exist")
		return nil, err
	}
	nick := seller[1]
	nick, err = url.QueryUnescape(nick)
	if err != nil {
		log.Error(err)
		return nil, nil
	}
	log.Info("nick, ", nick)
	re = regexp.MustCompile("reservePrice' : '(\\d+\\.\\d+)")
	pricetag := re.FindStringSubmatch(fonthtml)
	if len(pricetag) == 0 {
		err = errors.New("price is not exist")
		return nil, err
	}
	price, _ := strconv.ParseFloat(pricetag[1], 64)
	log.Info("price ", price)
	re = regexp.MustCompile("isOnline':(true|false)")
	online := re.FindStringSubmatch(fonthtml)
	var isOnline bool
	if len(online) == 0 {
		isOnline = true
	}
	isOnline, _ = strconv.ParseBool(online[1])
	log.Info("is online", isOnline)
	var imgs []string
	re = regexp.MustCompile(IMG_POSTFIX)
	fimg := doc.Find("img#J_ImgBooth")
	if fimg.Length() == 0 {
		err = errors.New("no img")
		log.Error(err)
		return nil, err
	}
	fimgurl, _ := fimg.Attr("src")
	fimgurl = re.ReplaceAllString(fimgurl, "")
	imgs = append(imgs, fimgurl)
	log.Info("fimgurl", fimgurl)
	optimgs := doc.Find("ul#J_UlThumb li a img")
	optimgs.Each(func(i int, sq *goquery.Selection) {
		op, _ := sq.Attr("src")
		op = re.ReplaceAllString(op, "")
		log.Info(op)
		imgs = append(imgs, op)
	})
	cd, _ := iconv.Open("utf-8", "gbk")
	defer cd.Close()
	titletg := doc.Find("title")
	title := titletg.Text()
	title = cd.ConvString(title)
	rt := []rune(title)
	rtsub := rt[0 : len(rt)-12]
	title = string(rtsub)
	log.Info(title)
	attrs := make(map[string]string)
	doc.Find("div.attributes div ul#J_AttrUL li").Each(func(i int, sq *goquery.Selection) {
		att := sq.Text()
		att = strings.TrimSpace(att)
		kv := strings.Split(att, ":")
		attrs[kv[0]] = kv[1]
	})
	data := Info{
		Title:     title,
		Nick:      nick,
		Cid:       cid,
		Price:     price,
		Promprice: price,
		ShopType:  "tmall.com",
		Imgs:      imgs,
		InStock:   isOnline,
		Attr:      attrs,
	}
	return &data, nil

}

func ParseWebFontTaobao(fonthtml string) (*Info, error) {
	cd, _ := iconv.Open("utf-8", "gbk")
	defer cd.Close()
	fonthtml = cd.ConvString(fonthtml)
	re := regexp.MustCompile(" cid:'(\\d+)")
	cats := re.FindStringSubmatch(fonthtml)
	if len(cats) == 0 {
		re = regexp.MustCompile("data-catid =\"(\\d+)")
		cats = re.FindStringSubmatch(fonthtml)
	}
	if len(cats) == 0 {
		err := errors.New("no categoryId")
		return nil, err
	}
	log.Info(cats)
	cid, _ := strconv.Atoi(cats[1])
	log.Info("cid", cid)
	re = regexp.MustCompile("sellerNick:\"(.+)\"")
	seller := re.FindStringSubmatch(fonthtml)
	if len(seller) == 0 {
		err := errors.New("seller nick is not exist")
		return nil, err
	}
	nick := seller[1]
	log.Info("nick ", nick)

	reader := strings.NewReader(fonthtml)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	ptag := doc.Find("div.tb-wrap-newshop ul li strong em.tb-rmb-num")
	if ptag.Length() == 0 {
		err = errors.New("No price")
		log.Error(err)
		return nil, err
	}
	pr := ptag.Text()
	re = regexp.MustCompile("\\d+\\.\\d+")
	ps := re.FindAllString(pr, 1)
	if len(ps) == 0 {
		err = errors.New("No price")
		return nil, err
	}
	price, _ := strconv.ParseFloat(ps[0], 64)
	log.Info("price ", price)
	fimg := doc.Find("img#J_ImgBooth")
	var imgs []string
	fimgurl, _ := fimg.Attr("data-src")
	re = regexp.MustCompile(IMG_POSTFIX)
	fimgurl = re.ReplaceAllString(fimgurl, "")
	log.Info(fimgurl)
	imgs = append(imgs, fimgurl)

	optimgs := doc.Find("ul#J_UlThumb li div a img")
	log.Info("optimgs size", optimgs.Length())
	optimgs.Each(func(i int, sq *goquery.Selection) {
		op, _ := sq.Attr("data-src")
		op = re.ReplaceAllString(op, "")
		log.Info("optimgs", op)
		imgs = append(imgs, op)
	})
	title := doc.Find("title").Text()
	rt := []rune(title)
	title = string(rt[0 : len(rt)-4])
	log.Info(title)

	isonline := doc.Find("div.tb-out-of-date div p.tb-hint strong")
	online := true
	if isonline.Length() > 0 {
		online = false
	}
	log.Info(isonline.Text())
	log.Info("is online ", online)
	attrs := make(map[string]string)
	doc.Find("div.attributes ul li").Each(func(i int, sq *goquery.Selection) {
		att := sq.Text()
		att = strings.TrimSpace(att)
		kv := strings.Split(att, ":")
		if len(kv) < 2 {
			return
		}
		attrs[kv[0]] = kv[1]
	})
	data := Info{
		Title:     title,
		Nick:      nick,
		Cid:       cid,
		Price:     price,
		Promprice: price,
		ShopType:  "taobao.com",
		Imgs:      imgs,
		InStock:   online,
		Attr:      attrs,
	}
	return &data, nil
}
