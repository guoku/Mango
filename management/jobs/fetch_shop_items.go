package main

import (
	"Mango/management/crawler"
	"Mango/management/models"
	"Mango/management/utils"
	"fmt"
	"github.com/qiniu/log"
	"labix.org/v2/mgo/bson"
	"time"
)

func main() {
	for {
		GetItems()
		time.Sleep(2 * time.Minute)
	}
}

type Shopitems struct {
	Sid       int       `bson:"shop_id"`
	Date      time.Time `bson:"date"`
	ItemsList []string  `bson:"items_list"`
	ItemNum   int       `bson:"item_num"`
	State     string    `bson:"state"`
	ShopType  string    `bson:"shop_type"`
}

func GetItems() {
	mgoShop := utils.MongoInit("10.0.1.23", "mango", "taobao_shops_depot")
	minerals := utils.MongoInit("10.0.1.23", "zerg", "minerals")
	shops := make([]models.ShopItem, 0)
	err := mgoShop.Find(bson.M{"status": "crawling"}).Limit(10).All(&shops)
	if err != nil {
		log.Error(err)
		return
	}
	for _, shop := range shops {
		log.Infof("%+v", shop)
		sid := shop.ShopInfo.Sid
		shoplink := shop.ShopInfo.ShopLink
		if shoplink == "" {
			shoplink = fmt.Sprintf("http://shop%d.taobao.com", sid)
		}
		items, err := crawler.GetShopItems(shoplink)
		if err != nil {
			log.Infof("抓取店铺 %d 的时候出现错误", sid)
			log.Error(err)
			continue
		}
		sits := Shopitems{Sid: sid, Date: time.Now(), State: "posted", ItemNum: len(items), ItemsList: items, ShopType: shop.ShopInfo.ShopType}
		minerals.Upsert(bson.M{"shop_id": sid}, bson.M{"$set": sits})
		mgoShop.Update(bson.M{"shop_info.sid": sid}, bson.M{"$set": bson.M{"status": "crawling"}})
	}
}