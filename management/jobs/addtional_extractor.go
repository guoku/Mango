package main

import (
    "fmt"
    "time"
    "Mango/management/models"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
)

const NUM_EVERY_TIME = 100000


func main() {
    session, err := mgo.Dial("10.0.1.23")
    defer session.Close()
    if err != nil {
        panic(err)
    }
    c := session.DB("mango").C("taobao_items_depot")
    res := bson.M{}
    info := session.DB("words").C("brand_process_info")
    bc := session.DB("words").C("brands")
    info.Find(nil).One(&res)
    startTime := res["last_processed_timestamp"].(time.Time)
    for {
        brands := make(map[string]int)
        items := make([]models.TaobaoItemStd, 0 )
        c.Find(bson.M{"data_updated_time" : bson.M{"$gt" : startTime}}).Sort("data_updated_time").Limit(NUM_EVERY_TIME).All(&items)
        l := len(items)
        if l == 0 {
            break
        }
        for i := 0; i < l; i++ {
            props := items[i].Props
            if brand, ok := props["品牌"]; ok {
                brands[brand] = brands[brand] + 1
            }
        }
        for k, v := range brands {
            fmt.Println(k, v)
            bc.Upsert(bson.M{"name": k}, bson.M{"$inc" : bson.M{"freq": v}})
        }
        startTime = items[l - 1].DataUpdatedTime
        info.UpdateAll(bson.M{}, bson.M{"$set" : bson.M{"last_processed_timestamp": startTime}})
    }
}
