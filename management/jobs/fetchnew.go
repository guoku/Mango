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

func main() {
	var t int
	flag.IntVar(&t, "t", 1, "启动多少个线程,默认为1")
	flag.Parse()
	for {
		FetchTaobaoItem(t)
	}
}
func FetchTaobaoItem(threadnum int) {
	var mgopages *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "pages")
	var mgofailed *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "failed")
	var mgominer *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "minerals")
	var mgoMango *mgo.Collection = utils.MongoInit(MGOHOST, MANGO, "taobao_items_depot")
	var mgoShop *mgo.Collection = utils.MongoInit(MGOHOST, MANGO, "taobao_shops_depot")
	var shops []*crawler.ShopItem
	mgominer.Find(bson.M{"state": "posted"}).Sort("-date").Limit(10).All(&shops)
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
		istmall, err := crawler.IsTmall(items[0])
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
						err := crawler.Save(info, mgoMango)
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
		err = mgoShop.Update(bson.M{"shop_info.sid": sid}, bson.M{"$set": bson.M{"status": "finished"}})
		if err != nil {
			log.Error(err)
		}

	}
}

/*
import (
	"Mango/management/utils"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"strconv"
	"sync"
	"time"
)

const THREADSNUM int = 20
const (
	MGOHOST string = "10.0.1.23"
	MGODB   string = "zerg"
	TAOBAO  string = "taobao.com"
	TMALL   string = "tmall.com"
	MANGO   string = "mango"
)

func main() {

	log.Println("hello")
	shopitem := new(utils.ShopItem)
	minerals := utils.MongoInit(MGOHOST, MGODB, "minerals")
	iter := minerals.Find(bson.M{"state": "posted"}).Iter()
	for iter.Next(&shopitem) {
		shopid := strconv.Itoa(shopitem.Shop_id)
		items := shopitem.Items_list
		run(shopid, items)
	}
}

var mgopages *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "pages")
var mgofailed *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "failed")
var mgominer *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "minerals")
var mgoMango *mgo.Collection = utils.MongoInit(MGOHOST, MANGO, "taobao_items_depot")

func run(shopid string, items []string) {
	var allowchan chan bool = make(chan bool, THREADSNUM)

	log.Printf("\n\nStart to run fetch")
	shoptype := TAOBAO
	if len(items) == 0 {
		log.Println("%s has no items", shopid)
		return
	} else {
		istmall, err := utils.IsTmall(items[0])
		if err != nil {
			log.Println("there is an error during judge")
			log.Println(err.Error())
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
				log.Printf("start to fetch %s", itemid)
				page, err, detail := utils.Fetch(itemid, shoptype)
				if err != nil {
					log.Printf("%s failed", itemid)
					if err.Error() != "404" {

						failed := utils.FailedPages{ItemId: itemid, ShopId: shopid, ShopType: shoptype, UpdateTime: time.Now().Unix(), InStock: true}
						err = mgofailed.Insert(&failed)
						if err != nil {
							log.Println(err.Error())
							mgofailed.Update(bson.M{"itemid": itemid}, bson.M{"$set": failed})
						}
					}
				} else {

					log.Printf("%s 成功", itemid)
					info, missing, err := utils.Parse(page, detail, itemid, shopid, shoptype)
					log.Println("解析完毕")
					instock := true
					parsed := false
					if err != nil {
						log.Println(err.Error())
						if missing {
							parsed = true
							instock = false
						} else if err.Error() != "聚划算" {
							//聚划算数据不予以保存
							parsed = false
							if err.Error() == "cattag" {
								//有可能该商品找不到了
								instock = false
							} else {
								failed := utils.FailedPages{ItemId: itemid, ShopId: shopid, ShopType: shoptype, UpdateTime: time.Now().Unix(), InStock: true}
								err = mgofailed.Insert(&failed)
								if err != nil {
									log.Println(err.Error())
									mgofailed.Update(bson.M{"itemid": itemid}, bson.M{"$set": failed})
								}
								return
							}
						}
					} else {
						instock = info.InStock
						log.Println("开始发送")
						//err = utils.Post(info)
						err = utils.Save(info, mgoMango)
						if err != nil {
							log.Println("发送出现错误")
							log.Println(err.Error())
							parsed = false

						} else {
							log.Println("发送完毕")
							parsed = true
						}
					}
					successpage := utils.Pages{ItemId: itemid, ShopId: shopid, ShopType: shoptype, FontPage: page, UpdateTime: time.Now().Unix(), DetailPage: detail, Parsed: parsed, InStock: instock}
					err = mgopages.Insert(&successpage)
					if err != nil {
						log.Println(err.Error())
						mgopages.Update(bson.M{"itemid": itemid}, bson.M{"$set": successpage})
					}
				}
			}(itemid)
		}
		wg.Wait()
	}
	close(allowchan)
	sid, _ := strconv.Atoi(shopid)
	err := mgominer.Update(bson.M{"shop_id": sid}, bson.M{"$set": bson.M{"state": "fetched", "date": time.Now()}})
	if err != nil {
		log.Println("update minerals state error")
		log.Println(err.Error())
	}
}

*/
