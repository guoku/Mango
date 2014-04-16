package jobs

import (
    "Mango/gojobs/crawler"
    "Mango/gojobs/log"
    "Mango/gojobs/models"
    "fmt"
    "github.com/astaxie/beego"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
    "time"
)

type Fetchfailed struct {
    Base
}

func (this *Fetchfailed) run() {
    defer func() {
        this.start = false
    }()

    for {
        if this.start == false {
            return
        }
        this.fetchfailed()

    }
}

func (this *Fetchfailed) fetchfailed() {
    tnum, _ := beego.AppConfig.Int("fetchfailed::thread")
    var allowchan chan bool = make(chan bool, tnum)
    session, err := mgo.Dial(MGOHOST)
    defer func() {
        if session != nil {
            session.Close()
        }
    }()
    if err != nil {
        log.ErrorfType("mongo err", "%s", err.Error())
        return
    }
    fmt.Println("连接成功")
    mgopages := session.DB(ZERG).C(PAGES)
    mgofailed := session.DB(ZERG).C(FAILED)
    itemDepot := session.DB(MANGO).C(ITEMS_DEPOT)
    shopDepot := session.DB(MANGO).C(SHOPS_DEPOT)
    iter := mgofailed.Find(nil).Iter()
    failed := new(crawler.FailedPages)
    i := 0
    for iter.Next(&failed) {
        i = i + 1
        if this.start == false {
            break
        }
        allowchan <- true
        fmt.Println("发出", failed.ItemId)
        go fetchFailedItem(failed, mgofailed, mgopages, itemDepot, shopDepot, allowchan)
        fmt.Println("发出完成")
    }
    if i == 0 {
        //说明已经没有数据了
        time.Sleep(60 * time.Second)
    }
    fmt.Println("等待")
    for {
        if len(allowchan) > 0 {
            time.Sleep(1 * time.Second)
        } else {
            break
        }
    }
}

func fetchFailedItem(
    failed *crawler.FailedPages,
    mgofailed, mgopages, itemDepot, shopDepot *mgo.Collection,
    allowchan chan bool) {
    success := false
    defer func() {
        if success {
            SAdd("jobs:fetchfailed", failed.ItemId)
            mgofailed.RemoveAll(bson.M{"itemid": failed.ItemId})
        }
        <-allowchan
        fmt.Println("channel size ", len(allowchan))
        fmt.Println("完成", failed.ItemId)
    }()
    for i := 0; i < 10; i++ {
        if failed.ItemId == "" {
            success = true
            break
        }
        page, detail, instock, err, isWeb := crawler.FetchItem(failed.ItemId, failed.ShopType)
        if err != nil {
            //log.Error(err)
            if instock {
                if i == 9 {
                    //crawler.SaveFailed(failed.ItemId, failed.ShopId, failed.ShopType, mgofailed)
                    break
                } else {
                    continue
                }
            } else {
                if i == 9 {
                    //连续抓了9次返回的都是404错误，应该可以说明这个商品下架了
                    success = true
                    break
                }
                continue
            }
        } else {
            if isWeb {
                info, err := crawler.ParseWeb(page, detail, failed.ItemId, failed.ShopId, failed.ShopType)
                if err != nil {
                    log.Error(err)
                    break
                }
                fetchShop(info.Sid, shopDepot)
                err = crawler.Save(info, itemDepot)
            } else {
                info, instock, err := crawler.ParsePage(page, detail, failed.ItemId, failed.ShopId, failed.ShopType)
                if err != nil {
                    log.Error(err)
                    break
                }
                instock = info.InStock
                fetchShop(info.Sid, shopDepot)
                err = crawler.Save(info, itemDepot)
                if err != nil {
                    log.Error(err)
                    break
                }
                crawler.SaveSuccessed(failed.ItemId, failed.ShopId, failed.ShopType, page, detail, true, instock, mgopages)
                success = true
            }
        }
        break //执行到这里已经没有错了

    }
}

func fetchShop(sid int, mgoShop *mgo.Collection) {
    sp := new(models.ShopItem)
    //如果店铺不存在，就抓取存入
    mgoShop.Find(bson.M{"shop_info.sid": sid}).One(&sp)
    if sp.ShopInfo == nil {
        for i := 0; i < 10; i++ {
            log.Info("开始爬取店铺")
            shoplink := fmt.Sprintf("http://shop%d.taobao.com", sid)
            shopinfo, err := crawler.FetchShopDetail(shoplink)
            if err != nil {
                log.Error(err)
                continue
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
            break
        }
    }
}
