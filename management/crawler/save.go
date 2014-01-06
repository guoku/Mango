package crawler

import (
	"Mango/management/models"
	"encoding/json"
	"github.com/qiniu/log"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
	"time"
)

//把解析后的商品数据存储到mongo里
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
		change["uploaded"] = false

	} else {
		change["data_last_revised_time"] = time.Now()
		change["refreshed"] = true //这个字段表明该商品之前已经爬取了，现在是更新数据,需要refresh

		change["refresh_time"] = time.Now()
	}
	err = mgocol.Update(bson.M{"num_iid": int(item.ItemId)}, bson.M{"$set": change})
	if err != nil {
		return err
	}
	log.Info("解析数据保存成功")
	return nil
}

//把解析好的商品数据发送到指定API去
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

func SaveFailed(itemid, shopid, shoptype string, mgofailed *mgo.Collection) {
	failed := FailedPages{ItemId: itemid, ShopId: shopid, ShopType: shoptype, UpdateTime: time.Now().Unix(), InStock: true}
	_, err := mgofailed.Upsert(bson.M{"itemid": itemid}, bson.M{"$set": failed})
	if err != nil {
		log.Error(err)
	}
}

func SaveSuccessed(itemid, shopid, shoptype, font, detail string, parsed, instock bool, mgopages *mgo.Collection) {
	font = Compress(font)
	//log.Info("压缩后的字符", font)
	detail = Compress(detail)
	successpage := Pages{ItemId: itemid, ShopId: shopid, ShopType: shoptype, FontPage: font, UpdateTime: time.Now().Unix(), DetailPage: detail, Parsed: parsed, InStock: instock}
	_, err := mgopages.Upsert(bson.M{"itemid": itemid}, bson.M{"$set": successpage})
	if err != nil {
		log.Error(err)
	}
	log.Info("保存页面数据成功")
}
