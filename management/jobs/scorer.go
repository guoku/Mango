package main

import (
    "strconv"

    "Mango/management/old_guoku_models"
    "Mango/management/models"
	"github.com/astaxie/beego/orm"
    "labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const NUM_EVERY_TIME =50000

func init() {
    orm.RegisterDriver("mysql", orm.DR_MySQL)
    orm.RegisterDataBase("default", "mysql", "root:123456@tcp(localhost:3306)/guoku?charset=utf8", 30)
    orm.RegisterModel(new(old_guoku_models.BaseEntity), new(old_guoku_models.BaseItem), new(old_guoku_models.BaseTaobaoItem), new(old_guoku_models.GuokuEntityLike))
    orm.RunCommand()
    orm.Debug = true
}

func getScore(c *mgo.Collection, item models.TaobaoItem) {
    o := orm.NewOrm()
    taobaoId := strconv.Itoa(item.NumIid)
    o.QueryTable("base_taobao_item").Filter("taobao_id", taobaoId).One(&result)
    
    taobaoId
}

func scanTaobaoItems() {
    c := MgoSession.DB(dbName).C("raw_taobao_items_depot") 
    results := make([]models.TaobaoItem, 0)
    err := c.Find(bson.M{"score" : 0}).Limit(NUM_EVERY_TIME).All(&results)
    if err != nil {
        panic(err)
    }
    for _, record := range results {
        fmt.Println(record.Sid, record.NumIid)
        getScore(c, record)
    }
}

func main() {
   for {
       scanTaobaoItems()
       time.Sleep(1 * time.Minute)
   }
}

