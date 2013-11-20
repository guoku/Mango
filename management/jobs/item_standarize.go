package main
import (
    "fmt"
    "runtime"
    "strings"
    "strconv"
    "time"

    "Mango/management/models"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
)
const NUM_CPU = 4
var MgoSession *mgo.Session
var dbName = "mango"
var channel = make(chan int, 20)
const NUM_EVERY_TIME = 50000
var c *mgo.Collection
var nc *mgo.Collection
var pc *mgo.Collection
func init() {
    session, err := mgo.Dial("10.0.1.23")
    if err != nil {
        panic(err)
    }
    MgoSession = session
    c = MgoSession.DB(dbName).C("raw_taobao_items_depot")
    nc = MgoSession.DB(dbName).C("taobao_items_depot")
    pc = MgoSession.DB(dbName).C("taobao_props")
}

func extractProps(propsName string, props *map[string]string) {
    fmt.Println("1", propsName)
    propsArray := strings.Split(propsName, ";")
    fmt.Println("2", propsArray)
    for _, v := range propsArray {
        fmt.Println(v)
        ps := strings.Split(v, ":")
        if len(ps) != 4 {
            continue
        }
        fmt.Println(ps)
        pid, _ := strconv.Atoi(ps[0])
        vid, _ := strconv.Atoi(ps[1])
        pname := strings.Trim(ps[2], " ")
        vname := strings.Trim(ps[3], " ")
        /*
        p := models.TaobaoProp{}
        err := pc.Find(bson.M{"type": "Prop", "taobao_id": int(pid)}).One(&p)
        if err != nil && err.Error() == "not found" {
            p.TaobaoId = int(pid)
            p.Name = pname
            p.Type = "Prop"
            pc.Insert(&p)
        }
        */
        info, _ := pc.Upsert(bson.M{"type" : "Prop", "taobao_id" : int(pid)}, bson.M{"$set" : bson.M{"name" : pname}})
        fmt.Println(info)
        /*v := models.TaobaoProp{}
        err = pc.Find(bson.M{"type": "Value", "taobao_id": int(pid)}).One(&v)
        if err != nil && err.Error() == "not found" {
            v.TaobaoId = int(vid)
            v.Name = vname
            v.Type = "Value"
            pc.Insert(&v)
        }
        */

        info, _ = pc.Upsert(bson.M{"type" : "Value", "taobao_id" : int(vid)}, bson.M{"$set" : bson.M{"name" : vname}})
        fmt.Println(info)
        (*props)[pname] = vname
    }
    fmt.Println(*props)
}

func convertItem(item *models.TaobaoItem) {
    itemStd := models.TaobaoItemStd{}
    itemStd.NumIid = item.NumIid
    itemStd.Sid = item.Sid
    itemStd.Score = item.Score
    itemStd.ScoreUpdatedTime = item.ScoreUpdatedTime
    itemStd.CreatedTime = item.CreatedTime
    itemStd.DataUpdatedTime = item.ApiDataUpdatedTime
    itemStd.ScoreInfo = item.ScoreInfo
    itemStd.ItemId = item.ItemId
    itemStd.Uploaded = item.Uploaded
    itemStd.InStock = true
    if item.ApiDataReady {
        itemStd.Nick = item.ApiData.Nick
        itemStd.DetailUrl = item.ApiData.DetailUrl
        itemStd.Title = item.ApiData.Title
        itemStd.Desc = item.ApiData.Desc
        itemStd.Cid = item.ApiData.Cid
        itemStd.Price = item.ApiData.Price
        itemStd.Location = item.ApiData.Location
        itemStd.Props = make(map[string]string)
        if strings.Trim(item.ApiData.PropsName, " ") != "" {
            extractProps(item.ApiData.PropsName, &itemStd.Props)
        }
    }
    c.Update(bson.M{"num_iid" : item.NumIid}, bson.M{"$set" : bson.M{"converted" : true}})
    nc.Insert(&itemStd)
    <-channel
}

func scanTaobaoItems() {
    for {
        fmt.Println("start")
        results := make([]models.TaobaoItem, 0)
        err := c.Find(bson.M{"converted" : nil}).Limit(NUM_EVERY_TIME).All(&results)
        fmt.Println("2")
        if err != nil {
            fmt.Println(err.Error())
            time.Sleep(time.Minute)
            continue
        }
        fmt.Println("3")
        if len(results) == 0 {
            break
        }
        fmt.Println("4")
        for i := range results {
            channel <- 1
            fmt.Println(results[i].NumIid)
            go convertItem(&results[i])
        }
    }
}
func convertItemImgs(item *models.TaobaoItem) {
    itemStd := models.TaobaoItemStd{}
    err := nc.Find(bson.M{"num_iid":item.NumIid}).One(&itemStd)
    if err != nil && err.Error() == "not found" {
        convertItem(item)
        nc.Find(bson.M{"num_iid":item.NumIid}).One(&itemStd)
    }
    if item.ApiDataReady && item.ApiData.ItemImgs != nil {
        urls := make([]string, 0)
        for _, v := range item.ApiData.ItemImgs.ItemImgArray {
            urls = append(urls, v.Url)
        }
        nc.Update(bson.M{"num_iid":item.NumIid}, bson.M{"$set": bson.M{"item_imgs": urls}})
    }
    c.Update(bson.M{"num_iid" : item.NumIid}, bson.M{"$set" : bson.M{"img_converted" : true}})
    <-channel
}

func scanTaobaoItemImgs() {
    for {
        fmt.Println("start")
        results := make([]models.TaobaoItem, 0)
        err := c.Find(bson.M{"img_converted" : nil}).Limit(NUM_EVERY_TIME).All(&results)
        fmt.Println("2")
        if err != nil {
            fmt.Println(err.Error())
            time.Sleep(time.Minute)
            continue
        }
        fmt.Println("3")
        if len(results) == 0 {
            break
        }
        fmt.Println("4")
        for i := range results {
            channel <- 1
            fmt.Println(results[i].NumIid)
            go convertItemImgs(&results[i])
        }
    }
}

func main() {
    runtime.GOMAXPROCS(NUM_CPU)
    scanTaobaoItemImgs()
}
