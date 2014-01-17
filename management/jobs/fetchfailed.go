package main

import (
	"Mango/management/crawler"
	"Mango/management/models"
	"Mango/management/utils"
	"flag"
	"fmt"
	"github.com/qiniu/log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"sync"
	"time"
)

const (
	MGOHOST string = "10.0.1.23"
	MGODB   string = "zerg"
	TAOBAO  string = "taobao.com"
	TMALL   string = "tmall.com"
	MANGO   string = "mango"
)

var mgoShop *mgo.Collection = utils.MongoInit(MGOHOST, MANGO, "taobao_shops_depot")

func main() {
	log.SetOutputLevel(log.Lerror)
	var t int
	flag.IntVar(&t, "t", 1, "启动线程的数量，默认为1")
	flag.Parse()
	for {
		run(t)
	}
}

func run(t int) {
	var allowchan chan bool = make(chan bool, t)
	mgopages := utils.MongoInit(MGOHOST, MGODB, "pages")
	mgofailed := utils.MongoInit(MGOHOST, MGODB, "failed")
	mgoMango := utils.MongoInit(MGOHOST, MANGO, "taobao_items_depot")
	log.Info("start to refetch")
	iter := mgofailed.Find(nil).Iter()
	failed := new(crawler.FailedPages)
	var wg sync.WaitGroup
	for iter.Next(&failed) {
		allowchan <- true
		wg.Add(1)
		go func(failed *crawler.FailedPages) {
			defer wg.Done()
			defer func() { <-allowchan }()
			info, err := mgofailed.RemoveAll(bson.M{"itemid": failed.ItemId})
			if err != nil {
				log.Info(info.Removed)
				log.Info(err.Error())
			}
			page, detail, instock, err, isWeb := crawler.FetchItem(failed.ItemId, failed.ShopType)
			if err != nil {
				log.Error(err)

				if instock {
					crawler.SaveFailed(failed.ItemId, failed.ShopId, failed.ShopType, mgofailed)
				} else {
					mgofailed.RemoveAll(bson.M{"itemid": failed.ItemId})
				}
			} else {
				log.Info("%s refetch successed", failed.ItemId)
				if isWeb {
					info, err := crawler.ParseWeb(page, detail, failed.ItemId, failed.ShopId, failed.ShopType)
					if err != nil {
						log.Error(err)
						return
					}
					fetchShop(info.Sid)
					err = crawler.Save(info, mgoMango)
					if err != nil {
						log.Error(err)
						return
					}
				} else {
					info, instock, err := crawler.ParsePage(page, detail, failed.ItemId, failed.ShopId, failed.ShopType)

					if err != nil {
						log.Error(err)
						return
					}
					instock = info.InStock
					fetchShop(info.Sid)
					err = crawler.Save(info, mgoMango)
					if err != nil {
						log.Error(err)
						return
					}
					crawler.SaveSuccessed(failed.ItemId, failed.ShopId, failed.ShopType, page, detail, true, instock, mgopages)
				}
			}
		}(failed)
	}
	wg.Wait()
	close(allowchan)
	if err := iter.Close(); err != nil {
		log.Info(err.Error())
	}
}
func fetchShop(sid int) {
	sp := new(models.ShopItem)
	//如果店铺不存在，就抓取存入
	mgoShop.Find(bson.M{"shop_info.sid": sid}).One(&sp)
	if sp.ShopInfo == nil {
		log.Info("开始爬取店铺")
		shoplink := fmt.Sprintf("http://shop%d.taobao.com", sid)
		shopinfo, err := crawler.FetchShopDetail(shoplink)
		if err != nil {
			log.Error(err)
			return
		}
		log.Infof("%+v", shopinfo)
		shop := models.ShopItem{}
		shop.ShopInfo = shopinfo
		shop.CreatedTime = time.Now()
		shop.LastUpdatedTime = time.Now()
		shop.LastCrawledTime = time.Now()
		shop.Status = "queued"
		shop.CrawlerInfo = &models.CrawlerInfo{Priority: 10, Cycle: 720}
		shop.ExtendedInfo = &models.TaobaoShopExtendedInfo{Type: shopinfo.ShopType, Orientational: false, CommissionRate: -1}
		mgoShop.Insert(shop)
	}
}
