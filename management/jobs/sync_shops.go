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
)

func main() {
	syncOnlineShops()
}
func syncOnlineShops() {
	count := 50
	offset := 4560
	var mgoShop *mgo.Collection = utils.MongoInit("10.0.1.23", "mango", "taobao_shops_depot")
	for {
		link := fmt.Sprintf("http://b.guoku.com/sync/shop?all=true&count=%d&offset=%d", count, offset)
		resp, err := http.Get(link)
		if err != nil {
			log.Error(err)
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
			return
		}
		for _, shop := range shops {
			//如果存在就直接覆盖，如果不存在，就直接插入
			sid := shop.ShopInfo.Sid
			log.Info(sid)
			log.Info(shop.ShopInfo.Title)
			change, err := mgoShop.Upsert(bson.M{"shop_info.sid": sid}, bson.M{"$set": shop})
			if err != nil {
				log.Error(err)
			}
			log.Info(change.Updated)
		}
		offset = offset + count
	}
}
