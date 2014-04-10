package jobs

import (
    "Mango/gojobs/crawler"
    "Mango/gojobs/log"
    "fmt"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
    "strconv"
    "time"
)

type FetchNew struct {
    start bool
}

func (f *FetchNew) Start(arg string, result *string) error {
    if arg == START {
        if f.start {
            *result = "已经启动"
            return nil
        }
        *result = "开始启动"
        f.start = true
        //这里rpc相当奇怪，根据http://golang.org/src/pkg/net/rpc/server.go
        //这个文件286行的方法，要求rpc方法有两个参数，第一个不能为指针
        //第二个必须为指针，必须有一个接收者（这里是FetchNew),必须有一个返回值
        //但是为什么Run也需要达到这个要求呢?否则就报错
        //看了源代码发现，所有public方法，都必须按照这个要求，但是私有方法则不需要
        go f.run()
    }

    return nil
}

func (f *FetchNew) Stop(arg string, result *string) error {
    if arg == STOP {
        f.start = false
        *result = "已经停止运行"
    }
    return nil
}

func (f *FetchNew) Statu(arg string, result *string) error {
    if f.start {
        *result = "已经启动"
    } else {
        *result = "已经停止"
    }
    return nil
}

func (f *FetchNew) run() error {

    defer func() {
        fmt.Println("\n\n running over \n\n")
        f.start = false
    }()
    for {
        if f.start == false {
            return nil
        }

        FetchTaobaoItem(THREADNUM)
    }
    return nil
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
    if len(shops) == 0 {
        fmt.Println("睡眠一个小时")
        time.Sleep(time.Hour * 1)
    }
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

        for _, itemid := range items {
            allowchan <- true
            go fetch(itemid, shopid, shoptype, mgofailed, item_depot, mgopages, allowchan)
        }
        close(allowchan)
        updateshop(shopitem.Shop_id, mgominer, shop_depot)
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
    err = mgoshop.Update(bson.M{"shop_info.sid": shopid}, bson.M{"$set": bson.M{"status": "finished"}})
    if err != nil {
        log.Errorf("update shop err, shop id is %d, err is %s", shopid, err.Error())
    }
}
func fetch(itemid, shopid, shoptype string,
    mgofailed, item_depot, mgopages *mgo.Collection,
    allowchan chan bool) {
    defer func() {
        <-allowchan
    }()
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
            SAdd(FETCHNEW, itemid)

        } else {
            info, instock, err := crawler.ParsePage(font, detail, itemid, shopid, shoptype)
            if err != nil {
                if instock {
                    crawler.SaveSuccessed(itemid, shopid, shoptype,
                        font, detail, false, instock, mgopages)
                    SAdd(FETCHNEW, itemid)
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
                SAdd(FETCHNEW, itemid)
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
