package controllers

import (
	"Mango/management/crawler"
	"Mango/management/models"
	"Mango/management/models/apiresponse"
	"Mango/management/taobaoclient"
	"Mango/management/utils"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/jason-zou/taobaosdk/rest"
	"github.com/qiniu/log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	//"regexp"
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
var TaobaoShopTypes = [3]string{"taobao.com", "tmall.com", "global"}
var Gifts = []string{"果库福利", "应用市场活动", "微博微信活动"}

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
		re := bson.RegEx{nick, "i"}
		brs := bson.M{"$regex": re}
		query["shop_info.nick"] = brs
	}
	sortOn := this.GetString("sorton")
	sortCon := "-created_time"
	if sortOn != "" {
		if sortOn == "priority" {
			sortCon = "crawler_info.priority"
		} else if sortOn == "status" {
			sortCon = "status"
		} else if sortOn != "created_time" {
			//根据礼品进行筛选
			sortCon = ""
		}
	}
	results := make([]models.ShopItem, 0)
	if sortCon != "" {
		err = c.Find(query).Sort(sortCon).Skip(int((page - 1) * NumInOnePage)).Limit(NumInOnePage).All(&results)
		if err != nil {
			this.Abort("500")
			return
		}
	} else {
		//根据礼品活动类型去查询店铺数据
		if sortOn == "commission" {
			query = bson.M{"extended_info.commission": true}
		} else {
			query = bson.M{"extended_info.gifts": sortOn}
		}
		err = c.Find(query).Skip(int((page - 1) * NumInOnePage)).Limit(NumInOnePage).All(&results)
		if err != nil {
			this.Abort("500")
			return
		}
	}
	total, _ := c.Find(query).Count()
	paginator := models.NewSimplePaginator(int(page), total, NumInOnePage, this.Input())
	this.Data["ShopList"] = results
	this.Data["Paginator"] = paginator
	this.Data["SortOn"] = sortOn
	this.Data["Gifts"] = Gifts
	this.Layout = DefaultLayoutFile
	this.TplNames = "list_shop.tpl"
}

type AddShopController struct {
	SchedulerController
}

func (this *AddShopController) Post() {
	shoplink := this.GetString("shop_name")
	//	re := regexp.MustCompile("http://[A-Za-z0-9]+\\.(taobao|tmall)\\.com")
	//	shopurl := re.FindString(shopName)
	//	link := strings.Replace(shopurl, ".", ".m.", 1)
	log.Info(shoplink)
	shopInfo, topErr := crawler.FetchShopDetail(shoplink)
	if topErr != nil {
		log.Info(topErr.Error())
		this.Redirect("/scheduler/list_shops", 302)
		return
	}
	addShopItem(shopInfo)
	this.Redirect("/scheduler/list_shops", 302)
}

//添加店铺
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
	result.ShopInfo.Synced = false
	result.CreatedTime = time.Now()
	result.LastUpdatedTime = time.Now()
	result.Status = "queued"
	result.CrawlerInfo = &models.CrawlerInfo{Priority: 10, Cycle: 720}
	result.ExtendedInfo = &models.TaobaoShopExtendedInfo{Type: strings.TrimSpace(shopInfo.ShopType), Orientational: false, CommissionRate: -1}
	err := c.Insert(&result)
	if err != nil {
		return false
	}
	return true
}

func updateShopItem(shopInfo *rest.Shop) bool {
	shopLock.Lock()
	defer shopLock.Unlock()
	result := models.ShopItem{}
	c := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
	c.Find(bson.M{"shop_info.sid": shopInfo.Sid}).One(&result)
	result.ShopInfo = shopInfo
	result.LastUpdatedTime = time.Now()
	if result.ExtendedInfo.Type == "taobao.com" || result.ExtendedInfo.Type == "unknown" {
		//防止把全球购的信息给覆盖了
		result.ExtendedInfo.Type = strings.TrimSpace(shopInfo.ShopType)
	}
	err := c.Update(bson.M{"shop_info.sid": shopInfo.Sid}, bson.M{"$set": result})
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

type TaobaoShopDetailController struct {
	SchedulerController
}

type GiftsWithStatu struct {
	Name string
	On   bool
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
	now := time.Now()
	if now.Sub(shop.LastUpdatedTime) > 30*time.Hour*24 {
		//一个月没有更新店铺信息，进行更新
		log.Info("start to update shop info")
		shoplink := fmt.Sprintf("http://shop%d.taobao.com", sid)
		shopinfo, err := crawler.FetchShopDetail(shoplink)
		if err != nil {
			log.Error(err)
		}
		log.Infof("%+v", shopinfo)
		updateShopItem(shopinfo)
	}
	page, perr := this.GetInt("p")
	if perr != nil {
		page = 1
	}
	c = MgoSession.DB(MgoDbName).C("taobao_items_depot")
	results := make([]models.TaobaoItemStd, 0)
	err = c.Find(bson.M{"sid": sid}).Skip(int((page - 1) * NumInOnePage)).Limit(NumInOnePage).All(&results)
	if err != nil {
		this.Abort("500")
		return
	}
	total, _ := c.Find(bson.M{"sid": sid}).Count()

	this.Data["Shop"] = shop
	gifts := shop.ExtendedInfo.Gifts
	var giftwithstatus []*GiftsWithStatu
	for _, gift := range Gifts {
		on := false
		for _, g := range gifts {
			if gift == g {
				on = true
			}
		}
		tmp := GiftsWithStatu{Name: gift, On: on}
		giftwithstatus = append(giftwithstatus, &tmp)
	}
	log.Info(shop.ExtendedInfo.Type)
	this.Data["Paginator"] = models.NewSimplePaginator(int(page), total, NumInOnePage, this.Input())
	this.Data["ItemList"] = results
	this.Data["Priorities"] = Priorities
	this.Data["TaobaoShopTypes"] = TaobaoShopTypes
	this.Data["Gifts"] = giftwithstatus
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
	c := MgoSession.DB(MgoDbName).C("taobao_items_depot")
	result := models.TaobaoItemStd{}
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
	taobaoItem := models.TaobaoItemStd{}
	c := MgoSession.DB(MgoDbName).C("taobao_items_depot")
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
		c.Update(bson.M{"shop_info.sid": sid}, bson.M{"$set": bson.M{"status": "finished"}})
		t := MgoSession.DB("test").C("status")
		t.Update(bson.M{"_id": 1}, bson.M{"$set": bson.M{"timestamp": time.Now()}})
	}
	this.Data["json"] = map[string]string{"status": "succeeded"}
	this.ServeJson()
}

type SendItemDataController struct {
	CrawlerApiController
}

func (this *SendItemDataController) Post() {
	log.Info(this.Ctx.Input.RequestBody)
	item := models.TaobaoItemStd{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &item)
	if err != nil {
		log.Info(err)
		this.Data["json"] = map[string]string{"status": "Data Error"}
		this.ServeJson()
		return
	}
	log.Info(item)
	session := utils.GetNewMongoSession()
	if session == nil {
		this.Data["json"] = map[string]string{"status": "DB Error"}
		this.ServeJson()
	}
	defer session.Close()
	ic := session.DB(MgoDbName).C("taobao_items_depot")
	tItem := models.TaobaoItemStd{}
	change := bson.M{
		"detail_url":        item.DetailUrl,
		"title":             item.Title,
		"nick":              item.Nick,
		"desc":              item.Desc,
		"sid":               item.Sid,
		"cid":               item.Cid,
		"price":             item.Price,
		"location":          item.Location,
		"promotion_price":   item.PromotionPrice,
		"shop_type":         item.ShopType,
		"reviews_count":     item.ReviewsCount,
		"monthly_sales_num": item.MonthlySalesVolume,
		"props":             item.Props,
		"item_imgs":         item.ItemImgs,
		"in_stock":          item.InStock,
	}
	ic.Find(bson.M{"num_iid": int(item.NumIid)}).One(&tItem)
	if tItem.Title == "" {
		t := time.Now()
		change["data_updated_time"] = t
		change["data_last_revised_time"] = t
	} else {
		change["data_last_revised_time"] = time.Now()
	}

	ic.Update(bson.M{"num_iid": int(item.NumIid)},
		bson.M{"$set": change})

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
	original, _ := this.GetBool("original")
	single_tail, _ := this.GetBool("singletail")
	var gifts []string
	for _, v := range Gifts {
		on := this.GetString(v)
		if on == "on" {
			log.Info(v)
			gifts = append(gifts, v)
		}
	}
	log.Info(gifts)
	commission, _ := this.GetBool("commission")
	main_products := this.GetString("main_products")
	log.Info(main_products)
	extended_info := &models.TaobaoShopExtendedInfo{Type: shop_type, Orientational: orientational, CommissionRate: float32(commission_rate),
		SingleTail: single_tail, Original: original, Gifts: gifts, Commission: commission}
	crawler_info := &models.CrawlerInfo{Priority: int(priority), Cycle: int(cycle)}
	c := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
	err := c.Update(bson.M{"shop_info.sid": sid}, bson.M{"$set": bson.M{"extended_info": extended_info, "crawler_info": crawler_info, "shop_info.main_products": main_products}})
	if err != nil {
		this.Abort("404")
		return
	}
	this.Redirect(fmt.Sprintf("/scheduler/shop_detail/taobao/?sid=%d", sid), 302)
}
