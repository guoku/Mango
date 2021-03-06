package main

import (
    "errors"
    "flag"
    "fmt"
    "runtime"
    "time"

    "Mango/management/models"
	"Mango/management/taobaoclient"

    "github.com/pelletier/go-toml"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var MgoSession *mgo.Session
var dbName string

const NUM_CPU = 4
var channel = make(chan int, 20)
const NUM_EVERY_TIME = 5000
func init() {
    var env string
    flag.StringVar(&env, "env", "test", "program environment")
    flag.Parse()
    var mongoSetting *toml.TomlTree
    conf, err := toml.LoadFile("conf/config.toml")
    switch env {
        case "debug":
            mongoSetting = conf.Get("mongodb.debug").(*toml.TomlTree)
        case "staging":
            mongoSetting = conf.Get("mongodb.staging").(*toml.TomlTree)
        case "prod":
            mongoSetting = conf.Get("mongodb.prod").(*toml.TomlTree)
        case "test":
            mongoSetting = conf.Get("mongodb.test").(*toml.TomlTree)
        default:
            panic(errors.New("Wrong Environment Flag Value. Should be 'debug', 'staging' or 'prod'"))
    }
    fmt.Println(mongoSetting.Get("host").(string), mongoSetting.Get("db").(string))
    //session, err := mgo.Dial(mongoSetting.Get("host").(string))
    session, err := mgo.Dial("10.0.1.23")
    if err != nil {
        panic(err)
    }
    MgoSession = session
    dbName = mongoSetting.Get("db").(string)
}

func getApiData(c *mgo.Collection, numIid int) {
    itemInfo, topErr := taobaoclient.GetTaobaoItemInfo(numIid)
    fmt.Println("getting data", numIid)
    if topErr != nil {
        fmt.Println(topErr.Error())
        if topErr.SubCode == "isv.item-get-service-error:ITEM_NOT_FOUND" ||
           topErr.SubCode == "isv.item-is-delete:invalid-numIid-or-iid" ||
           topErr.SubCode == "isv.invalid-permission:get-item" {
            fmt.Println("remove")
            info, err := c.RemoveAll(bson.M{"num_iid" : numIid})
            fmt.Println(info)
            if err != nil {
                fmt.Println(err.Error())
            }
        }
        <-channel
        return
    }
    if itemInfo == nil {
        <-channel
        return
    }
    fmt.Println(numIid, itemInfo.Title)
    change := bson.M{"$set" : bson.M{"api_data" : *itemInfo, "api_data_ready" : true, "api_data_updated_time" : time.Now()}}
    c.Update(bson.M{"num_iid" : numIid}, change)
    <-channel
}

func scanTaobaoItems() {
    c := MgoSession.DB(dbName).C("raw_taobao_items_depot")
    results := make([]models.TaobaoItem, 0)
    err := c.Find(bson.M{"api_data_ready" : false}).Sort("-score").Limit(NUM_EVERY_TIME).All(&results)
    if err != nil {
        panic(err)
    }
    for _, record := range results {
        channel <- 1
        fmt.Println("start", record.Sid, record.NumIid, record.Score)
        go getApiData(c, record.NumIid)
    }
    time.Sleep(time.Second)
}

func main() {
    runtime.GOMAXPROCS(NUM_CPU)
    for {
        scanTaobaoItems()
    }
}
