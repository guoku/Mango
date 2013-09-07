package controllers

import (
	"fmt"
	"strconv"
	"time"

	"Mango/management/models"
	"Mango/management/utils"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
	_ "github.com/go-sql-driver/mysql"
	"github.com/riobard/go-mailgun"
)

const (
	DefaultLayoutFile = "layout.html"
	MailgunKey        = "key-7n8gut3y8rpk1u-0edgmgaj7vs50gig8"
)

type UserSessionController struct {
	beego.Controller
}

func (this *UserSessionController) Prepare() {
	v := this.GetSession("user_id")
	if v == nil {
		this.Redirect("/login", 302)
		return
	}
	userId, _ := strconv.Atoi(string(v.([]byte)))
	user := models.User{Id: userId}
	o := orm.NewOrm()
	err := o.Read(&user)
	if err != nil {
		this.DestroySession()
		this.Redirect("/login", 302)
		return
	}
	this.Data["User"] = &user
}

type AdminSessionController struct {
	UserSessionController
}

func (this *AdminSessionController) Prepare() {
	this.UserSessionController.Prepare()
	user := this.Data["User"].(*models.User)
	if !user.IsAdmin {
		this.Redirect("/list_users", 301)
	}
}

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
		flash := beego.ReadFromRequest(&this.Controller)
		if _, ok := flash.Data["error"]; ok {
			this.Data["HasErrors"] = true
			fmt.Println("true")
		} else {
			this.Data["HasErrors"] = false
			fmt.Println("false")
		}
		this.Data["Invitation"] = invitation
		this.Data["Title"] = "Registration"
		this.Layout = DefaultLayoutFile
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
		profile := models.UserProfile{}
		user.Email = invitation.Email
		salt := utils.GenerateSalt(user.Email)
		user.Password = utils.EncryptPassword(rForm.Password, salt)
		user.Name = rForm.Name
		user.Nickname = rForm.Nickname
		o.Insert(&user)
		profile.Salt = salt
		profile.Department = rForm.Department
		profile.Mobile = rForm.Mobile
		profile.User = &user
		o.Insert(&profile)
		invitation.Expired = true
		o.Update(&invitation)
		user.Profile = &profile
		o.Update(&user)
		this.SetSession("user_id", int(user.Id))
		this.Redirect("/list_users", 302)
	}
}

type ListUsersController struct {
	UserSessionController
}

func (this *ListUsersController) Get() {
	var users []*models.User
	user := models.User{}
	o := orm.NewOrm()
	o.QueryTable(&user).Limit(20).All(&users)
	ad := models.UserProfile{}
	for _, v := range users {
		v.Profile = &models.UserProfile{}
		o.QueryTable(&ad).Filter("user_id", v.Id).One(v.Profile)
	}
	this.Data["Users"] = &users
	this.Layout = DefaultLayoutFile
	this.TplNames = "list_users.tpl"
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
	this.Layout = DefaultLayoutFile
	this.TplNames = "login.tpl"
}

func (this *LoginController) Post() {
	email := this.GetString("email")
	password := this.GetString("password")
	user := models.User{}
	additional := models.UserProfile{}
	o := orm.NewOrm()
	err := o.QueryTable(&user).Filter("email", email).One(&user)
	if err != nil {
	}
	if user.Id != 0 {
		fmt.Println("login", user.Id)
		o.QueryTable(&additional).Filter("user_id", user.Id).One(&additional)
		if user.Password == utils.EncryptPassword(password, additional.Salt) {
			this.SetSession("user_id", int(user.Id))
			this.Redirect("/list_users", 302)
			return
		}
	} else {
		this.Redirect("/login", 302)
		return
	}

}

type LogoutController struct {
	beego.Controller
}

func (this *LogoutController) Get() {
	this.DestroySession()
	this.Redirect("/login", 302)
}

type InviteController struct {
	AdminSessionController
}

func (this *InviteController) Get() {
	this.Layout = DefaultLayoutFile
	this.TplNames = "invite.tpl"
}

func (this *InviteController) Post() {
	email := this.GetString("email")
	valid := validation.Validation{}
	valid.Required(email, "email_empty")
	valid.Email(email, "email_invalid")
	if valid.HasErrors() {
		this.Ctx.WriteString("email invalid")
		return
	}
	token := utils.GenerateRegisterToken(email)
	m := models.NewRegisterMail(email, token)
	client := mailgun.New(MailgunKey)
	client.Send(m)
	invitation := models.RegisterInvitation{
		Token: token,
		Email: email,
	}
	o := orm.NewOrm()
	o.Insert(&invitation)
	this.Redirect("/invite", 302)
}
