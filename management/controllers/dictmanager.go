package controllers

import (
	"Mango/management/models"
	"fmt"
	//"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type WordsController struct {
	UserSessionController
}

func (this *WordsController) Prepare() {
	this.UserSessionController.Prepare()
	user := this.Data["User"].(*models.User)
	this.Data["Tab"] = &models.Tab{TabName: "Words"}
	if !CheckPermission(user.Id, SchedulerCodeName) {
		this.Abort("401")
		return
	}
}

type DictManagerController struct {
	WordsController
}

func (this *DictManagerController) Get() {
	page, err := this.GetInt("p")
	if err != nil {
		page = 1
	}
	numOnePage := 300
	c := MgoSession.DB("words").C("dict_chi_eng")
	words := make([]models.DictWord, 0)

	c.Find(nil).Sort("-freq").Skip(int(page-1) * numOnePage).Limit(numOnePage).All(&words)
	total, _ := c.Find(nil).Count()
	this.Data["Paginator"] = models.NewSimplePaginator(int(page), total, numOnePage, this.Input())
	this.Data["Words"] = words
	this.Layout = DefaultLayoutFile
	this.TplNames = "dict_manage.tpl"
}

type DictUpdateController struct {
	WordsController
}

func (this *DictUpdateController) Post() {
	w := this.GetString("w")
	fmt.Println(w)
	blacklist, _ := this.GetBool("blacklist")
	c := MgoSession.DB("words").C("dict_chi_eng")
	if err := c.Update(bson.M{"word": w}, bson.M{"$set": bson.M{"blacklisted": blacklist}}); err != nil {
		fmt.Println(err)
		this.Ctx.WriteString("error")
	}
	if blacklist {
		this.Ctx.WriteString("0")
	} else {
		this.Ctx.WriteString("1")
	}
}

type DictDeleteController struct {
	WordsController
}

func (this *DictDeleteController) Post() {
	w := this.GetString("w")
	fmt.Println(w)
	toDelete, _ := this.GetBool("delete")
	c := MgoSession.DB("words").C("dict_chi_eng")
	if err := c.Update(bson.M{"word": w}, bson.M{"$set": bson.M{"deleted": toDelete}}); err != nil {
		fmt.Println(err)
		this.Ctx.WriteString("error")
	}
	if toDelete {
		this.Ctx.WriteString("0")
	} else {
		this.Ctx.WriteString("1")
	}
}
