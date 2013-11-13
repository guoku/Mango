package main


import (
    "fmt"
    "time"

	//"github.com/jason-zou/taobaosdk/rest"
    "Mango/management/models"
    "Mango/management/taobaoclient"
    "labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var MgoSession *mgo.Session
var MgoDbName string = "mango"
func init() {
    session, err := mgo.Dial("localhost")
    if err != nil {
        panic(err)
    }
    MgoSession = session
}

func getTaobaoCatsWithPid(parentId int) {
    cats, err := taobaoclient.GetItemCatsInfo(parentId)
    if err != nil {
        fmt.Println(err.Error())
        return
    }
    c := MgoSession.DB(MgoDbName).C("taobao_cats")
    for _, v := range cats {
        fmt.Println(v.Cid, v.Name)
        /*
        tic := models.TaobaoItemCat{Id : bson.NewObjectId(),
                                    ItemCat : v,
                                    UpdatedTime : time.Now()}
        */
        changeInfo, e := c.Upsert(bson.M{"item_cat.cid" : v.Cid}, bson.M{"$set" : bson.M{"item_cat" : v, "updated_time":  time.Now(), "extended" : false}})
        if e != nil {
            fmt.Println(e.Error())
        }
        fmt.Println("Updated:", changeInfo.Updated, "Removed:", changeInfo.Removed, "UpsertedId:", changeInfo.UpsertedId)
    }
}

func getTaobaoCats() {
    c := MgoSession.DB(MgoDbName).C("taobao_cats")
    count, _ := c.Find(nil).Count()
    if count == 0 {
        getTaobaoCatsWithPid(0)
    }
    for {
        cats := make([]models.TaobaoItemCat, 0)
        c.Find(bson.M{"item_cat.is_parent" : true, "extended" : false}).All(&cats)
        for _, v := range cats {
            getTaobaoCatsWithPid(v.ItemCat.Cid)
            c.Update(bson.M{"_id" : v.Id}, bson.M{"$set" : bson.M{"extended" : true}})
        }
    }
}

func getCatsItemNum(parentCid int) (int, int) {
    fmt.Println("start", parentCid)
    c := MgoSession.DB(MgoDbName).C("taobao_cats")
    cats := make([]models.TaobaoItemCat, 0)
    fmt.Println("query parent", parentCid)
    c.Find(bson.M{"item_cat.parent_cid" : parentCid}).All(&cats)
    fmt.Println("query finished")
    ic := MgoSession.DB(MgoDbName).C("raw_taobao_items_depot")
    fmt.Println("query item num", parentCid)
    itemNum, err := ic.Find(bson.M{"api_data.cid" : parentCid}).Count()
    fmt.Println("query finished", parentCid,itemNum)
    if err != nil {
        itemNum = 0
    }
    selectionNum, err := ic.Find(bson.M{"api_data.cid" : parentCid, "score_info.is_selection" : true}).Count()
    if err != nil {
        selectionNum = 0
    }
    for _, v := range cats {
        fmt.Println("son", v.ItemCat.Cid)
        in, sn := getCatsItemNum(v.ItemCat.Cid)
        itemNum += in
        selectionNum += sn
        fmt.Println("after son", v.ItemCat.Cid, itemNum, selectionNum)
    }
    fmt.Println("p === ", parentCid, itemNum, selectionNum)
    c.Update(bson.M{"item_cat.cid" : parentCid}, bson.M{"$set" : bson.M{"item_num" : itemNum, "selection_num" : selectionNum}})
    return itemNum, selectionNum
}

func main() {
     //getTaobaoCats()
     for {
        fmt.Println("all:", getCatsItemNum(0))
        time.Sleep(time.Hour * 12)
     }
}
