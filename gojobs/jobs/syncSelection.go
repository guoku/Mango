package jobs

import (
    "Mango/gojobs/crawler"
    "Mango/gojobs/log"
    "Mango/gojobs/models"
    "encoding/json"
    "fmt"
    "github.com/astaxie/beego"
    "io/ioutil"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
    "net/http"
    "strconv"
    "time"
)

type SyncSelection struct {
    Base
}

func (this *SyncSelection) run() {
    defer func() {
        this.start = false
    }()

    for {
        if this.start == false {
            return
        }
        t := time.Now()
        t = t.Add(-time.Hour * 24 * 30)
        syncSelection(t.Unix(), 0)
        time.Sleep(1 * time.Hour)
    }
}

func syncSelection(t int64, offset int) {
    count := 200
    session, err := mgo.Dial(MGOHOST)
    if err != nil {
        log.ErrorfType("mongo err", "%s", err.Error())
        return
    }

    mgofailed := session.DB(ZERG).C(FAILED)
    mgopages := session.DB(ZERG).C(PAGES)
    shopDepot := session.DB(MANGO).C(SHOPS_DEPOT)
    itemDepot := session.DB(MANGO).C(ITEMS_DEPOT)

    syncLink := beego.AppConfig.String("syncselection::link")

    for {
        link := fmt.Sprintf(syncLink+"?count=%d&offset=%d", count, offset)
        offset = offset + count
        fmt.Println("\n\n", offset)
        resp, err := http.Get(link)
        if err != nil {
            log.ErrorfType("server err", "%s %s", "服务器错误", err.Error())
            return
        }
        if resp.StatusCode != 200 {
            log.ErrorfType("server err", "%s", "服务器错误")
            return
        }
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            log.Error(err)
            return
        }

        entities := make([]Entity, 100)
        json.Unmarshal(body, &entities)
        if len(entities) == 0 {
            resp.Body.Close()
        }

        allowchan := make(chan bool, 5)
        for _, ent := range entities {
            if ent.PostTime < float64(t) {
                continue
            }
            allowchan <- true
            go s_process(ent, allowchan, mgofailed, mgopages, itemDepot, shopDepot)
        }
        resp.Body.Close()
    }

}

func s_process(ent Entity, allowchan chan bool, mgofailed, mgopages, itemDepot, shopDepot *mgo.Collection) {
    for _, item := range ent.Items {
        itemid := item.Id
        nick := item.Nick
        shop := new(models.ShopItem)
        err := shopDepot.Find(bson.M{"shop_info.nick": nick}).One(shop)
        if err != nil && err.Error() == mgo.ErrNotFound.Error() {
            SAdd("jobs:syncselection", itemid)
            s_fetch(itemid, mgofailed, mgopages, itemDepot, shopDepot)
            continue
        }
        if shop.ShopInfo == nil {
            s_fetch(itemid, mgofailed, mgopages, itemDepot, shopDepot)
            SAdd("jobs:syncselection", itemid)
            continue
        }

        fetchWithShopid(itemid, shop.ShopInfo.Sid, shop.ShopInfo.ShopType, mgofailed, mgopages, itemDepot)
        SAdd("jobs:syncselection", itemid)
    }
    defer func() {
        <-allowchan
    }()
}

func s_fetch(itemid string, mgofailed, mgopages, itemDepot, shopDepot *mgo.Collection) {
    for i := 0; i < 10; i++ {
        font, detail, shoptype, instock, err := crawler.FetchWithOutType(itemid)
        if err != nil {
            crawler.SaveFailed(itemid, "", shoptype, mgofailed)
            log.Error(err)
            continue
        }
        info, instock, err := crawler.ParsePage(font, detail, itemid, "", shoptype)
        if err != nil {
            crawler.SaveFailed(itemid, "", shoptype, mgofailed)
            log.Error(err)
            continue
        }
        crawler.Save(info, itemDepot)
        sid := strconv.Itoa(info.Sid)
        crawler.SaveSuccessed(itemid, sid, shoptype, font, detail, true, instock, mgopages)
        s_fetchShop(sid, shopDepot)
        break
    }

}

func s_fetchShop(sid string, shopDepot *mgo.Collection) {
    for i := 0; i < 10; i++ {
        shoplink := fmt.Sprintf("http://shop%s.taobao.com", sid)
        shopinfo, err := crawler.FetchShopDetail(shoplink)
        if err != nil {
            log.ErrorfType("shop err", "%s", err.Error())
            continue
        }
        shop := models.ShopItem{}
        shop.ShopInfo = shopinfo
        shop.CreatedTime = time.Now()
        shop.LastUpdatedTime = time.Now()
        shop.LastCrawledTime = time.Now()
        shop.Status = "queued"
        shop.CrawlerInfo = &models.CrawlerInfo{Priority: 10, Cycle: 720}
        shop.ExtendedInfo = &models.TaobaoShopExtendedInfo{
            Type:           shopinfo.ShopType,
            Orientational:  false,
            CommissionRate: -1,
        }
        shopDepot.Insert(shop)
        break
    }
}
func fetchWithShopid(itemid string, shopid int, shoptype string, mgofailed, mgopages, itemDepot *mgo.Collection) {
    font, detail, instock, err, isWeb := crawler.FetchItem(itemid, shoptype)
    sid := strconv.Itoa(shopid)
    if err != nil {
        crawler.SaveFailed(itemid, sid, shoptype, mgofailed)
        //log.Error(err)
        return
    }
    if isWeb {
        info, err := crawler.ParseWeb(font, detail, itemid, sid, shoptype)
        if err != nil {
            log.Error(err)
            crawler.SaveFailed(itemid, sid, shoptype, mgofailed)
        }
        crawler.Save(info, itemDepot)
        crawler.SaveSuccessed(itemid, sid, shoptype, font, detail, true, instock, mgopages)

    } else {
        info, instock, err := crawler.ParsePage(font, detail, itemid, sid, shoptype)
        if err != nil {
            crawler.SaveFailed(itemid, sid, shoptype, mgofailed)
            log.Error(err)
            return
        }
        crawler.Save(info, itemDepot)
        crawler.SaveSuccessed(itemid, sid, shoptype, font, detail, true, instock, mgopages)
    }
}

type Item struct {
    Id   string `json:"taobao_id"`
    Nick string `json:"shop_nick"`
}

type Entity struct {
    Id       int64   `json:"entity_id"`
    PostTime float64 `json:"post_time"`
    NoteId   int64   `json:"note_id"`
    Items    []Item  `json:"taobao_item_list"`
}
