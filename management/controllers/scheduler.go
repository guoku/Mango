package controllers

import (
	"Mango/management/models"
	"Mango/management/models/apiresponse"
	"Mango/management/taobaoclient"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/jason-zou/taobaosdk/rest"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"regexp"
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
	re := regexp.MustCompile("http://[A-Za-z0-9]+\\.(taobao|tmall)\\.com")
	shopurl := re.FindString(shopName)
	link := strings.Replace(shopurl, ".", ".m.", 1)
	fmt.Println(link)
	shopInfo, topErr := fetch(link)
	if topErr != nil {
		fmt.Println(topErr.Error())
		this.Redirect("/scheduler/list_shops", 302)
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
	result.CrawlerInfo = &models.CrawlerInfo{Priority: 10, Cycle: 720}
	result.ExtendedInfo = &models.TaobaoShopExtendedInfo{Type: "unknown", Orientational: false, CommissionRate: -1}
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
	c = MgoSession.DB(MgoDbName).C("taobao_items_depot")
	results := make([]models.TaobaoItemStd, 0)
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
	fmt.Println(this.Ctx.Input.RequestBody)

	item := models.TaobaoItemStd{}
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &item)
	if err != nil {
		fmt.Println(err)
		this.Data["json"] = map[string]string{"status": "Error"}
		this.ServeJson()
		return
	}
	fmt.Println(item)
	/*
	   numIid, err := this.GetInt("num_iid")
	   if err != nil {
	       this.Abort("404")
	       return
	   }
	   url := fmt.Sprintf("http://item.taobao.com/item.htm?id=%d", numIid)
	   title := this.GetString("title")
	   nick := this.GetString("nick")
	   desc := this.GetString("desc")
	   cid, _ := this.GetInt("cid")
	   price, _ := this.GetFloat("price")
	   city := this.GetString("city")
	   state := this.GetString("state")
	   promotionPrice, _ := this.GetFloat("promotion_price")
	   shopType := this.GetString("shop_type")
	   reviewsCount, _ := this.GetInt("reviews_count")
	   salesNum, _ := this.GetInt("sales_num")
	   propsStr := this.GetString("props")
	   propsArray := strings.Split(propsStr, ";")
	   props := make(map[string]string)
	   inStockFlag := this.GetString("instock")
	   inStock := inStockFlag == "1"
	   for _, v := range propsArray {
	       vs := strings.Split(v, ":")
	       props[vs[0]] = vs[1]
	   }
	   itemImgs := this.GetStrings("item_img")

	   ic := MgoSession.DB(MgoDbName).C("taobao_items_depot")
	   ic.Update(bson.M{"num_iid": int(numIid)},
	             bson.M{"$set" :
	                     bson.M {"detail_url": url,
	                             "title" : title,
	                             "nick" : nick,
	                             "desc" : desc,
	                             "cid" : cid,
	                             "price" : price,
	                             "location.city" : city,
	                             "location.state" : state,
	                             "promotion_price": promotionPrice,
	                             "shop_type" : shopType,
	                             "reviews_count" : reviewsCount,
	                             "monthly_sales_num" : salesNum,
	                             "props" : props,
	                             "item_imgs" : itemImgs,
	                             "in_stock" : inStock,
	                             "data_updated_time" : time.Now()}})
	*/
	ic := MgoSession.DB(MgoDbName).C("taobao_items_depot")
	ic.Upsert(bson.M{"num_iid": int(item.NumIid)},
		bson.M{"$set": bson.M{"detail_url": item.DetailUrl,
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
			"data_updated_time": time.Now()}})

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
	extended_info := &models.TaobaoShopExtendedInfo{Type: shop_type, Orientational: orientational, CommissionRate: float32(commission_rate)}
	crawler_info := &models.CrawlerInfo{Priority: int(priority), Cycle: int(cycle)}
	c := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
	err := c.Update(bson.M{"shop_info.sid": sid}, bson.M{"$set": bson.M{"extended_info": extended_info, "crawler_info": crawler_info}})
	if err != nil {
		this.Abort("404")
		return
	}
	this.Redirect(fmt.Sprintf("/scheduler/shop_detail/taobao/?sid=%d", sid), 302)
}
