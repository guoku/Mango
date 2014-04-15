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
    session, err := mgo.Dial(MGOHOST)
    if err != nil {
        log.ErrorfType("mongo err", "%s", err.Error())
        return
    }
    mgopages := session.DB(ZERG).C(PAGES)
    mgofailed := session.DB(ZERG).C(FAILED)
    itemDepot := session.DB(MANGO).C(ITEMS_DEPOT)
    shopDepot := session.DB(MANGO).C(SHOPS_DEPOT)
    iter := mgofailed.Find(nil).Iter()
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
        return
    }
    for i := 0; i < 10; i++ {
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
                    continue
                }
                fetchShop(info.Sid, shopDepot)
                err = crawler.Save(info, itemDepot)
                SAdd("jobs:fetchfailed", fmt.Sprintf("%s", info.Sid))
                if err != nil {
                    log.Error(err)
                    continue
                }
            } else {
                info, instock, err := crawler.ParsePage(page, detail, failed.ItemId, failed.ShopId, failed.ShopType)
                if err != nil {
                    log.Error(err)
                    continue
                }
                instock = info.InStock
                fetchShop(info.Sid, shopDepot)
                err = crawler.Save(info, itemDepot)
                if err != nil {
                    log.Error(err)
                    continue
                }
                SAdd("jobs:fetchfailed", fmt.Sprintf("%s", info.Sid))
                crawler.SaveSuccessed(failed.ItemId, failed.ShopId, failed.ShopType, page, detail, true, instock, mgopages)
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
