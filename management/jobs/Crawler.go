package main

import (
	"Mango/management/crawler"
	"Mango/management/utils"
	"flag"
	"fmt"
	"github.com/qiniu/log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strconv"
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

var mgopages *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "pages")
var mgofailed *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "failed")
var mgominer *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "minerals")
var mgoMango *mgo.Collection = utils.MongoInit(MGOHOST, MANGO, "taobao_items_depot")

func main() {
	//按照state的更新时间来抓取数据
	var t int
	flag.IntVar(&t, "t", 1, "启动多少个线程,默认为1")
	flag.Parse()
	for {
		FetchTaobaoItem(t)
	}
}
func FetchTaobaoItem(threadnum int) {
	var shops []*utils.ShopItem
	mgominer.Find(bson.M{"state": "posted"}).Sort("date").Limit(10).All(&shops)
	log.Infof("t is %d", threadnum)
	for _, shopitem := range shops {

		var allowchan chan bool = make(chan bool, threadnum)
		log.Infof("\n\nStart to run fetch")
		shoptype := TAOBAO
		shopid := strconv.Itoa(shopitem.Shop_id)
		log.Infof("start to fetch shop %s", shopid)
		items := shopitem.Items_list
		if len(items) == 0 {
			return
		}
		istmall, err := utils.IsTmall(items[0])
		if err != nil {
			log.Error(err)
			return
		}
		if istmall {
			shoptype = TMALL
		}

		var wg sync.WaitGroup
		for _, itemid := range items {
			allowchan <- true
			wg.Add(1)
			go func(itemid string) {
				defer wg.Done()
				defer func() { <-allowchan }()
				font, detail, instock, err := crawler.FetchItem(itemid, shoptype)
				if err != nil {
					if instock {
						crawler.SaveFailed(itemid, shopid, shoptype, mgofailed)
					}
				} else {
					info, instock, err := crawler.ParsePage(font, detail, itemid, shopid, shoptype)
					if err != nil {
						if instock {
							crawler.SaveSuccessed(itemid, shopid, shoptype, font, detail, false, instock, mgopages)
						}
					} else {
						//保存解析结果到mongo
						err := utils.Save(info, mgoMango)
						fmt.Printf("%+v", info)
						parsed := false
						if err != nil {
							log.Error(err)
							parsed = false
						} else {
							parsed = true
						}
						crawler.SaveSuccessed(itemid, shopid, shoptype, font, detail, parsed, instock, mgopages)
					}
				}
			}(itemid)
		}
		wg.Wait()
		close(allowchan)
		sid, _ := strconv.Atoi(shopid)
		err = mgominer.Update(bson.M{"shop_id": sid}, bson.M{"$set": bson.M{"state": "fetched", "date": time.Now()}})
		if err != nil {
			log.Info("update minerals state error")
			log.Info(err.Error())

		}

	}
}
