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
		fmt.Println("not admin")
		this.Redirect("/list_users", 301)
	} 
    /*else {
		//admins have their own view which lists permission
		fmt.Println("is admin")
		this.Redirect("/admins_view", 301)
	}*/
}

type AdminPermissionController struct {
	UserSessionController
}

type UserData struct {
	Id    int
	Name  string
	Perms []*Permdata
}

func (this *AdminPermissionController) Post() {

	/*
		o := orm.NewOrm()
		var permissions []*models.Permission
		perm := models.Permission{}

		if rForm.Perm_password == true {

			p := models.Permission{}
			o.QueryTable(&perm).Filter("ContentTypeId", 1).One(&p)
			permissions = append(permissions, &p)
		}
		if rForm.Perm_crawler == true {
			p := models.Permission{}
			o.QueryTable(&perm).Filter("ContentTypeId", 2).One(&p)
			permissions = append(permissions, &p)
		}
		if rForm.Perm_product == true {
			p := models.Permission{}
			o.QueryTable(&perm).Filter("ContentTypeId", 3).One(&p)
			permissions = append(permissions, &p)
		}
		user := models.User{Id: rForm.Id}
	*/
	Id := this.Input().Get("Id")
	Uid, _ := strconv.Atoi(Id)
	Name := this.GetString("Name")
	user := models.User{Id: Uid, Name: Name}
	permission := models.Permission{}
	var permissions []*models.Permission
	var allpm []*models.Permission
	o := orm.NewOrm()
	o.QueryTable(&permission).All(&allpm)
	for _, p := range allpm {
		perm := this.GetString(p.Codename)
		if perm == "on" {
			pm := models.Permission{}
			o.QueryTable(&permission).Filter("codename", p.Codename).One(&pm)
			permissions = append(permissions, &pm)
		}
	}
	if o.Read(&user) == nil && len(permissions) > 0 {

		m2m := o.QueryM2M(&user, "Permissions")
		for _, temp := range permissions {
			if !m2m.Exist(temp) {

				m2m.Add(temp)
			}
		}

	}

	this.Redirect("/list_users", 302)
	return

}

type Permdata struct {
	PermName string
	Hold     bool
	Id       int
}

func (this *AdminPermissionController) Get() {
	id, _ := this.GetInt("id")
	var target models.User
	user := models.User{}
	o := orm.NewOrm()
	o.QueryTable(&user).Filter("id", id).One(&target)

	ud := new(UserData)
	ud.Id = target.Id
	ud.Name = target.Name
	permission := models.Permission{}
	var pm []*models.Permission
	o.QueryTable(&permission).Filter("Users__User__Id", target.Id).All(&pm)
	var allpm []*models.Permission
	o.QueryTable(&permission).All(&allpm)
	var allPermData []*Permdata
	for _, p := range allpm {
		pd := new(Permdata)
		pd.PermName = p.Codename
		pd.Id = p.ContentTypeId
		allPermData = append(allPermData, pd)
	}
	for _, p := range pm {
		for _, q := range allPermData {
			if p.Codename == q.PermName {
				q.Hold = true
			}
		}
	}
	ud.Perms = allPermData

	this.Data["UserData"] = &ud
	this.Data["Tab"] = &models.Tab{TabName: "Admin"}

	this.Layout = DefaultLayoutFile
	this.TplNames = "admin_view.tpl"
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
		user.IsActive = true
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

func (this *ListUsersController) Prepare() {

	this.UserSessionController.Prepare()
	user := this.Data["User"].(*models.User)
	this.Layout = DefaultLayoutFile

	if !user.IsAdmin {
		this.TplNames = "list_users.tpl"
	} else {
		this.TplNames = "admin_list_users.tpl"
	}
}

func (this *ListUsersController) Get() {

	var users []*models.User
	user := models.User{}
	o := orm.NewOrm()
	o.QueryTable(&user).All(&users)
	ad := models.UserProfile{}
	for _, v := range users {
		v.Profile = &models.UserProfile{}
		o.QueryTable(&ad).Filter("user_id", v.Id).One(v.Profile)
	}
	this.Data["Users"] = &users
	this.Data["Tab"] = &models.Tab{TabName: "Index"}

}

type UserProfileController struct {
	UserSessionController
}

/*
个人信息的展示和修改
*/
func (this *UserProfileController) Get() {
	this.UserSessionController.Prepare()
	user := this.Data["User"].(*models.User)
	o := orm.NewOrm()
	permission := models.Permission{}
	var pm []*models.Permission
	o.QueryTable(&permission).Filter("Users__User__Id", user.Id).All(&pm)

	user.Permissions = pm
	var profile models.UserProfile
	err := o.QueryTable("user_profile").Filter("User__Id", user.Id).One(&profile)
	if err == nil {
		user.Profile = &profile
	}
	option := "<option id=\"selected\" selected=\"selected\" >" + profile.Department + "</option>"
	//fmt.Print(option)
	this.Data["Option"] = option
	this.Data["User"] = &user
	this.Data["Tab"] = &models.Tab{TabName: "Profile"}

	this.Layout = DefaultLayoutFile
	this.TplNames = "user_profile.tpl"
}

type User_Profile struct {
	Name       string
	Nickname   string
	Mobile     string
	Department string
}

func (this *UserProfileController) Post() {
	//密码修改此处不涉及
	flash := beego.NewFlash()
	rForm := User_Profile{}
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
			this.Ctx.WriteString("have some err")
			//this.Redirect("/register?token="+rForm.Token, 302)
			break
		}
		return
	}

	o := orm.NewOrm()
	this.UserSessionController.Prepare()
	var user *models.User = this.Data["User"].(*models.User)

	var profile models.UserProfile
	o.QueryTable("user_profile").Filter("User__Id", user.Id).One(&profile)
	profile.Mobile = rForm.Mobile
	profile.Department = rForm.Department
	if o.Read(&profile) == nil {
		o.Update(&profile)
	}

	user.Name = rForm.Name
	user.Nickname = rForm.Nickname
	if o.Read(user) == nil {
		o.Update(user)

	}

	this.Redirect("/list_users", 302)
	return
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
			fmt.Println(user.IsAdmin)
			this.SetSession("is_admin", user.IsAdmin)
			this.Redirect("/list_users", 302)
			return
		}
	}
	this.Redirect("/login", 302)
	return

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
	m := utils.NewRegisterMail(email, token)
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
