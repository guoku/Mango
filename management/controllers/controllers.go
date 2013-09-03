package controllers

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"strconv"
	"time"

	"Mango/management/models"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
	_ "github.com/go-sql-driver/mysql"
)

type IndexController struct {
	UserSessionController
}

func (this *IndexController) Get() {
	this.Redirect("/list_users", 302)
}

type RegisterForm struct {
	Token      string `form:"token"`
	Password   string `form:"password" valid:"MinSize(8)"`
	Name       string `form:"name" valid:"Required"`
	Nickname   string `form:"nickname"`
	Mobile     string `form:"mobile" valid:"Required;Mobile"`
	Department string `form:"department"`
}

type RegisterController struct {
	beego.Controller
}

func (this *RegisterController) Get() {
	token := this.GetString("token")
	invitation := models.RegisterInvitation{}
	o := orm.NewOrm()
	err := o.QueryTable(&invitation).Filter("token", token).One(&invitation)
	if err == orm.ErrNoRows || err == orm.ErrMissPK {
		fmt.Println(err.Error())
		this.Ctx.WriteString("Token doesn't exist")
	} else if invitation.Expired {
		this.Ctx.WriteString("Token expired")
	} else if time.Since(invitation.IssueDate) > time.Duration(time.Hour*24) {
		invitation.Expired = true
		o.Update(&invitation)
		this.Ctx.WriteString("Token expired")
	} else {
		beego.ReadFromRequest(&this.Controller)
		this.Data["Invitation"] = invitation
		this.Data["Title"] = "Registration"
		this.Layout = "layout.html"
		this.TplNames = "register.tpl"
	}
}

func (this *RegisterController) Post() {
	flash := beego.NewFlash()
	rForm := RegisterForm{}
	if err := this.ParseForm(&rForm); err != nil {
		this.Ctx.WriteString("Error!")
		return
	}
	valid := validation.Validation{}
	b, vErr := valid.Valid(rForm)
	if vErr != nil {
		this.Ctx.WriteString("Error!")
		return
	}
	if !b {
		for _, e := range valid.Errors {
			flash.Error(fmt.Sprintf("%s %s", e.Key, e.Message))
			flash.Store(&this.Controller)
			this.Redirect("/register?token="+rForm.Token, 302)
			break
		}
		return
	}
	invitation := models.RegisterInvitation{}
	o := orm.NewOrm()
	err := o.QueryTable(&invitation).Filter("token", rForm.Token).One(&invitation)
	if err == orm.ErrNoRows || err == orm.ErrMissPK {
		fmt.Println(err.Error())
		this.Ctx.WriteString("Token doesn't exist")
	} else if invitation.Expired {
		this.Ctx.WriteString("Token expired")
	} else if time.Since(invitation.IssueDate) > time.Duration(time.Hour*24) {
		invitation.Expired = true
		o.Update(&invitation)
		this.Ctx.WriteString("Token expired")
	} else {
		user := models.User{}
		addtional := models.UserAdditional{}
		user.Email = invitation.Email
		salt := GenerateSalt(user.Email)
		user.Password = EncryptPassword(rForm.Password, salt)
		user.Name = rForm.Name
		user.Nickname = rForm.Nickname
		o.Insert(&user)
		addtional.Salt = salt
		addtional.Department = rForm.Department
		addtional.Mobile = rForm.Mobile
		addtional.User = &user
		o.Insert(&addtional)
		invitation.Expired = true
		o.Update(&invitation)
		this.Redirect("/list_users", 302)
	}
}

func GetMd5Digest(seed string) string {
	h := md5.New()
	h.Write([]byte(seed))
	return fmt.Sprintf("%x", h.Sum([]byte("")))
}

func GetSha1Digest(seed string) string {
	h := sha1.New()
	h.Write([]byte(seed))
	return fmt.Sprintf("%x", h.Sum([]byte("")))
}

func GenerateSalt(seed string) string {
	return GetMd5Digest(time.Now().String() + seed)
}

func EncryptPassword(origin, salt string) string {
	return GetSha1Digest(salt + GetMd5Digest(origin) + salt)
}

type ListUsersController struct {
	UserSessionController
}

func (this *ListUsersController) Get() {
	var users []*models.User
	user := models.User{}
	o := orm.NewOrm()
	o.QueryTable(&user).Limit(20).All(&users)
	this.Ctx.WriteString(users[0].Email)
}

type LoginController struct {
	beego.Controller
}

func (this *LoginController) Get() {
	v := this.GetSession("user_id")
	if v != nil {
		this.Redirect("/list_users", 302)
		return
	}
	this.Layout = "layout.html"
	this.TplNames = "login.tpl"
}

func (this *LoginController) Post() {
	email := this.GetString("email")
	password := this.GetString("password")
	user := models.User{}
	additional := models.UserAdditional{}
	o := orm.NewOrm()
	o.QueryTable(&user).Filter("email", email).One(&user)
	if user.Id != 0 {
		o.QueryTable(&additional).Filter("user_id", user.Id).One(&additional)
		if user.Password == EncryptPassword(password, additional.Salt) {
			this.SetSession("user_id", int(user.Id))
			this.Redirect("/list_users", 302)
			return
		}
	} else {
		this.Redirect("/login", 302)
		return
	}

}

type UserSessionController struct {
	beego.Controller
}

func (this *UserSessionController) Prepare() {
	v := this.GetSession("user_id")
	if v == nil {
		this.Redirect("/login", 302)
	}
	this.Data["UserId"], _ = strconv.Atoi(string(v.([]byte)))
	fmt.Println("user_id", this.Data["UserId"])
}

type LogoutController struct {
	beego.Controller
}

func (this *LogoutController) Get() {
	this.DestroySession()
	this.Redirect("/login", 302)
}
