package controllers

import (
	"fmt"
	"Mango/management/models"
	"Mango/management/models/apiresponse"
	"Mango/management/taobaoclient"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/jason-zou/taobaosdk/rest"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strconv"
	"strings"
	"sync"
	"time"
)

var MgoSession *mgo.Session
var MgoDbName string
var shopLock sync.Mutex
var itemLock sync.Mutex
var Priorities = [10]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
var TaobaoShopTypes = [3]string{"unknown", "tmall", "global"}
const SchedulerCodeName = "manage_crawler"
const NumInOnePage = 50

type SchedulerController struct {
	UserSessionController
}

func (this *SchedulerController) Prepare() {
	this.UserSessionController.Prepare()
	user := this.Data["User"].(*models.User)
	this.Data["Tab"] = &models.Tab{TabName: "Scheduler"}
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
	nick := this.GetString("nick")
	query := bson.M{}
	if nick != "" {
		query["shop_info.nick"] = nick
	}
	results := make([]models.ShopItem, 0)
	err = c.Find(query).Sort("-created_time").Skip(int((page - 1) * NumInOnePage)).Limit(NumInOnePage).All(&results)
	if err != nil {
		this.Abort("500")
		return
	}
	total, _ := c.Find(query).Count()
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
        fmt.Println(topErr.Error())
		this.Redirect("/scheduler/list_shops", 301)
		return
	}
    fmt.Println("aaaa")
	addShopItem(shopInfo)
    fmt.Println("aaab")
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
    result.CrawlerInfo = &models.CrawlerInfo{Priority:10, Cycle:720}
    result.ExtendedInfo = &models.TaobaoShopExtendedInfo{Type : "unknown", Orientational : false, CommissionRate : -1}
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
	sid, err := this.GetInt("sid")
	if err != nil {
		this.Abort("404")
		return
	}
	c := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
	shop := models.ShopItem{}
	err = c.Find(bson.M{"shop_info.sid": sid}).One(&shop)
	if err != nil || shop.ShopInfo == nil {
		this.Abort("404")
		return
	}
	page, perr := this.GetInt("p")
	if perr != nil {
		page = 1
	}
	c = MgoSession.DB(MgoDbName).C("raw_taobao_items_depot")
	results := make([]models.TaobaoItem, 0)
	err = c.Find(bson.M{"sid": sid}).Skip(int((page - 1) * NumInOnePage)).Limit(NumInOnePage).All(&results)
	if err != nil {
		this.Abort("500")
		return
	}
	total, _ := c.Find(bson.M{"sid": sid}).Count()
	this.Data["Shop"] = shop
	this.Data["Paginator"] = models.NewSimplePaginator(int(page), total, NumInOnePage, this.Input())
	this.Data["ItemList"] = results
    this.Data["Priorities"] = Priorities
    this.Data["TaobaoShopTypes"] = TaobaoShopTypes
	this.Layout = DefaultLayoutFile
	this.TplNames = "taobao_shop_detail.tpl"
}

type TaobaoItemDetailController struct {
	SchedulerController
}

func (this *TaobaoItemDetailController) Get() {
	num_iid, err := this.GetInt("id")
	if err != nil {
		this.Abort("404")
		return
	}
	c := MgoSession.DB(MgoDbName).C("raw_taobao_items_depot")
	result := models.TaobaoItem{}
	err = c.Find(bson.M{"num_iid": num_iid}).One(&result)
	if err != nil {
		this.Abort("500")
		return
	}
	this.Data["Item"] = result
	this.Layout = DefaultLayoutFile
	this.TplNames = "taobao_item_detail.tpl"
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
		this.Data["json"] = map[string]string{"status": "no such shop"}
		this.ServeJson()
		return
	}
	addShopItem(shopInfo)
	this.Data["json"] = map[string]string{"status": "succeeded"}
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
	sid, _ := strconv.Atoi(this.GetString("sid"))
	itemsString := this.GetString("item_ids")
	itemIds := strings.Split(itemsString, ",")
    for _, v := range itemIds {
		numIid, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		addTaobaoItem(sid, numIid)
	}
    finish, _ := this.GetInt("finish")
    if finish == 1 {
        c := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
        c.Update(bson.M{"shop_info.sid": sid}, bson.M{"$set" : bson.M{"status" : "finished"}})
    }
	this.Data["json"] = map[string]string{"status": "succeeded"}
	this.ServeJson()
}

type GetShopFromQueueController struct {
	CrawlerApiController
}

func (this *GetShopFromQueueController) Get() {
	c := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
	result := models.ShopItem{}
	c.Find(bson.M{"status": "queued"}).Sort("crawler_info.priority", "last_crawled_time").One(&result)
	if result.ShopInfo == nil {
		this.Ctx.WriteString("0")
		return
	} else {
		change := bson.M{"$set": bson.M{"status": "crawling", "last_crawled_time": time.Now()}}
		c.Update(bson.M{"shop_info.sid": result.ShopInfo.Sid, "status": "queued"}, change)
		shopResult := apiresponse.GetShopResult{result.ShopInfo.Nick, result.ShopInfo.Sid, result.ShopInfo.Title}
		this.Data["json"] = &shopResult
	}
	this.ServeJson()

}

type UpdateTaobaoShopController struct {
    SchedulerController
}

func (this *UpdateTaobaoShopController) Post() {
    sid, _ := this.GetInt("sid")
    priority, _ := this.GetInt("priority")
    cycle, _ := this.GetInt("cycle")
    shop_type := this.GetString("shoptype")
    orientational, _ := this.GetBool("orientational")
    commission_rate, _ := this.GetFloat("commission_rate")
    extended_info := &models.TaobaoShopExtendedInfo{Type : shop_type, Orientational : orientational, CommissionRate : float32(commission_rate)}
    crawler_info := &models.CrawlerInfo{Priority: int(priority), Cycle : int(cycle)}
	c := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
	err := c.Update(bson.M{"shop_info.sid": sid}, bson.M{"$set" : bson.M{"extended_info" : extended_info, "crawler_info" : crawler_info}})
	if err != nil {
		this.Abort("404")
		return
	}
    this.Redirect(fmt.Sprintf("/scheduler/shop_detail/taobao/?sid=%d", sid), 301)
}
