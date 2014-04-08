package jobs

import (
    "Mango/gojobs/crawler"
    "Mango/gojobs/log"
    "github.com/astaxie/beego"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
    "strconv"
    "sync"
    "time"
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

    THREADNUM, _ = beego.AppConfig.Int("fetchnew:threadnum")
}

type FetchNew struct {
    start bool
}

func (f *FetchNew) Start(arg interface{}, result *string) error {
    if arg.(string) == START {
        if f.start {
            *result = "已经启动"
            return nil
        }
        *result = "开始启动"
        f.start = true
        go f.Run()
    }

    return nil
}

func (f *FetchNew) Stop(arg interface{}, result *string) error {
    if arg.(string) == STOP {
        f.start = false
        *result = "已经停止运行"
    }
    return nil
}

func (f *FetchNew) Run() {
    defer func() {
        f.start = false
    }()
    for {
        if f.start == false {
            return
        }

        FetchTaobaoItem(THREADNUM)
    }
}

func FetchTaobaoItem(threadnum int) {
    mgopages := MongoInit(MGOHOST, ZERG, PAGES)
    mgofailed := MongoInit(MGOHOST, ZERG, FAILED)
    mgominer := MongoInit(MGOHOST, ZERG, MINERALS)
    item_depot := MongoInit(MGOHOST, MANGO, ITEMS_DEPOT)
    shop_depot := MongoInit(MGOHOST, MANGO, SHOPS_DEPOT)

    var shops []*crawler.ShopItem
    mgominer.Find(bson.M{"state": "posted"}).Sort("-date").Limit(10).All(&shops)
    log.Infof("shop length is %d", len(shops))

    for _, shopitem := range shops {
        var allowchan chan bool = make(chan bool, threadnum)
        log.Infof("start to fetch %d", shopitem.Shop_id)
        shoptype := TAOBAO
        shopid := strconv.Itoa(shopitem.Shop_id)

        items := shopitem.Items_list
        if len(items) == 0 {
            process_none_item_shop(mgominer, shopitem.Shop_id)
            continue
        }

        istmall, err := crawler.IsTmall(items[0])
        if err != nil {
            log.Errorf("判断是否为天猫商品失败， shopid is %d, err is %s", shopid, err.Error)
            process_judge_shoptype_err(mgominer, shopitem.Shop_id)
            continue
        }
        if istmall {
            shoptype = TMALL
        }

        var wg sync.WaitGroup
        for _, itemid := range items {
            allowchan <- true
            wg.Add(1)
            go fetch(itemid, shopid, shoptype, mgofailed, item_depot, mgopages, wg, allowchan)
        }
        wg.Wait()
        close(allowchan)
        updateshop(shopitem.Shop_id, mgominer, mgoshop)
    }
}

func updateshop(shopid int, mgominer, mgoshop *mgo.Collection) {
    err := mgominer.Update(bson.M{"shop_id": shopid},
        bson.M{"$set": bson.M{"state": "fetched",
            "date": time.Now()}})
    if err != nil {
        log.Errorf("update minerals state error, shopid is %d , err is %s",
            shopid, err.Error())
    }
    err = mgoshop.Update(bson.M{"shop_info.sid"})
}
func fetch(itemid, shopid, shoptype string,
    mgofailed, item_depot, mgopages *mgo.Collection,
    wg sync.WaitGroup,
    allowchan chan bool) {
    defer wg.Done()
    defer func() { <-allowchan }()
    font, detail, instock, err, isWeb := crawler.FetchItem(itemid, shoptype)
    if err != nil {
        log.Errorf("抓取页面失败，itemid 是 %s, 错误为 %s", itemid, err.Error())
        if instock {
            crawler.SaveFailed(itemid, shopid, shoptype, mgofailed)
        }
        return
    } else {
        if isWeb {
            info, err := crawler.ParseWeb(font, detail, itemid, shopid, shoptype)
            if err != nil {
                log.Errorf("解析web页面失败，itemid is %s, err is %s", itemid, err.Error())
                crawler.SaveFailed(itemid, shopid, shoptype, mgofailed)
                return
            }
            err = crawler.Save(info, item_depot)
            if err != nil {
                log.Errorf("保存解析结果失败，itemid is %s, err is %s", itemid, err.Error())
                crawler.SaveFailed(itemid, shopid, shoptype, mgofailed)
                return
            }
        } else {
            info, instock, err := crawler.ParsePage(font, detail, itemid, shopid, shoptype)
            if err != nil {
                if instock {
                    crawler.SaveSuccessed(itemid, shopid, shoptype,
                        font, detail, false, instock, mgopages)
                    return
                }
            } else {
                //保存解析结果到mongo
                err := crawler.Save(info, item_depot)
                parsed := false
                if err != nil {
                    log.Errorf("保存解析结果是出错，itemid is %s, err is %s",
                        itemid, err.Error())
                } else {
                    parsed = true
                }
                crawler.SaveSuccessed(itemid, shopid, shoptype,
                    font, detail, parsed, instock, mgopages)
            }
        }
    }
}

func process_judge_shoptype_err(mgominer *mgo.Collection, shopid int) {
    mgominer.Update(bson.M{"shop_id": shopid},
        bson.M{"$set": bson.M{"state": "fetched", "date": time.Now()}})
}

func process_none_item_shop(mgominer *mgo.Collection, shopid int) {
    log.Infof("shop %d has no items", shopid)
    err := mgominer.Update(bson.M{"shop_id": shopid}, bson.M{"$set": bson.M{"state": "fetched", "date": time.Now()}})
    if err != nil {
        log.Errorf("更新店铺失败，shopid is %d, err is %s", shopid, err.Error())
    }
}
