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
    "net/url"
    "time"
)

type SyncRefreshItem struct {
    Base
}

func (this *SyncRefreshItem) run() {
    defer func() {
        this.start = false
    }()

    for {
        if this.start == false {
            return
        }

        syncRefresh()
        time.Sleep(1 * time.Hour)
    }
}

/*
   把之前已经post上去过的商品数据，再次post，更新数据
*/
func syncRefresh() {
    session, err := mgo.Dial(MGOHOST)
    if err != nil {
        log.ErrorfType("mongo err", "%s", err.Error())
        return
    }

    taobaoCats := session.DB(MANGO).C("taobao_cats")
    itemDepot := session.DB(MANGO).C(ITEMS_DEPOT)

    uploadLink := beego.AppConfig.String("syncnewitem::link")
    readyCats := make([]models.TaobaoItemCat, 0)
    taobaoCats.Find(bson.M{"matched_guoku_cid": bson.M{"$gt": 0}}).All(&readyCats)

    for _, v := range readyCats {
        items := make([]models.TaobaoItemStd, 0)
        /*
           refreshed:true 表示这个商品之前上传过了，然后，爬虫再次更新了数据
           所以需要再次上传
        */
        itemDepot.Find(bson.M{"cid": v.ItemCat.Cid, "refreshed": true,
            "item_imgs.0": bson.M{"$exists": true},
            "score":       bson.M{"$gt": 2}}).Sort("-refresh_time").All(&items)

        for i := range items {
            if items[i].Title == "" {
                continue
            }

            params := url.Values{}
            GetUploadItemParams(&items[i], &params, v.MatchedGuokuCid)
            resp, err := http.PostForm(uploadLink, params)
            if err != nil {
                log.Error(err)
                continue
            }

            if resp.StatusCode != 200 {
                log.ErrorfType("server err", "%s", "服务器错误")
                return
            }
            body, err := ioutil.ReadAll(resp.Body)
            resp.Body.Close()
            if err != nil {
                log.Error(err)
                continue
            }

            r := CreateItemsResp{}
            json.Unmarshal(body, &r)

            if r.Status == "success" || r.Status == "updated" {
                SAdd("jobs:syncrefreshitem", items[i].ItemId)
                err = itemDepot.Update(bson.M{"num_iid": items[i].NumIid},
                    bson.M{"$set": bson.M{"item_id": r.ItemId, "refreshed": false,
                        "refresh_time": time.Now()}})
                if err != nil {
                    fmt.Println(items[i].NumIid)
                    fmt.Println("%v", r)
                    log.Error(err)
                    return
                }
            } else if r.ItemId != "" {
                err = itemDepot.Update(
                    bson.M{"num_iid": items[i].NumIid},
                    bson.M{"$set": bson.M{"item_id": r.ItemId}},
                )
                if err != nil {
                    log.Error(err)
                    continue
                }
            } else {
                log.ErrorfType("server err", "%s", "服务器错误")
            }
        }
    }
}
