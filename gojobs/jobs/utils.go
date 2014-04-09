package jobs

import (
    "Mango/gojobs/log"
    "fmt"
    "github.com/astaxie/beego"
    "github.com/xuyu/goredis"
    "labix.org/v2/mgo"
)

var (
    MGOHOST   string
    THREADNUM int

    MANGO       string
    SHOPS_DEPOT string
    ITEMS_DEPOT string

    ZERG     string
    PAGES    string
    FAILED   string
    MINERALS string

    TAOBAO string = "taobao.com"
    TMALL  string = "tmall.com"

    REDIS_CLIENT *goredis.Redis
)

func init() {
    MGOHOST = beego.AppConfig.String("mongo::server")
    MANGO = beego.AppConfig.String("mango::mango")
    SHOPS_DEPOT = beego.AppConfig.String("mango::shops_depot")
    ITEMS_DEPOT = beego.AppConfig.String("mango::items_depot")

    ZERG = beego.AppConfig.String("zerg::zerg")
    PAGES = beego.AppConfig.String("zerg::pages")
    FAILED = beego.AppConfig.String("zerg::failed")
    MINERALS = beego.AppConfig.String("zerg::minerals")
    var err error
    THREADNUM, err = beego.AppConfig.Int("fetchnew::threadnum")
    if err != nil {
        log.Error(err)
        THREADNUM = 1
    }

    redis_server := beego.AppConfig.String("redis::server")
    redis_port := beego.AppConfig.String("redis::port")
    REDIS_CLIENT, err = goredis.Dial(&goredis.DialConfig{Address: fmt.Sprintf("%s:%s", redis_server, redis_port)})
    if err != nil {
        panic(err)
    }

}

func sadd(key, value string) {
    _, err := REDIS_CLIENT.SAdd(key, value)
    if err != nil {
        log.ErrorfType("redis err", "%s", err.Error())
    }
    num, err := REDIS_CLIENT.SCard(key)
    if err != nil {
        log.ErrorfType("redis err", "%s", err.Error())
        return
    }
    if num == 1 {
        REDIS_CLIENT.PExpire(key, 24*3600*1000) //设置过期时间为一天
    }
}

func MongoInit(host, db, collection string) *mgo.Collection {
    session, err := mgo.Dial(host)
    if err != nil {
        log.Error(err)
        panic(err)
    }
    return session.DB(db).C(collection)
}
