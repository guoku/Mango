package main

import (
	"Mango/management/utils"
	"github.com/qiniu/log"
	"labix.org/v2/mgo/bson"
	"time"
)

const (
	MGOHOST string = "10.0.1.23"
	MGODB   string = "zerg"
	TAOBAO  string = "taobao.com"
	TMALL   string = "tmall.com"
)

func main() {
	for {
		run()
	}
}

func run() {
	log.SetOutputLevel(log.Ldebug)
	mgopages := utils.MongoInit(MGOHOST, MGODB, "pages")
	mgofailed := utils.MongoInit(MGOHOST, MGODB, "failed")
	log.Info("start to refetch")
	iter := mgofailed.Find(nil).Iter()
	failed := new(utils.FailedPages)
	for iter.Next(&failed) {
		info, err := mgofailed.RemoveAll(bson.M{"itemid": failed.ItemId})
		if err != nil {
			log.Info(info.Removed)
			log.Info(err.Error())
		}
		page, err, detail := utils.Fetch(failed.ItemId, failed.ShopType)
		if err != nil {
			if err.Error() == "404" {
				log.Info("start to remove item")
				_, err = mgofailed.RemoveAll(bson.M{"itemid": failed.ItemId})
				if err != nil {
					log.Info(err.Error())
				}
			} else {
				log.Infof("%s refetch failed\n", failed.ItemId)
				newfail := utils.FailedPages{ItemId: failed.ItemId, ShopId: failed.ShopId, ShopType: failed.ShopType, UpdateTime: time.Now().Unix(), InStock: failed.InStock}
				err = mgofailed.Insert(&newfail)
				if err != nil {
					log.Info(err.Error())
					mgofailed.Update(bson.M{"itemid": failed.ItemId}, bson.M{"$set": newfail})
				}
			}
		} else {
			log.Info("%s refetch successed", failed.ItemId)
			info, missing, err := utils.Parse(page, detail, failed.ItemId, failed.ShopId, failed.ShopType)
			instock := true
			parsed := false
			if err != nil {
				if missing {
					parsed = true
					instock = false
				} else {
					parsed = false
					if err.Error() == "cattag" {
						instock = false
					}

				}
				log.Info(err.Error())
			} else {
				instock = info.InStock
				err = utils.Post(info)
				if err != nil {
					log.Info(err.Error())

				}
			}
			successpage := utils.Pages{ItemId: failed.ItemId, ShopId: failed.ShopId, ShopType: failed.ShopType, FontPage: page, DetailPage: detail, UpdateTime: time.Now().Unix(), Parsed: parsed, InStock: instock}
			err = mgopages.Insert(&successpage)
			if err != nil {
				log.Info(err.Error())
				mgopages.Update(bson.M{"itemid": failed.ItemId}, bson.M{"$set": successpage})
			}
		}
	}
	if err := iter.Close(); err != nil {
		log.Info(err.Error())
	}
}
