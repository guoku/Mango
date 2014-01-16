package main

import (
	"Mango/management/crawler"
	"Mango/management/utils"
	"flag"
	"github.com/qiniu/log"
	"labix.org/v2/mgo/bson"
	"sync"
)

const (
	MGOHOST string = "10.0.1.23"
	MGODB   string = "zerg"
	TAOBAO  string = "taobao.com"
	TMALL   string = "tmall.com"
	MANGO   string = "mango"
)

func main() {
	var t int
	flag.IntVar(&t, "t", 1, "启动线程的数量，默认为1")
	flag.Parse()
	for {
		run(t)
	}
}

func run(t int) {
	var allowchan chan bool = make(chan bool, t)
	log.SetOutputLevel(log.Ldebug)
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
