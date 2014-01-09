package main

import (
	"Mango/management/crawler"
	"Mango/management/models"
	"Mango/management/utils"
	"fmt"
	"github.com/qiniu/log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

func main() {
	for {
		refresh()
		time.Sleep(2 * time.Minute)

	}
}

func refresh() {
	var mgoShop *mgo.Collection = utils.MongoInit("10.0.1.23", "mango", "taobao_shops_depot")
	shops := make([]models.ShopItem, 0)
	err := mgoShop.Find(bson.M{"shop_info.seller_id": 0}).Limit(50).All(&shops)
	if err != nil {
		log.Error(err)
		return
	}

	for _, shop := range shops {
		sid := shop.ShopInfo.Sid
		log.Info(sid)
		cid := shop.ShopInfo.Cid
		nick := shop.ShopInfo.Nick
		title := shop.ShopInfo.Title
		log.Info(title)
		shoplink := fmt.Sprintf("http://shop%d.taobao.com", sid)
		shopinfo, err := crawler.FetchShopDetail(shoplink)
		if err != nil {
			log.Error(err)
			if err.Error() == "the shop is no longer exist" {
				mgoShop.Remove(bson.M{"shop_info.sid": sid})
			}
			continue
		}
		shopinfo.Nick = nick
		shopinfo.Title = title
		shopinfo.Cid = cid
		if shop.ExtendedInfo.Type == "global" {
			shopinfo.ShopType = "global"
		} else {
			shop.ExtendedInfo.Type = shopinfo.ShopType
		}
		shop.ShopInfo = shopinfo
		mgoShop.Update(bson.M{"shop_info.sid": sid}, bson.M{"$set": shop})
	}
}
