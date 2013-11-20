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

type PasswordController struct {
	UserSessionController
}

func (this *PasswordController) Prepare() {
	this.UserSessionController.Prepare()
	tab := &models.Tab{TabName: "Password"}
	this.Data["Tab"] = tab
}

type ListPassController struct {
	PasswordController
}

func (this *ListPassController) Get() {
	user := this.Data["User"].(*models.User)
	p := models.PasswordPermission{}
	o := orm.NewOrm()
	o.QueryTable(&p).Filter("User", user.Id).All(&user.PasswordPermissions)
	for _, v := range user.PasswordPermissions {
		o.Read(v.Password)
	}

	//per := models.Permission{}
	//o.QueryTable(&per).Filter("Users__Id", user.Id).All(&user.Permissions)
	key := utils.GetTheKey()
	for _, v := range this.Data["User"].(*models.User).PasswordPermissions {
		v.Password.Password = utils.DecryptStringInAES(v.Password.Password, key)
		v.Password.Account = utils.DecryptStringInAES(v.Password.Account, key)
	}
	this.Layout = DefaultLayoutFile
	this.TplNames = "list_pass.tpl"
}

type AddPassController struct {
	PasswordController
}

func (this *AddPassController) Get() {
	beego.ReadFromRequest(&this.Controller)
	this.Layout = DefaultLayoutFile
	this.TplNames = "add_pass.tpl"
}

func (this *AddPassController) Post() {
	passwordInfo := models.PasswordInfo{}
	if err := this.ParseForm(&passwordInfo); err != nil {
		this.Ctx.WriteString("Error!" + err.Error())
		return
	}
	valid := validation.Validation{}
	b, vErr := valid.Valid(passwordInfo)
	if vErr != nil {
		this.Ctx.WriteString("Error!" + vErr.Error())
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
	passwordPermission.Level = models.CanManage
	o.Insert(&passwordPermission)
	this.Redirect("/list_pass", 302)
}

type PasswordUser struct {
	User            *models.User
	PermissionLevel int
}

func (this *PasswordUser) CheckPermission(level int) bool {
	fmt.Print(this.PermissionLevel, level)
	if this.PermissionLevel == level {
		fmt.Println("true")
		return true
	}
	fmt.Println("false")
	return false
}

type PassSelectController struct {
	PasswordController
}

func (this *PassSelectController) Prepare() {
	this.UserSessionController.Prepare()
	passId, err := strconv.Atoi(this.Ctx.Input.Params(":id"))
	if err != nil {
		this.Abort("404")
		return
	}
	passInfo := models.PasswordInfo{Id: passId}
	o := orm.NewOrm()
	err = o.Read(&passInfo)
	if err != nil {
		this.Abort("404")
		return
	}
	this.Data["Password"] = &passInfo
}

type EditPassController struct {
	PassSelectController
}

func (this *EditPassController) Get() {
	passInfo := this.Data["Password"].(*models.PasswordInfo)
	user := this.Data["User"].(*models.User)
	permLevel := GetPassPermissionLevel(user.Id, passInfo.Id)
	if permLevel < models.CanUpdate {
		this.Abort("401")
		return
	}
	key := utils.GetTheKey()
	o := orm.NewOrm()
	passInfo.Password = utils.DecryptStringInAES(passInfo.Password, key)
	passInfo.Account = utils.DecryptStringInAES(passInfo.Account, key)
	this.Data["Password"] = &passInfo
	var users []*models.User
	o.QueryTable("user").All(&users)
	pusers := make([]*PasswordUser, 0)
	if permLevel == models.CanManage {
		this.Data["CanManage"] = true
		for _, v := range users {
			pusers = append(pusers, &PasswordUser{User: v,
				PermissionLevel: GetPassPermissionLevel(v.Id, passInfo.Id)})
		}
		this.Data["PassUsers"] = pusers
	}
	this.Layout = DefaultLayoutFile
	this.TplNames = "edit_pass.tpl"
}

func (this *EditPassController) Post() {
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
			this.Redirect("/edit_pass", 302)
		}
		return
	}
	passInfo := this.Data["Password"].(*models.PasswordInfo)
	passInfo.Name = passwordInfo.Name
	passInfo.Account = passwordInfo.Account
	passInfo.Password = passwordInfo.Password
	passInfo.Desc = passwordInfo.Desc
	o := orm.NewOrm()
	o.Update(passInfo)
	this.Redirect("/edit_pass", 302)
}

type DeletePassController struct {
	PassSelectController
}

func (this *DeletePassController) Get() {
	passInfo := this.Data["Password"].(*models.PasswordInfo)
	user := this.Data["User"].(*models.User)
	permLevel := GetPassPermissionLevel(user.Id, passInfo.Id)
	if permLevel < models.CanManage {
		this.Abort("401")
		return
	}
	o := orm.NewOrm()
	o.Delete(passInfo)
	this.Redirect("/list_pass", 302)
}

type EditPassPermissionController struct {
	PassSelectController
}

func (this *EditPassPermissionController) Post() {
	passInfo := this.Data["Password"].(*models.PasswordInfo)
	user := this.Data["User"].(*models.User)
	permLevel := GetPassPermissionLevel(user.Id, passInfo.Id)
	if permLevel < models.CanManage {
		this.Abort("401")
		return
	}
	editUserId, err := strconv.Atoi(this.GetString("edit_user_id"))
	if err != nil {
		this.Ctx.WriteString("Error!")
		return
	}
	editUser := &models.User{Id: editUserId}
	o := orm.NewOrm()
	err = o.Read(editUser)
	if err == orm.ErrNoRows || err == orm.ErrMissPK {
		this.Ctx.WriteString("Error!")
		return
	}
	level, lerr := strconv.Atoi(this.GetString("user_permissions"))
	if lerr != nil || level < models.NoPermission || level > models.CanManage {
		this.Ctx.WriteString("Error!")
		return
	}

	permission := &models.PasswordPermission{}
	o.QueryTable(permission).Filter("User__Id", editUserId).Filter("Password__Id", passInfo.Id).One(permission)
	permission.Level = level
	if permission.Id == 0 {
		permission.Password = passInfo
		permission.User = editUser
		o.Insert(permission)
	} else {
		o.Update(permission)
	}
	this.Redirect(fmt.Sprintf("/edit_pass/%d", passInfo.Id), 302)
}

func GetPassPermissionLevel(userId, passId int) int {
	o := orm.NewOrm()
	permission := &models.PasswordPermission{}
	o.QueryTable(permission).Filter("User__Id", userId).Filter("Password__Id", passId).One(permission)
	return permission.Level
}
