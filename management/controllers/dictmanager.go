package controllers

import (
	"Mango/management/models"
	"fmt"
	//"labix.org/v2/mgo"
	"github.com/qiniu/log"
	"labix.org/v2/mgo/bson"
)

type WordsController struct {
	UserSessionController
}

func (this *WordsController) Prepare() {
	log.Info("prepare")
	this.Data["Tab"] = &models.Tab{TabName: "Words"}
	this.UserSessionController.Prepare()
	user := this.Data["User"].(*models.User)
	if !CheckPermission(user.Id, SchedulerCodeName) {
		this.Abort("401")
		return
	}
}

type DictManagerController struct {
	WordsController
}

func (this *DictManagerController) Get() {
	this.Redirect("/dict_manage/blacklist/", 302)
}

type BlacklistManager struct {
	WordsController
}

func (this *BlacklistManager) Get() {
	page, err := this.GetInt("p")
	if err != nil {
		page = 1
	}
	cond := bson.M{}
	q := this.GetString("q")
	if q != "" {
		cond["word"] = bson.M{"$regex": bson.RegEx{q, "i"}}
	}
	numOnePage := 200
	c := MgoSession.DB("words").C("dict_chi_eng")
	words := make([]models.DictWord, 0)
	c.Find(cond).Sort("-freq").Skip(int(page-1) * numOnePage).Limit(numOnePage).All(&words)
	total, _ := c.Find(nil).Count()
	this.Data["Paginator"] = models.NewSimplePaginator(int(page), total, numOnePage, this.Input())
	this.Data["Words"] = words
	this.Data["SearchQuery"] = q
	this.Data["DictTab"] = &models.Tab{TabName: "Blacklist"}
	this.Layout = DefaultLayoutFile
	this.TplNames = "dict_manage.tpl"
}

type BlacklistUpdateController struct {
	WordsController
}

func (this *BlacklistUpdateController) Post() {
	w := this.GetString("w")
	fmt.Println(w)
	blacklist, _ := this.GetBool("blacklist")
	c := MgoSession.DB("words").C("dict_chi_eng")
	if err := c.Update(bson.M{"word": w}, bson.M{"$set": bson.M{"blacklisted": blacklist}}); err != nil {
		fmt.Println(err)
		this.Data["json"] = map[string]bool{"error": true}
		this.ServeJson()
		return
	}
	word := models.DictWord{}
	c.Find(bson.M{"word": w}).One(&word)
	this.Data["json"] = map[string]bool{"blacklisted": word.Blacklisted,
		"deleted": word.Deleted, "error": false}
	this.ServeJson()
	/*
		    if blacklist {
				this.Ctx.WriteString("0")
			} else {
				this.Ctx.WriteString("1")
			}
	*/
}

type BlacklistDeleteController struct {
	WordsController
}

func (this *BlacklistDeleteController) Post() {
	w := this.GetString("w")
	fmt.Println(w)
	toDelete, _ := this.GetBool("delete")
	c := MgoSession.DB("words").C("dict_chi_eng")
	if err := c.Update(bson.M{"word": w}, bson.M{"$set": bson.M{"deleted": toDelete}}); err != nil {
		fmt.Println(err)
		this.Data["json"] = map[string]bool{"error": true}
		this.ServeJson()
		return
	}
	word := models.DictWord{}
	c.Find(bson.M{"word": w}).One(&word)
	this.Data["json"] = map[string]bool{"blacklisted": word.Blacklisted,
		"deleted": word.Deleted, "error": false}
	this.ServeJson()
}

type BlacklistAddController struct {
	WordsController
}

func (this *BlacklistAddController) Post() {
	w := this.GetString("w")
	fmt.Println(w)
	c := MgoSession.DB("words").C("dict_chi_eng")
	word := models.DictWord{}
	if err := c.Find(bson.M{"word": w}).One(&word); err != nil && err.Error() == "not found" {
		word.Word = w
		word.Type = "manual"
		e := c.Insert(&word)
		if e != nil {
			this.Ctx.WriteString("Error" + e.Error())
		} else {
			this.Ctx.WriteString("Success")
		}
	} else {
		this.Ctx.WriteString("Existed")
	}
}

type BrandsManageController struct {
	WordsController
}

func (this *BrandsManageController) Get() {
	page, err := this.GetInt("p")
	if err != nil {
		page = 1
	}
	cond := bson.M{}
	q := this.GetString("q")
	if q != "" {
		cond["name"] = bson.M{"$regex": bson.RegEx{q, "i"}}
	}
	numOnePage := 200
	c := MgoSession.DB("words").C("brands")
	words := make([]models.BrandsWord, 0)
	c.Find(cond).Sort("-freq").Skip(int(page-1) * numOnePage).Limit(numOnePage).All(&words)
	total, _ := c.Find(nil).Count()
	this.Data["Paginator"] = models.NewSimplePaginator(int(page), total, numOnePage, this.Input())
	this.Data["Words"] = words
	this.Data["SearchQuery"] = q
	this.Data["DictTab"] = &models.Tab{TabName: "Brands"}
	this.Layout = DefaultLayoutFile
	this.TplNames = "dict_brands.tpl"

}

type BrandsUpdateController struct {
	WordsController
}

func (this *BrandsUpdateController) Post() {
	w := this.GetString("w")
	valid, _ := this.GetBool("valid")
	c := MgoSession.DB("words").C("brands")
	err := c.Update(bson.M{"name": w}, bson.M{"$set": bson.M{"valid": valid}})
	if err != nil {
		this.Data["json"] = map[string]bool{"error": true}
		this.ServeJson()
		return
	}
	word := models.BrandsWord{}
	c.Find(bson.M{"name": w}).One(&word)
	this.Data["json"] = map[string]bool{"valid": word.Valid, "deleted": word.Deleted, "error": false}
	this.ServeJson()
}

type BrandsAddController struct {
	WordsController
}

func (this *BrandsAddController) Post() {
	w := this.GetString("w")
	c := MgoSession.DB("words").C("brands")
	word := models.BrandsWord{}
	err := c.Find(bson.M{"name": w}).One(&word)
	if err != nil && err.Error() == "not found" {
		word.Name = w
		word.Type = "manual"
		e := c.Insert(&word)
		if e != nil {
			this.Ctx.WriteString("Error" + e.Error())
		} else {
			this.Ctx.WriteString("Success")
		}
	} else {
		this.Ctx.WriteString("Existed")
	}
}

type BrandsDeleteController struct {
	WordsController
}

func (this *BrandsDeleteController) Post() {
	w := this.GetString("w")
	toDelete, _ := this.GetBool("delete")
	c := MgoSession.DB("words").C("brands")
	err := c.Update(bson.M{"name": w}, bson.M{"$set": bson.M{"deleted": toDelete, "valid": false}})
	if err != nil {
		this.Data["json"] = map[string]bool{"error": true}
		this.ServeJson()
		return
	}
	word := models.BrandsWord{}
	c.Find(bson.M{"name": w}).One(&word)
	this.Data["json"] = map[string]bool{"valid": word.Valid, "deleted": word.Deleted, "error": false}
	this.ServeJson()
}
