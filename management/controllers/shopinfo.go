package controllers

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/jason-zou/taobaosdk/rest"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func fetch(link string) (*rest.Shop, error) {
	doc, e := goquery.NewDocument(link)
	if e != nil {
		log.Println(e.Error())
		return nil, e
	}

	titletag := doc.Find("p.box").Text()
	title := titletag[7:]
	title = strings.TrimSpace(title)
	log.Println(title)
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
	detail := wwimg.Parent().Parent().Text()
	re := regexp.MustCompile("\\d\\.\\d分")
	scores := re.FindAllString(detail, -1)
	var item_score, service_score, delivery_score float32
	item_score2, _ := strconv.ParseFloat(scores[0][0:len(scores[0])-3], 32)
	service_score2, _ := strconv.ParseFloat(scores[1][0:len(scores[0])-3], 32)
	delivery_score2, _ := strconv.ParseFloat(scores[2][0:len(scores[0])-3], 32)
	service_score = float32(service_score2)
	item_score = float32(item_score2)
	delivery_score = float32(delivery_score2)
	var shopscore *rest.ShopScore
	deliver := &rest.RateScore{Score: delivery_score}
	desc := &rest.RateScore{Score: item_score}
	service := &rest.RateScore{Score: service_score}
	kps := &rest.ShopKPS{DeliveryScore: deliver, DescScore: desc, ServiceScore: service}
	shopscore.SemiScore = kps
	shopinfo := &rest.Shop{Nick: nick, Title: title, PicPath: piclink, ShopScore: shopscore, Sid: sid}
	return shopinfo, nil
}
