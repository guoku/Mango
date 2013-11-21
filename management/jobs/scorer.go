package main

import (
    "fmt"
    "math"
    "strconv"
    "time"

    "Mango/management/old_guoku_models"
    "Mango/management/models"
	"github.com/astaxie/beego/orm"
    _ "github.com/go-sql-driver/mysql"
    "labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const NUM_EVERY_TIME = 10000
var MgoSession *mgo.Session
var MgoDbName string = "mango"
func init() {
    orm.RegisterDriver("mysql", orm.DR_MySQL)
    orm.RegisterDataBase("default", "mysql", "root:123456@tcp(localhost:3306)/guoku?charset=utf8", 30)
    orm.RegisterModel(new(old_guoku_models.BaseEntity), new(old_guoku_models.BaseItem), new(old_guoku_models.BaseTaobaoItem), new(old_guoku_models.GuokuEntityLike))
    orm.RunCommand()
    //orm.Debug = true
    session, err := mgo.Dial("10.0.1.23")
    if err != nil {
        panic(err)
    }
    MgoSession = session
}

func getLikes(c *mgo.Collection, item models.TaobaoItemStd) {
    o := orm.NewOrm()
    taobaoId := strconv.Itoa(item.NumIid)
    entity := &old_guoku_models.BaseEntity{}
    err := o.QueryTable("base_entity").Filter("BaseItem__BaseTaobaoItem__TaobaoId", taobaoId).One(entity)
    if err != nil || entity.Id == 0 {
        fmt.Println(err.Error())
        return
    }
    fmt.Println(entity.Id)
    count, cerr := o.QueryTable("guoku_entity_like").Filter("EntityId", entity.Id).Count()
    fmt.Println(count)
    if cerr != nil {
        count = 0
    }
    isSelection := false
    if count > 30 {
       isSelection = true
    }
    c.Update(bson.M{"num_iid": item.NumIid}, bson.M{"$set": bson.M{"score_info.likes" : count, "is_selection" : isSelection}})
}

func getTaobaoItemLikes() {
    o := orm.NewOrm()
    var maps []orm.Params
    num, err := o.Raw("select taobao_id from base_taobao_item").Values(&maps)
    var array []int
    if num > 0 && err ==nil {
        for i, _ := range maps {
            iid, e :=  strconv.Atoi(maps[i]["taobao_id"].(string))
            if e == nil {
                array = append(array, iid)
            }
        }
    }
    for {
        c := MgoSession.DB(MgoDbName).C("taobao_items_depot") 
        results := make([]models.TaobaoItemStd, 0)
        err = c.Find(bson.M{"score_info.likes" : nil, "num_iid" : bson.M{"$in" : array}}).Limit(NUM_EVERY_TIME).All(&results)
        if err != nil {
            fmt.Println(err.Error())
            return
        }
        for _, record := range results {
            fmt.Println(record.Sid, record.NumIid)
            getLikes(c, record)
        }
        time.Sleep(1 * time.Minute)
    }
}

func getSingleTaobaoShopScoreInfo(record *models.ShopItem) {
        c := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
        ic := MgoSession.DB(MgoDbName).C("taobao_items_depot")
        o := orm.NewOrm()
        nick := record.ShopInfo.Nick
        items := make([]old_guoku_models.BaseTaobaoItem, 0)
        o.QueryTable("base_taobao_item").Filter("shop_nick", nick).All(&items)
        totalLikes := 0
        totalSelections := 0
        tbItems := make([]models.TaobaoItemStd, 0)
        for _, item := range items {
            entity := &old_guoku_models.BaseEntity{}
            err := o.QueryTable("base_entity").Filter("BaseItem__BaseTaobaoItem__TaobaoId", item.TaobaoId).One(entity)
            if err != nil || entity.Id == 0 {
                fmt.Println(err.Error())
                continue
            }
            temp, _ := o.QueryTable("guoku_entity_like").Filter("EntityId", entity.Id).Count()
            count := int(temp)
            taobaoItem := models.TaobaoItemStd{}
            tNumIid, _ :=  strconv.Atoi(item.TaobaoId)
            ic.Find(bson.M{"num_iid" : tNumIid}).One(&taobaoItem)
            if taobaoItem.NumIid == 0 {
                taobaoItem.NumIid = tNumIid
                taobaoItem.Sid = record.ShopInfo.Sid
                taobaoItem.CreatedTime = time.Now()
                ic.Insert(&taobaoItem)
            }
            scoreInfo := models.ScoreInfo{}
            scoreInfo.Likes = count
            if count > 20 {
                scoreInfo.IsSelection = true
                totalSelections += 1
            }
            scoreInfo.UpdatedTime =  time.Now()
            taobaoItem.ScoreInfo = &scoreInfo
            tbItems = append(tbItems, taobaoItem)
            ic.Update(bson.M{"num_iid" : tNumIid},
                      bson.M{"$set" : bson.M{"score_info":&scoreInfo}})
            totalLikes += count

        }
        fmt.Println("total:", totalLikes, totalSelections)
        err := c.Update(bson.M{"shop_info.sid": record.ShopInfo.Sid}, bson.M{"$set": bson.M{"score_info" : bson.M{}}})
        err = c.Update(bson.M{"shop_info.sid": record.ShopInfo.Sid}, bson.M{"$set": bson.M{"score_info.total_likes" : totalLikes,
                                                  "score_info.total_selections" : totalSelections, 
                                                  "score_info.total_items" : len(items),
                                                  "score_info.updated_time" : time.Now()}})
        if err != nil {
            fmt.Println(err.Error())
        }
        record.ScoreInfo = &models.ShopScoreInfo{}
        record.ScoreInfo.TotalSelections = totalSelections
        record.ScoreInfo.TotalLikes = totalLikes
        record.ScoreInfo.TotalItems = len(items)
        shopScore := math.Log(float64(1 + totalLikes + 20.0 * totalSelections + 10.0 * totalLikes / (1 + totalSelections)))
        for _, item := range tbItems {
            score := float64(item.ScoreInfo.Likes * 10.0)
            if item.ScoreInfo.IsSelection {
                score = score * 1.1
            }
            score = (1 + math.Log(score + 1) * math.Log(score + 1)) * ( 1 + shopScore)
            fmt.Println(score)
            ic.Update(bson.M{"num_iid" : item.NumIid}, bson.M{"$set" : bson.M{"score": score, "score_updated_time" : time.Now()}})
        }
}

/*
func getTaobaoShopScoreInfo() {
    c := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
    ic := MgoSession.DB(MgoDbName).C("raw_taobao_items_depot")
    results := make([]models.ShopItem, 0)
    c.Find(nil).Select(bson.M{"shop_info.nick": 1, "shop_info.sid" : 1}).All(&results)
    o := orm.NewOrm()
    for _, record := range results {
        fmt.Println("sid:" , record.ShopInfo.Sid, record.ShopInfo.Nick)
        nick := record.ShopInfo.Nick
        items := make([]old_guoku_models.BaseTaobaoItem, 0)
        o.QueryTable("base_taobao_item").Filter("shop_nick", nick).All(&items)
        totalLikes := 0
        totalSelections := 0
        tbItems := make([]models.TaobaoItem, 0)
        for _, item := range items {
            entity := &old_guoku_models.BaseEntity{}
            err := o.QueryTable("base_entity").Filter("BaseItem__BaseTaobaoItem__TaobaoId", item.TaobaoId).One(entity)
            if err != nil || entity.Id == 0 {
                fmt.Println(err.Error())
                continue
            }
            temp, _ := o.QueryTable("guoku_entity_like").Filter("EntityId", entity.Id).Count()
            count := int(temp)
            taobaoItem := models.TaobaoItem{}
            tNumIid, _ :=  strconv.Atoi(item.TaobaoId)
            ic.Find(bson.M{"num_iid" : tNumIid}).One(&taobaoItem)
            if taobaoItem.NumIid == 0 {
                taobaoItem.NumIid = tNumIid
                taobaoItem.Sid = record.ShopInfo.Sid
                taobaoItem.CreatedTime = time.Now()
                ic.Insert(&taobaoItem)
            }
            scoreInfo := models.ScoreInfo{}
            scoreInfo.Likes = count
            if count > 20 {
                scoreInfo.IsSelection = true
                totalSelections += 1
            }
            scoreInfo.UpdatedTime =  time.Now()
            taobaoItem.ScoreInfo = &scoreInfo
            tbItems = append(tbItems, taobaoItem)
            ic.Update(bson.M{"num_iid" : tNumIid},
                      bson.M{"$set" : bson.M{"score_info":&scoreInfo}})
            totalLikes += count

        }
        fmt.Println("total:", totalLikes, totalSelections)
        err := c.Update(bson.M{"shop_info.sid": record.ShopInfo.Sid}, bson.M{"$set": bson.M{"score_info.total_likes" : totalLikes,
                                                  "score_info.total_selections" : totalSelections,
                                                  "score_info.total_items" : len(items),
                                                  "score_info.updated_time" : time.Now()}})
        if err != nil {
            fmt.Println(err.Error())
        }
        shopScore := math.Log(float64(1 + totalLikes + 20.0 * totalSelections + 10.0 * totalLikes / (1 + totalSelections)))
        for _, item := range tbItems {
            score := float64(item.ScoreInfo.Likes * 10.0)
            if item.ScoreInfo.IsSelection {
                score = score * 1.1
            }
            score = (1 + math.Log(score + 1) * math.Log(score + 1)) * ( 1 + shopScore)
            fmt.Println(score)
            ic.Update(bson.M{"num_iid" : item.NumIid}, bson.M{"$set" : bson.M{"score": score, "score_updated_time" : time.Now()}})
        }

    }
}
*/
func calculateScore() {
    c := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
    shops := make([]models.ShopItem, 0)
    c.Find(nil).All(&shops)
    ic := MgoSession.DB(MgoDbName).C("taobao_items_depot")
    for i := range shops {
        fmt.Println("shop_name", shops[i].ShopInfo.Nick)
        items := make([]models.TaobaoItemStd, 0)
        ic.Find(bson.M{"sid" : shops[i].ShopInfo.Sid, "score" : bson.M{"$eq" : 0}}).All(&items)
        if shops[i].ScoreInfo == nil {
            fmt.Println("score info null")
            getSingleTaobaoShopScoreInfo(&shops[i])
            fmt.Println("after", shops[i].ScoreInfo.TotalLikes)
        }
        totalLikes := shops[i].ScoreInfo.TotalLikes
        totalSelections := shops[i].ScoreInfo.TotalSelections
        shopScore := math.Log(float64(1 + totalLikes + 20.0 * totalSelections + 10.0 * totalLikes / (1 + totalSelections)))
        for _, item := range items {
            var score float64
            if item.ScoreInfo == nil {
                item.ScoreInfo = &models.ScoreInfo{Likes : 0, IsSelection:false, UpdatedTime:time.Now()}
                ic.Update(bson.M{"num_iid" : item.NumIid}, bson.M{"$set" : bson.M{"score_info": item.ScoreInfo}})
            }
            score = float64(item.ScoreInfo.Likes * 10.0)
            if item.ScoreInfo.IsSelection {
                score = score * 1.1
            }
            score = (1 + math.Log(score + 1) * math.Log(score + 1)) * ( 1 + shopScore)
            fmt.Println(item.NumIid, score)
            ic.Update(bson.M{"num_iid" : item.NumIid}, bson.M{"$set" : bson.M{"score": score, "score_updated_time" : time.Now()}})
        }
        fmt.Println("score ends")
    }
}

func main() {
    /*
   for {
       scanTaobaoItems()
       time.Sleep(1 * time.Minute)
   }
   */
    //go getTaobaoItemLikes()
    //getTaobaoShopScoreInfo()
    for {
        calculateScore()
        time.Sleep(time.Hour)
    }
}

