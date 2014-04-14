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
        fetchfailed()
    }
}

func fetchfailed() {
    tnum, _ := beego.AppConfig.Int("fetchfailed::thread")
    var allowchan chan bool = make(chan bool, tnum)

    mgopages := MongoInit(MGOHOST, ZERG, PAGES)
    mgofailed := MongoInit(MGOHOST, ZERG, FAILED)
    itemDepot := MongoInit(MGOHOST, MANGO, ITEMS_DEPOT)
    iter := mgofailed.Find(nil).Iter()
    shopDepot := MongoInit(MGOHOST, MANGO, SHOPS_DEPOT)
    failed := new(crawler.FailedPages)
    for iter.Next(&failed) {
        allowchan <- true
        go fetchFailedItem(failed, mgofailed, mgopages, itemDepot, shopDepot, allowchan)
    }
}

func fetchFailedItem(failed *crawler.FailedPages, mgofailed, mgopages, itemDepot, shopDepot *mgo.Collection, allowchan chan bool) {
    defer func() { <-allowchan }()
    _, err := mgofailed.RemoveAll(bson.M{"itemid": failed.ItemId})
    if err != nil {
        log.Error(err)
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
        if isWeb {
            info, err := crawler.ParseWeb(page, detail, failed.ItemId, failed.ShopId, failed.ShopType)
            if err != nil {
                log.Error(err)
                return
            }
            fetchShop(info.Sid, shopDepot)
            err = crawler.Save(info, itemDepot)
            SAdd("jobs:fetchfailed", fmt.Sprintf("%s", info.Sid))
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
            fetchShop(info.Sid, shopDepot)
            err = crawler.Save(info, itemDepot)
            if err != nil {
                log.Error(err)
                return
            }
            SAdd("jobs:fetchfailed", fmt.Sprintf("%s", info.Sid))
            crawler.SaveSuccessed(failed.ItemId, failed.ShopId, failed.ShopType, page, detail, true, instock, mgopages)
        }
    }
}

func fetchShop(sid int, mgoShop *mgo.Collection) {
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
