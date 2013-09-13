package controllers

import (
    "sync"
    "time"
    "Mango/management/models"
    "Mango/management/models/apiresponse"
    "Mango/management/taobaoclient"
    "github.com/astaxie/beego"
    "github.com/jason-zou/taobaosdk/rest"
    //"github.com/astaxie/beego/orm"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var MgoSession *mgo.Session
var MgoDbName string
var lock sync.Mutex

const SchedulerCodeName = "manage_crawler"
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
    results := make([]models.ShopItem, 0)
    err := c.Find(bson.M{}).All(&results)
    if err != nil {
        this.Abort("500")
        return
    }
    this.Data["ShopList"] = results
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
    lock.Lock()
    defer lock.Unlock()
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
    return true
}

type AddShopFromApiController struct {
    CrawlerApiController
}

func (this *AddShopFromApiController) Post() {
}

type SendItemsController struct {
    CrawlerApiController
}

func (this *SendItemsController) Post() {
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

