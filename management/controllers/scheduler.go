package controllers

import (
    "fmt"
    "strconv"
    "strings"
    "sync"
    "time"
    "Mango/management/models"
    "Mango/management/models/apiresponse"
    "Mango/management/taobaoclient"
    "github.com/astaxie/beego"
    "github.com/jason-zou/taobaosdk/rest"
    "github.com/astaxie/beego/orm"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var MgoSession *mgo.Session
var MgoDbName string
var shopLock sync.Mutex
var itemLock sync.Mutex

const SchedulerCodeName = "manage_crawler"
const NumInOnePage = 100
type SchedulerController struct {
    UserSessionController
}

func (this *SchedulerController) Prepare() {
    this.UserSessionController.Prepare()
    user := this.Data["User"].(*models.User)
    this.Data["Tab"] = &models.Tab{TabName : "Scheduler"}
    if !CheckPermission(user.Id, SchedulerCodeName) {
        this.Abort("401")
        return
    }
}

type ShopListController struct {
    SchedulerController
}


func (this *ShopListController) Get() {
    c := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
    page, err := this.GetInt("p")
    if err != nil {
        page = 1
    }
    results := make([]models.ShopItem, 0)
    err = c.Find(bson.M{}).Skip(int((page - 1) * NumInOnePage)).Limit(NumInOnePage).All(&results)
    if err != nil {
        this.Abort("500")
        return
    }
    total, _  := c.Find(bson.M{}).Count()
    paginator := models.NewSimplePaginator(int(page), total, NumInOnePage, this.Input())
    this.Data["ShopList"] = results
    this.Data["Paginator"] = paginator
    this.Layout = DefaultLayoutFile
    this.TplNames = "list_shop.tpl"
}

type AddShopController struct {
    SchedulerController
}

func (this *AddShopController) Post() {
    shopName := this.GetString("shop_name")
    shopInfo, topErr := taobaoclient.GetTaobaoShopInfo(shopName)
    if topErr != nil {
        this.Redirect("/scheduler/list_shops", 301)
        return
    }
    addShopItem(shopInfo)
    this.Redirect("/scheduler/list_shops", 302)
}

func addShopItem(shopInfo *rest.Shop) bool {
    shopLock.Lock()
    defer shopLock.Unlock()
    result := models.ShopItem{}
    c := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
    c.Find(bson.M{"shop_info.sid": shopInfo.Sid}).One(&result)
    if result.ShopInfo != nil {
        return false
    }
    result.ShopInfo = shopInfo
    result.CreatedTime = time.Now()
    result.LastUpdatedTime = time.Now()
    result.Status = "queued"
    err := c.Insert(&result)
    if err != nil {
        return false
    }
    return true
}

type TaobaoShopDetailController struct {
    SchedulerController
}


func (this *TaobaoShopDetailController) Get() {
    sid, err := strconv.Atoi(this.Ctx.Params[":sid"])
    fmt.Println("xxxxxxxxxxxxxxxxxxx")
    if err != nil {
        fmt.Println(err, "=======xxxxxxxxxxxxx")
        this.Abort("404")
        return
    }
    page, perr := this.GetInt("p")
    if perr != nil {
        page = 1
    }
    c := MgoSession.DB(MgoDbName).C("raw_taobao_items_depot")
    results := make([]models.TaobaoItem, 0)
    err = c.Find(bson.M{"sid" : sid}).Skip(int((page - 1) * NumInOnePage)).Limit(NumInOnePage).All(&results)
    if err != nil {
        this.Abort("500")
        return
    }
    total, _ := c.Find(bson.M{"sid" : sid}).Count()
    fmt.Println("total", total)
    this.Data["Paginator"] = models.NewSimplePaginator(int(page), total, NumInOnePage, this.Input())
    this.Data["ItemList"] = results
    this.Layout = DefaultLayoutFile
    this.TplNames = "taobao_shop_detail.tpl"
}

type CrawlerApiController struct {
    beego.Controller
}

func (this *CrawlerApiController) Prepare() {
    token := this.GetString("token")
    if !CheckToken(token) {
        this.Ctx.WriteString("not authorized")
        return
    }
}

func CheckToken(token string) bool {
    t := models.MPApiToken{}
    o := orm.NewOrm()
    o.QueryTable(&t).Filter("token", token).One(&t)
    if t.Id != 0 {
        return true
    }
    return false
}

type AddShopFromApiController struct {
    CrawlerApiController
}

func (this *AddShopFromApiController) Get() {
    shopName := this.GetString("shop_name")
    shopInfo, topErr := taobaoclient.GetTaobaoShopInfo(shopName)
    if topErr != nil {
        this.Data["json"] = map[string]string{"status" : "no such shop"}
        this.ServeJson()
        return
    }
    addShopItem(shopInfo)
    this.Data["json"] = map[string]string{"status" : "succeeded"}
    this.ServeJson()
}

func addTaobaoItem(sid, numIid int) bool {
    itemLock.Lock()
    defer itemLock.Unlock()
    taobaoItem := models.TaobaoItem{}
    c := MgoSession.DB(MgoDbName).C("raw_taobao_items_depot")
    c.Find(bson.M{"num_iid": numIid}).One(&taobaoItem)
    if taobaoItem.NumIid == 0 {
        taobaoItem.Sid = sid
        taobaoItem.NumIid = numIid
        taobaoItem.CreatedTime = time.Now()
        err := c.Insert(&taobaoItem)
        if err != nil {
            return false
        }
        return true
    }
    return false
}

type SendItemsController struct {
    CrawlerApiController
}

func (this *SendItemsController) Post() {
    sid, _:= strconv.Atoi(this.GetString("sid"))
    itemsString := this.GetString("item_ids")
    itemIds := strings.Split(itemsString, ",")
    for _, v := range itemIds {
        numIid, err := strconv.Atoi(v)
        if err != nil {
            continue
        }
        addTaobaoItem(sid, numIid)
    }
    this.Data["json"] = map[string]string{"status" : "succeeded"}
    this.ServeJson()
}

type GetShopFromQueueController struct {
    CrawlerApiController
}

func (this *GetShopFromQueueController) Get() {
    c := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
    result := models.ShopItem{}
    c.Find(bson.M{"status": "queued"}).One(&result)
    if result.ShopInfo == nil {
        this.Ctx.WriteString("0")
        return
    } else {
        change := bson.M{"$set": bson.M{"status": "crawling", "last_crawled_time" : time.Now()}}
        c.Update(bson.M{"shop_info.sid": result.ShopInfo.Sid, "status" : "queued"}, change)
        shopResult := apiresponse.GetShopResult{result.ShopInfo.Nick, result.ShopInfo.Sid, result.ShopInfo.Title}
        this.Data["json"] = &shopResult
    }
    this.ServeJson()

}

type ShopDetailController struct {
}

func (this *ShopDetailController) Get() {
}

func (this *ShopDetailController) Post() {
}



