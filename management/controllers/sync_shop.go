package controllers

import (
	"Mango/management/models"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/qiniu/log"
	"labix.org/v2/mgo/bson"
	"strconv"
	"time"
)

type SyncShopController struct {
	beego.Controller
}

func (this *SyncShopController) Get() {
	c := this.Input().Get("count")
	count, err := strconv.Atoi(c)
	if err != nil || count == 0 {
		this.Ctx.WriteString("the count parameter is a wrong number or is 0")
		return
	}
	off := this.Input().Get("offset")
	log.Info(off)
	var offset int
	if off == "" {
		offset = 0
	} else {
		offset, err = strconv.Atoi(off)
		if err != nil {
			this.Ctx.WriteString(err.Error())
			return
		}
	}

	query := bson.M{}
	date := this.GetString("date")
	if date != "" {
		form := "2006010203"
		t, err := time.Parse(form, date)
		if err != nil {
			log.Error(err)
			this.Ctx.WriteString(err.Error())
			return
		}
		log.Info(t)
		query["shop_info.updated_time"] = bson.M{"$gte": t}
	}
	shops := make([]*models.ShopItem, count)
	mgoc := MgoSession.DB(MgoDbName).C("taobao_shop")
	err = mgoc.Find(query).Limit(count).Skip(offset).All(&shops)
	if err != nil {
		log.Error(err)
		this.Ctx.WriteString("mongo err," + err.Error())
		return
	}
	log.Infof("%+v", shops)
	this.Data["json"] = shops
	this.ServeJson()

}

func (this *SyncShopController) Post() {
	shop := new(models.ShopItem)
	json.Unmarshal(this.Ctx.Input.RequestBody, &shop)
	if shop == nil || shop.ShopInfo == nil {
		log.Info("the data posted is nil")
		this.Ctx.WriteString("data is wrong")
	}
	sid := shop.ShopInfo.Sid
	mgoc := MgoSession.DB(MgoDbName).C("taobao_shop")
	_, err := mgoc.Upsert(bson.M{"shop_info.sid": sid}, bson.M{"$set": shop})
	if err != nil {
		log.Error(err)
		this.Ctx.WriteString(err.Error())
	}
	this.Ctx.WriteString("success")
}
