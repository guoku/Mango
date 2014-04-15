package jobs

import (
    "Mango/gojobs/crawler"
    "Mango/gojobs/log"
    "Mango/gojobs/models"
    "fmt"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
    "time"
)

type StatuUpdate struct {
    Base
}

func (this *StatuUpdate) run() {
    defer func() {
        this.start = false
    }()

    for {
        if this.start == false {
            fmt.Println("停止了")
            return
        }
        update()
        updateShopItems()
        fmt.Println("执行一次")
        time.Sleep(1 * time.Hour)
    }
}

//把更新时隔超过指定期限的店铺变为queued状态
func update() {

    session, err := mgo.Dial(MGOHOST)
    if err != nil {
        log.ErrorfType("mongo err", "%s", err.Error())
        return
    }
    c := session.DB(MANGO).C(SHOPS_DEPOT)
    shops := make([]models.ShopItem, 200)
    c.Find(bson.M{"status": "finished"}).All(&shops)
    for _, shop := range shops {
        lastupdatetime := shop.LastCrawledTime
        now := time.Now()
        diff := now.Sub(lastupdatetime)
        cycle := shop.CrawlerInfo.Cycle
        if diff.Hours() > float64(cycle) {
            c.Update(bson.M{"shop_info.sid": shop.ShopInfo.Sid}, bson.M{"$set": bson.M{"status": "queued"}})
        }
    }
}

type MineralShop struct {
    Sid       int       `bson:"shop_id"`
    Date      time.Time `bson:"date"`
    ItemsList []string  `bson:"items_list"`
    ItemNum   int       `bson:"item_num"`
    State     string    `bson:"state"`
    ShopType  string    `bson:"shop_type"`
}

//把处于queued状态的店铺的商品列表爬取下来，放到zerg.minerals里面去
func updateShopItems() {
    mgoShop := MongoInit(MGOHOST, MANGO, SHOPS_DEPOT)
    mgoMinerals := MongoInit(MGOHOST, ZERG, MINERALS)
    shops := make([]models.ShopItem, 0)
    err := mgoShop.Find(bson.M{"status": "queued"}).Limit(10).All(&shops)
    if err != nil {
        log.ErrorfType("mongo err", "%s", err.Error())
        return
    }

    if len(shops) == 0 {
        fmt.Println("no shop to update")
        return
    }
    for _, shop := range shops {
        sid := shop.ShopInfo.Sid
        shoplink := shop.ShopInfo.ShopLink
        if shoplink == "" {
            shoplink = fmt.Sprintf("http://shop%d.taobao.com", sid)
        }
        items, err := crawler.GetShopItems(shoplink)
        if err != nil {
            log.ErrorfType("fetch shop err", "%d %s", sid, err.Error())
            if err.Error() == "超链接提取出错" {
                //店铺已经不存在了
                err = mgoShop.Remove(bson.M{"shop_info.sid": sid})
            }
            continue
        }

        sits := MineralShop{
            Sid:       sid,
            Date:      time.Now(),
            State:     "posted",
            ItemNum:   len(items),
            ItemsList: items,
            ShopType:  shop.ShopInfo.ShopType,
        }

        _, err = mgoMinerals.Upsert(bson.M{"shop_id": sid}, bson.M{"$set": sits})
        if err != nil {
            log.ErrorfType("mongo err", "%s", err.Error())
            continue
        }
        err = mgoShop.Update(bson.M{"shop_info.sid": sid}, bson.M{"$set": bson.M{"status": "crawling"}})
        if err != nil {
            fmt.Println(err)
        }

        SAdd("jobs:statuupdate", fmt.Sprintf("%d", shop.ShopInfo.Sid))
        fmt.Println("更新完一家", sid)
    }
}
