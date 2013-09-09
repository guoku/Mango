package controllers

import (
	"fmt"
	"strconv"
	//"time"

	"Mango/management/models"
	"Mango/management/utils"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
	//_ "github.com/go-sql-driver/mysql"
)

const (
	NoPermission = iota
	CanRead
	CanUpdate
	CanManage
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
	key := utils.GetTheKey()
	for _, v := range this.Data["User"].(*models.User).PasswordPermissions {
		v.Password.Password = utils.DecryptStringInAES(v.Password.Password, key)
		v.Password.Account = utils.DecryptStringInAES(v.Password.Account, key)
	}
	this.Layout = DefaultLayoutFile
	this.TplNames = "list_pass.tpl"
}

type AddPassController struct {
	UserSessionController
}

func (this *AddPassController) Get() {
	beego.ReadFromRequest(&this.Controller)
	this.Layout = DefaultLayoutFile
	this.TplNames = "add_pass.tpl"
}

func (this *AddPassController) Post() {
	passwordInfo := models.PasswordInfo{}
	if err := this.ParseForm(&passwordInfo); err != nil {
		this.Ctx.WriteString("Error!")
		return
	}
	valid := validation.Validation{}
	b, vErr := valid.Valid(passwordInfo)
	if vErr != nil {
		this.Ctx.WriteString("Error!")
		return
	}
	if !b {
		for _, e := range valid.Errors {
			flash := beego.NewFlash()
			flash.Error(fmt.Sprintf("%s %s", e.Key, e.Message))
			flash.Store(&this.Controller)
			this.Redirect("/add_pass", 302)
		}
		return
	}
	o := orm.NewOrm()
	key := utils.GetTheKey()
	passwordInfo.Password = utils.EncryptStringInAES(passwordInfo.Password, key)
	passwordInfo.Account = utils.EncryptStringInAES(passwordInfo.Account, key)
	o.Insert(&passwordInfo)
	passwordPermission := models.PasswordPermission{}
	passwordPermission.User = this.Data["User"].(*models.User)
	passwordPermission.Password = &passwordInfo
	passwordPermission.Level = CanManage
	o.Insert(&passwordPermission)
	this.Redirect("/list_pass", 302)
}

type EditPassController struct {
	UserSessionController
}

type PasswordUser struct {
    User *models.User
    PermissionLevel int
}
func (this *EditPassController) Get() {
    passId, err := strconv.Atoi(this.Ctx.Params[":id"])
    if err != nil {
        this.Abort("404")
        return
    }
    passInfo := models.PasswordInfo{Id : passId}
    o := orm.NewOrm()
    err = o.Read(&passInfo)
    if err != nil {
        this.Abort("404")
        return
    }
    key := utils.GetTheKey()
	passInfo.Password = utils.DecryptStringInAES(passInfo.Password, key)
	passInfo.Account = utils.DecryptStringInAES(passInfo.Account, key)
    this.Data["Password"] = &passInfo
    var users []*models.User
    o.QueryTable("user").All(&users)
    pusers := make([]*PasswordUser, 0)
    permission := models.PasswordPermission{}
    for _, v := range users {
        o.QueryTable(&permission).Filter("User__Id", v.Id).Filter("Password__Id", passId).One(&permission)
        pusers = append(pusers, &PasswordUser{User:v, PermissionLevel:permission.Level})
    }
    this.Data["PassUsers"] = pusers
    this.Layout = DefaultLayoutFile
    this.TplNames = "edit_pass.tpl"

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
