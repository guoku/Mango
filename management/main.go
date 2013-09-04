package main

import (
	"Mango/management/controllers"
	"Mango/management/models"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	orm.RegisterDriver("mysql", orm.DR_MySQL)
	orm.RegisterDataBase("default", "mysql", "root:123456@tcp(localhost:3306)/guokuer?charset=utf8", 30)
	orm.RegisterModel(new(models.User), new(models.UserAdditional))
	orm.RegisterModel(new(models.RegisterInvitation))
	orm.RunCommand()
	beego.UseHttps = true
	beego.CertFile = "server.crt"
	beego.KeyFile = "server.key"
	beego.SessionOn = true
	beego.SessionGCMaxLifetime = 86400
	beego.SessionProvider = "redis"
	beego.SessionSavePath = "127.0.0.1:6379"
}

func main() {
	beego.Router("/", &controllers.IndexController{})
	beego.Router("/register", &controllers.RegisterController{})
	beego.Router("/list_users", &controllers.ListUsersController{})
	beego.Router("/login", &controllers.LoginController{})
	beego.Router("/logout", &controllers.LogoutController{})
	beego.Run()
}
