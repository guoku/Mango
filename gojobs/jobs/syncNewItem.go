package jobs

import (
    "Mango/gojobs/log"
    "Mango/gojobs/models"
    "encoding/json"
    "github.com/astaxie/beego"
    "io/ioutil"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
    "net/http"
    "net/url"
    "time"
)

type SyncNewItem struct {
    Base
}

func (this *SyncNewItem) run() {
    defer func() {
        this.start = false
    }()

    for {
        if this.start == false {
            return
        }
        sync()
        time.Sleep(1 * time.Hour)
    }
}

type CreateItemsResp struct {
    ItemId   string `json:"item_id"`
    EntityId string `json:"entity_id"`
    Status   string `json:"status"`
}

/*
 把新抓取到的符合要求的数据post上去
*/
func sync() {
    //taobaoCats := MongoInit(MGOHOST, MANGO, "taobao_cats")
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
        log.Info(v.ItemCat.Cid)
        itemDepot.Find(
            bson.M{"cid": v.ItemCat.Cid, "uploaded": false,
                "item_imgs.0": bson.M{"$exists": true},
                "score":       bson.M{"$gt": 2}}).All(&items)
        for i := range items {
            if items[i].Title == "" {
                continue
            }
            params := url.Values{}
            GetUploadItemParams(&items[i], &params, v.MatchedGuokuCid)
            resp, err := http.PostForm(uploadLink, params)

            if err != nil {
                log.ErrorfType("sync err", "%s", err.Error())
                continue
            }
            if resp.StatusCode != 200 {
                log.ErrorfType("server err", "%s", "服务器错误")
                return
            }
            body, err := ioutil.ReadAll(resp.Body)
            resp.Body.Close()
            if err != nil {
                log.ErrorfType("sync err", "%s", err.Error())
                continue
            }

            r := CreateItemsResp{}
            json.Unmarshal(body, &r)
            if r.Status == "success" || r.Status == "updated" {
                SAdd("jobs:syncnewitem", items[i].ItemId)
                err = itemDepot.Update(
                    bson.M{"num_iid": items[i].NumIid},
                    bson.M{"$set": bson.M{"item_id": r.ItemId,
                        "uploaded": true}})
            } else if r.ItemId != "" {
                err = itemDepot.Update(bson.M{"num_iid": items[i].NumIid},
                    bson.M{"$set": bson.M{"item_id": r.ItemId,
                        "uploaded": false}})
                if err != nil {
                    log.ErrorfType("sync err", "%s", err.Error())
                    continue
                }
            } else {
                log.ErrorfType("server err", "%s", "服务器错误")
            }

        }
    }
}
