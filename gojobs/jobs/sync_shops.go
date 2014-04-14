package jobs

import (
    "Mango/gojobs/log"
    "Mango/gojobs/models"
    "encoding/json"
    "fmt"
    "github.com/astaxie/beego"
    "io/ioutil"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
    "net/http"
    "time"
)

type SyncShop struct {
    Base
}

/*
func (f *SyncShop) Start(arg string, result *string) error {
    if arg == START {
        if f.start {
            *result = START_STATU
            return nil
        }
        *result = "开始启动"
        f.start = true
        go f.run()
    }
    return nil
}

func (f *SyncShop) Stop(arg string, result *string) error {
    if arg == STOP {
        f.start = false
        *result = STOP_STATU
    }
    return nil
}

func (f *SyncShop) Statu(arg string, result *string) error {
    if f.start {
        *result = START_STATU
    } else {
        *result = STOP_STATU
    }
    return nil
}
*/
func (f *SyncShop) run() {
    defer func() {
        f.start = false
    }()

    for {
        if f.start == false {
            return
        }
        syncShop()
        time.Sleep(1 * time.Hour)
    }
}

//每次更新店铺数据之后，都把更新时间记录在mongo.zerg.time里面
type Uptime struct {
    Last time.Time `bson:"last"`
    Name string    `bson:"name"`
}

func syncShop() {
    count := 50
    offset := 0
    //这里貌似奇怪触发了mgo的一个bug，在这里用两个
    //mongoinit,就导致无论如何都找不到mongo的服务器
    //mgoUpdateTime := MongoInit("localhost:27017", "zerg", "time")
    session, err := mgo.Dial(MGOHOST)
    if err != nil {
        fmt.Println(err)
        return
    }
    mgoUpdateTime := session.DB(ZERG).C(UPDATE_TIME)
    mgoShopDepot := MongoInit(MGOHOST, MANGO, SHOPS_DEPOT)
    utime := new(Uptime)
    mgoUpdateTime.Find(bson.M{"name": "last"}).One(&utime)
    date := utime.Last.Format("2006010203")

    for {
        syncShopLink := beego.AppConfig.String("sync::syncshop")
        link := fmt.Sprintf(syncShopLink+"?count=%d&offset=%d&date=%s", count, offset, date)
        resp, err := http.Get(link)
        if err != nil {
            log.ErrorfType("http err", "%s", err.Error())
            return
        }
        if resp.StatusCode != 200 {
            log.ErrorfType("server err", "%s", "服务器错误")
            return
        }
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            log.ErrorfType("http err", "%s", err.Error())
            return
        }
        shops := make([]models.ShopItem, 30)
        json.Unmarshal(body, &shops)
        if len(shops) == 0 {
            err = mgoUpdateTime.Update(bson.M{"name": "last"}, bson.M{"$set": bson.M{"last": time.Now()}})
            if err != nil {
                log.ErrorfType("mgo err", "%s", err.Error())
            }
            fmt.Println("shop is null")
            return
        }

        for _, shop := range shops {
            if shop.ShopInfo == nil {
                continue
            }

            sid := shop.ShopInfo.Sid
            sp := new(models.ShopItem)
            mgoShopDepot.Find(bson.M{"shop_info.sid": sid}).One(&sp)

            if sp.ShopInfo == nil {
                log.Info("新添加的店铺")
                err = mgoShopDepot.Insert(shop)
                if err != nil {
                    log.Error(err)
                }
            } else {
                log.Info("更新的店铺")
                shoptype := shop.ExtendedInfo.Type
                err = mgoShopDepot.Update(bson.M{"shop_info.sid": sid}, bson.M{"$set": bson.M{"extended_info": shop.ExtendedInfo, "crawler_info": shop.CrawlerInfo, "shop_info.shop_type": shoptype}})
                if err != nil {
                    log.Error(err)
                }
            }

            SAdd("jobs:syncshop", fmt.Sprintf("%d", sid))
        }
        offset = offset + count
    }
}
