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

type Response struct {
    ItemId   string `json:"item_id"`
    TaobaoId string `json:"taobao_id"`
}

type SyncOnlineItem struct {
    Base
}

func (this *SyncOnlineItem) run() {
    defer func() {
        this.start = false
    }()

    for {
        if this.start == false {
            return
        }
        this.syncOnline()
        time.Sleep(1 * time.Hour)
    }
}

func (this *SyncOnlineItem) syncOnline() {
    session, err := mgo.Dial(MGOHOST)
    if err != nil {
        log.ErrorfType("mongo err", "%s", err.Error())
        return
    }
    defer func() {
        if session != nil {
            session.Close()
        }
    }()
    shopDepot := session.DB(MANGO).C(SHOPS_DEPOT)
    itemDepot := session.DB(MANGO).C(ITEMS_DEPOT)
    count := 1000
    offset := 0
    syncLink := beego.AppConfig.String("synconline::link")
    for {
        if this.start == false {
            return
        }
        resp, err := http.Get(fmt.Sprintf(syncLink+"?count=%d&offset=%d", count, offset))
        if err != nil {
            log.ErrorfType("server err", "%s %s", "服务器错误", err.Error())
            return
        }
        body, err := ioutil.ReadAll(resp.Body)
        resp.Body.Close()
        if err != nil {
            log.ErrorfType("server err", "%s %s", "服务器错误", err.Error())
            return
        }
        r := make([]Response, 0)
        json.Unmarshal(body, &r)
        if len(r) == 0 {
            if resp.StatusCode != 200 {
                log.ErrorfType("server err", "%s %s", "服务器错误")
            }
            fmt.Println("没有商品")
            break
        }

        allNew := true

        for _, v := range r {
            num_iid, _ := strconv.Atoi(v.TaobaoId)
            item := models.TaobaoItem{}
            err := itemDepot.Find(bson.M{"num_iid": int(num_iid)}).One(&item)
            for i := 0; i < 10; i++ {
                if err != nil && err.Error() == "not found" {
                    log.Error(err)
                    font, detail, shoptype, _, err := crawler.FetchWithOutType(v.TaobaoId)
                    if err != nil {
                        log.Error(err)
                        continue
                    }

                    nick, err := crawler.GetShopNick(font)
                    if err != nil {
                        log.Error(err)
                        continue
                    }
                    shop := models.ShopItem{}
                    err = shopDepot.Find(bson.M{"shop_info.nick": nick}).One(&shop)
                    if err != nil && err.Error() == "not found" {
                        log.Error(err)
                        link, err := crawler.GetShopLink(font)
                        if err != nil {
                            log.Error(err)
                            continue
                        }

                        sh, err := crawler.FetchShopDetail(link)
                        if err != nil {
                            log.Error(err)
                            continue
                        }

                        shop.ShopInfo = sh
                        shop.CreatedTime = time.Now()
                        shop.LastUpdatedTime = time.Now()
                        shop.Status = "queued"
                        shop.CrawlerInfo = &models.CrawlerInfo{Priority: 10, Cycle: 720}
                        shop.ExtendedInfo = &models.TaobaoShopExtendedInfo{
                            Type:           shoptype,
                            Orientational:  false,
                            CommissionRate: -1,
                        }

                        shopDepot.Insert(&shop)
                    }
                    if shop.ShopInfo == nil {
                        continue
                    }
                    sid := strconv.Itoa(shop.ShopInfo.Sid)
                    info, _, err := crawler.ParsePage(font, detail, v.TaobaoId, sid, shoptype)
                    if err != nil {
                        log.Error(err)
                        continue
                    }
                    info.GuokuItemid = v.ItemId
                    SAdd("jobs:synconlineitem", v.ItemId)
                    crawler.Save(info, itemDepot)
                }
                if item.ItemId != "" {
                    if offset > 10000 {
                        allNew = false
                        break
                    }
                } else {
                    itemDepot.Update(bson.M{"num_iid": int(num_iid)}, bson.M{"$set": bson.M{"item_id": v.ItemId}})
                    SAdd("jobs:synconlineitem", v.ItemId)
                }
                break //能执行到这里，说明这个item爬虫成功，不需要进行十次循环
            }
        }
        if !allNew {
            fmt.Println("break")
            break
        }
        offset += count
    }
}
