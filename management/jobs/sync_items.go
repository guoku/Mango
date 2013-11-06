package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "io/ioutil"
    "strconv"
    "time"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
    "Mango/management/models"
    "Mango/management/taobaoclient"
)

var MgoSession *mgo.Session
var MgoDbName string = "mango"
type Response struct {
    ItemId string `json:"item_id"`
    EntityId string `json:"entity_id"`
    TaobaoId string `json:"taobao_id"`
}

func init() {
    session, err := mgo.Dial("10.0.1.23")
    if err != nil {
        panic(err)
    }
    MgoSession = session
}

func syncOnlineItems() {
    count := 1000
    offset := 0
    ic := MgoSession.DB(MgoDbName).C("raw_taobao_items_depot")
    sc := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
    for {
        resp, err := http.Get(fmt.Sprintf("http://api.guoku.com:10080/management/taobao/item/sync/?count=%d&offset=%d", count, offset))
        if err != nil {
            time.Sleep(60)
            continue
        }
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            time.Sleep(60)
            continue
        }
        r := make([]Response, 0)
        json.Unmarshal(body, &r)
        if len(r) == 0 {
            break;
        }
        allNew := true
        for _, v := range r {
            iid, _ := strconv.Atoi(v.TaobaoId)
            item := models.TaobaoItem{}
            err := ic.Find(bson.M{"num_iid" : int(iid)}).One(&item)
            if err != nil && err.Error() == "not found" {
                //fmt.Println("not found", iid)
                ti, te := taobaoclient.GetTaobaoItemInfo(int(iid))
                if te != nil {
                    fmt.Println("error", te.Error())
                    continue;
                }
                item.ApiData = ti
                item.ApiDataReady = true
                item.NumIid = int(iid)
                shop := models.ShopItem{}
                se := sc.Find(bson.M{"shop_info.nick" : ti.Nick}).One(&shop)
                if se != nil {
                    if se.Error() == "not found" {
                        ts, te := taobaoclient.GetTaobaoShopInfo(ti.Nick)
                        if te != nil {
                            fmt.Println("shop error", te.Error())
                            continue
                        }

                        shop.ShopInfo = ts
                        shop.CreatedTime = time.Now()
                        shop.LastUpdatedTime = time.Now()
                        shop.Status = "queued"
                        shop.CrawlerInfo = &models.CrawlerInfo{Priority: 10, Cycle: 720}
                        shop.ExtendedInfo = &models.TaobaoShopExtendedInfo{Type: "unknown", Orientational: false, CommissionRate: -1}
                        se = sc.Insert(&shop)
                    } else {
                        continue
                    }
                }
                item.Sid = shop.ShopInfo.Sid
                item.CreatedTime = time.Now()
                item.ItemId = v.ItemId
                ic.Insert(&item)
                fmt.Println("insert", item.NumIid)

                continue
            }
            if item.ItemId != "" {
                allNew = false
                fmt.Println("already exists", item.NumIid)
                break
            } else {
                ic.Update(bson.M{"num_iid" : int(iid)}, bson.M{"$set" : bson.M{"item_id" : v.ItemId}})
                fmt.Println("update", item.NumIid)
            }
        }
        if !allNew {
            break
        }
        offset += count
    }
}

func uploadOfflineItems() {
    
}
func main() {
    go func() {
        syncOnlineItems()
        time.Sleep(60 * 60 * 2)
    }()
    g
    select {}
}
