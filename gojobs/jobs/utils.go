package jobs

import (
    "Mango/gojobs/log"
    "Mango/gojobs/models"
    "fmt"
    "github.com/astaxie/beego"
    "github.com/xuyu/goredis"
    "labix.org/v2/mgo"
    "net/url"
    "strconv"
    "time"
)

var (
    MGOHOST string

    MANGO       string
    SHOPS_DEPOT string
    ITEMS_DEPOT string

    ZERG        string
    PAGES       string
    FAILED      string
    MINERALS    string
    UPDATE_TIME string

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
    UPDATE_TIME = beego.AppConfig.String("zerg::time")
    redis_server := beego.AppConfig.String("redis::server")
    redis_port := beego.AppConfig.String("redis::port")
    var err error
    REDIS_CLIENT, err = goredis.Dial(&goredis.DialConfig{Address: fmt.Sprintf("%s:%s", redis_server, redis_port)})
    if err != nil {
        panic(err)
    }

}

//每成功抓取一个商品，就放到redis的集合里面
//便于统计每日成功量，过期时间为一天
func SAdd(key string, value string) {
    if REDIS_CLIENT == nil {
        reconnect()
    }
    _, err := REDIS_CLIENT.SAdd(key, value)
    if err != nil {
        reconnect()
        log.ErrorfType("redis err", "%s", err.Error())
    }
    reply, err := REDIS_CLIENT.TTL(key)
    if err != nil {
        fmt.Println(err)
    }
    if reply <= 0 {
        REDIS_CLIENT.PExpire(key, 24*3600*1000) //设置过期时间为一天
    }
}

func SCard(key string) int64 {

    num, err := REDIS_CLIENT.SCard(key)
    if err != nil {
        reconnect()
        log.ErrorfType("redis err", "%s", err.Error())
        return 0
    } else {
        return num
    }

}

func reconnect() {
    redis_server := beego.AppConfig.String("redis::server")
    redis_port := beego.AppConfig.String("redis::port")
    var err error
    REDIS_CLIENT, err = goredis.Dial(&goredis.DialConfig{Address: fmt.Sprintf("%s:%s", redis_server, redis_port)})
    if err != nil {
        panic(err)
    }
}
func MongoInit(host, db, collection string) *mgo.Collection {
    session, err := mgo.Dial(host)
    if err != nil {
        fmt.Println(err.Error())
        fmt.Println(host)
        fmt.Println(db, collection)
        log.Error(err)
        time.Sleep(30 * time.Second)
        panic(err)
    }
    return session.DB(db).C(collection)
}

func GetUploadItemParams(item *models.TaobaoItemStd, params *url.Values, matchedGuokuCid int) {
    params.Add("taobao_id", strconv.Itoa(item.NumIid))
    params.Add("cid", strconv.Itoa(item.Cid))
    params.Add("taobao_title", item.Title)
    params.Add("taobao_shop_nick", item.Nick)
    if item.PromotionPrice > 0.0 {
        s := fmt.Sprintf("%f", item.PromotionPrice)
        params.Add("taobao_price", s)
    } else {
        s := fmt.Sprintf("%f", item.Price)
        params.Add("taobao_price", s)
    }
    if item.InStock {
        params.Add("taobao_soldout", "0")
    } else {
        params.Add("taobao_soldout", "1")
    }

    itemImgs := item.ItemImgs
    if itemImgs != nil && len(itemImgs) > 0 {
        params.Add("chief_image_url", itemImgs[0])
        for i, _ := range itemImgs {
            params.Add("image_url", itemImgs[i])
        }
    }
    params.Add("category_id", strconv.Itoa(matchedGuokuCid))
}
