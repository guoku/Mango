package controllers

import (
	"fmt"
	//"strconv"
	//"time"

	"Mango/management/models"
	//"Mango/management/utils"

	//"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	//"github.com/astaxie/beego/validation"
	//_ "github.com/go-sql-driver/mysql"
)

type ListPassController struct {
	UserSessionController
}

func (this *ListPassController) Get() {
	user := this.Data["User"].(*models.User)
	p := models.PasswordPermission{}
	o := orm.NewOrm()
	o.QueryTable(&p).Filter("User", user.Id).All(&user.PasswordPermissions)
	for _, v := range user.PasswordPermissions {
		o.Read(v.Password)
	}

	per := models.Permission{}
	o.QueryTable(&per).Filter("Users__Id", user.Id).All(&user.Permissions)
	for _, v := range this.Data["User"].(*models.User).PasswordPermissions {
		fmt.Println(v.Password.Password)
	}
	this.Layout = "layout.html"
	this.TplNames = "list_pass.tpl"
}

type AddPassController struct {
	UserSessionController
}

func (this *AddPassController) Get() {
}

func (this *AddPassController) Post() {
}

type EditPassController struct {
	UserSessionController
}

func (this *EditPassController) Get() {
}

func (this *EditPassController) Post() {
}

type DeletePassController struct {
	UserSessionController
}

func (this *DeletePassController) Get() {
}

func (this *DeletePassController) Post() {
}
