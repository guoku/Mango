package main

import (
    "Mango/management/crawler"
    "Mango/management/utils"
    "flag"
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
    log.SetOutputLevel(log.Lerror)
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
    log.Info("shop length ", len(shops))
    for _, shopitem := range shops {

        var allowchan chan bool = make(chan bool, threadnum)
        log.Infof("\n\nStart to run fetch")
        shoptype := TAOBAO
        shopid := strconv.Itoa(shopitem.Shop_id)

        log.Errorf("start to fetch shop %s", shopid)
        items := shopitem.Items_list
        if len(items) == 0 {
            log.Info("itemid is none")
            err := mgominer.Update(bson.M{"shop_id": shopitem.Shop_id}, bson.M{"$set": bson.M{"state": "fetched", "date": time.Now()}})
            if err != nil {
                log.Errorf("更新店铺失败,shopid %d", shopitem.Shop_id)
                log.Error(err)
            }
            continue
        }
        istmall, err := crawler.IsTmall(items[0])
        if err != nil {
            log.Errorf("判断是否为天猫商品失败，itemid %s", items[0])
            log.Error(err)
            continue
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
                font, detail, instock, err, isWeb := crawler.FetchItem(itemid, shoptype)
                if err != nil {
                    log.Error("抓取页面失败,itemid ", itemid)
                    log.Error(err)
                    if instock {
                        crawler.SaveFailed(itemid, shopid, shoptype, mgofailed)
                    }
                } else {
                    if isWeb {
                        info, err := crawler.ParseWeb(font, detail, itemid, shopid, shoptype)
                        if err != nil {
                            log.Errorf("解析web页面失败，itemid %s", itemid)
                            log.Error(err)
                            crawler.SaveFailed(itemid, shopid, shoptype, mgofailed)
                        }
                        err = crawler.Save(info, mgoMango)
                        if err != nil {
                            log.Errorf("保存解析结果失败，itemid %s", itemid)
                            log.Error(err)
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
                            parsed := false
                            if err != nil {
                                log.Error(err)
                                log.Error(itemid)
                                parsed = false
                            } else {
                                parsed = true
                            }
                            crawler.SaveSuccessed(itemid, shopid, shoptype, font, detail, parsed, instock, mgopages)
                        }
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
