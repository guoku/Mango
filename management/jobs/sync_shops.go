package main

import (
	"Mango/management/models"
	"Mango/management/utils"
	"encoding/json"
	"fmt"
	"github.com/qiniu/log"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"time"
)

type Uptime struct {
	Last time.Time `bson:"last"`
	Name string    `bson:"name"`
}

func main() {
	syncOnlineShops()
	time.Sleep(time.Hour * 24)
}
func syncOnlineShops() {
	count := 50
	offset := 0
	var mgoShop *mgo.Collection = utils.MongoInit("10.0.1.23", "mango", "taobao_shops_depot")
	var mgoTime *mgo.Collection = utils.MongoInit("10.0.1.23", "zerg", "time")
	utime := new(Uptime)
	mgoTime.Find(bson.M{"name": "last"}).One(&utime)
	date := utime.Last.Format("2006010203")
	log.Info(date)
	for {
		link := fmt.Sprintf("http://b.guoku.com/sync/shop?count=%d&offset=%d&date=%s", count, offset, date)
		resp, err := http.Get(link)
		if err != nil {
			log.Error(err)
			return
		}
		if resp.StatusCode != 200 {
			log.Error(resp.Status)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			return
		}
		shops := make([]models.ShopItem, 30)
		json.Unmarshal(body, &shops)
		if len(shops) == 0 {
			log.Info("all shops are synced")
			mgoTime.Update(bson.M{"name": "last"}, bson.M{"$set": bson.M{"last": time.Now()}})
			return
		}
		for _, shop := range shops {
			if shop.ShopInfo == nil {
				continue
			}
			sid := shop.ShopInfo.Sid
			log.Info(sid)
			log.Info(shop.ShopInfo.Title)
			sp := new(models.ShopItem)
			mgoShop.Find(bson.M{"shop_info.sid": sid}).One(&sp)
			if sp.ShopInfo == nil {
				//说明这家店铺是新添加的
				log.Info("新添加的店铺")
				mgoShop.Insert(shop)
			} else {
				log.Info("更新的店铺")
				shoptype := shop.ExtendedInfo.Type
				mgoShop.Update(bson.M{"shop_info.sid": sid}, bson.M{"$set": bson.M{"extended_info": shop.ExtendedInfo, "crawler_info": shop.CrawlerInfo, "shop_info.shop_type": shoptype}})
			}

		}
		offset = offset + count
	}
}
