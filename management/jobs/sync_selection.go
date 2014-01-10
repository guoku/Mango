package main

import (
	"Mango/management/crawler"
	"Mango/management/models"
	"Mango/management/utils"
	"encoding/json"
	"fmt"
	"github.com/qiniu/log"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
	"time"
)

var mgopages *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "pages")
var mgofailed *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "failed")
var mgoMango *mgo.Collection = utils.MongoInit(MGOHOST, MANGO, "taobao_items_depot")
var mgoShop *mgo.Collection = utils.MongoInit(MGOHOST, MANGO, "taobao_shops_depot")

const (
	MGOHOST string = "10.0.1.23"
	MGODB   string = "zerg"
	TAOBAO  string = "taobao.com"
	TMALL   string = "tmall.com"
	MANGO   string = "mango"
)

func main() {
	t := time.Now()
	t = t.Add(-24 * time.Hour)
	sync(t.Unix())
}

func sync(t int64) {
	//t是时间戳，表示提取的数据大于这个时间就可以了
	count := 100
	offset := 0
	for {
		link := fmt.Sprintf("http://114.113.154.47:8000/management/selection/sync?count=%d&offset=%d", count, offset)
		resp, err := http.Get(link)
		if err != nil {
			log.Error(err)
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			continue
		}
		entities := make([]Entity, 100)
		json.Unmarshal(body, &entities)
		if len(entities) == 0 {
			continue
		}

		log.Info("start to sync selection")
		for _, ent := range entities {
			log.Infof("%+v", ent)
			if ent.PostTime < float64(t) {
				return
			}
			process(ent)

		}
	}
}

func process(ent Entity) {
	for _, item := range ent.Items {
		itemid := item.Id
		nick := item.Nick
		//应该先查找这个店铺的数据，补全一些数据，然后进行抓取
		//如果没有这家店铺，则进行比较复杂的抓取，同时把店铺一同抓取了
		shop := new(models.ShopItem)
		err := mgoShop.Find(bson.M{"shop_info.nick": nick}).One(shop)
		if err != nil && err.Error() == mgo.ErrNotFound.Error() {
			fetch(itemid)
			continue
		}
		fetchWithShopid(itemid, shop.ShopInfo.Sid, shop.ShopInfo.ShopType)

	}
}

func fetch(itemid string) {
	font, detail, shoptype, instock, er := crawler.FetchWithOutType(itemid)
	if er != nil {
		crawler.SaveFailed(itemid, "", shoptype, mgofailed)
		log.Error(er)
		return
	}
	info, instock, err := crawler.ParsePage(font, detail, itemid, "", shoptype)
	if err != nil {
		crawler.SaveFailed(itemid, "", shoptype, mgofailed)
		log.Error(err)
		return
	}
	crawler.Save(info, mgoMango)
	sid := strconv.Itoa(info.Sid)
	crawler.SaveSuccessed(itemid, sid, shoptype, font, detail, true, instock, mgopages)
}

func fetchWithShopid(itemid string, shopid int, shoptype string) {
	font, detail, instock, err := crawler.FetchItem(itemid, shoptype)
	sid := strconv.Itoa(shopid)
	if err != nil {
		crawler.SaveFailed(itemid, sid, shoptype, mgofailed)
		log.Error(err)
		return
	}
	info, instock, err := crawler.ParsePage(font, detail, itemid, sid, shoptype)
	if err != nil {
		crawler.SaveFailed(itemid, sid, shoptype, mgofailed)
		log.Error(err)
		return
	}
	crawler.Save(info, mgoMango)
	crawler.SaveSuccessed(itemid, sid, shoptype, font, detail, true, instock, mgopages)
}

type Item struct {
	Id   string `json:"taobao_id"`
	Nick string `json:"shop_nick"`
}

type Entity struct {
	Id       int64   `json:"entity_id"`
	PostTime float64 `json:"post_time"`
	NoteId   int64   `json:"note_id"`
	Items    []Item  `json:"taobao_item_list"`
}
