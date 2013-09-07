package main

import (
    "errors"
    "fmt"
    "flag"
	"Mango/management/controllers"
	"Mango/management/models"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
    var env string
    flag.StringVar(&env, "env", "debug", "program environment")
    flag.Parse()
    if env != "prod" && env != "staging" && env != "debug" {
        panic(errors.New("Wrong Environment Flag Value. Should be 'debug', 'staging' or 'prod'"))
    }
    beego.AppConfigPath = fmt.Sprintf("conf/%s.conf", env)
    beego.ParseConfig()
    mysqlUser := beego.AppConfig.String("mysqluser")
    mysqlPass := beego.AppConfig.String("mysqlpass")
    mysqlProtocol := beego.AppConfig.String("mysqlprotocol")
    mysqlHost := beego.AppConfig.String("mysqlhost")
    mysqlPort := beego.AppConfig.String("mysqlport")
	orm.RegisterDriver("mysql", orm.DR_MySQL)
	orm.RegisterDataBase("default", "mysql", fmt.Sprintf("%s%s@%s(%s%s)/guokuer?charset=utf8", mysqlUser, mysqlPass, mysqlProtocol, mysqlHost, mysqlPort), 30)
	//orm.RegisterDataBase("default", "mysql", "root@unix(/tmp/mysql.sock)/guokuer?charset=utf8", 30)
	orm.RegisterModel(new(models.User), new(models.UserProfile))
	orm.RegisterModel(new(models.RegisterInvitation))
	orm.RegisterModel(new(models.Permission))
	orm.RegisterModel(new(models.PasswordInfo))
	orm.RegisterModel(new(models.PasswordPermission))

	orm.RunCommand()
    orm.Debug = true
	beego.UseHttps = true
	beego.CertFile = "server.crt"
	beego.KeyFile = "server.key"
	beego.SessionOn = true
	beego.SessionProvider = "redis"
	beego.SessionSavePath = beego.AppConfig.String("redispath") 
}

func main() {
	beego.Router("/", &controllers.IndexController{})
	beego.Router("/register", &controllers.RegisterController{})
	beego.Router("/list_users", &controllers.ListUsersController{})
	beego.Router("/login", &controllers.LoginController{})
	beego.Router("/logout", &controllers.LogoutController{})
    beego.Router("/invite", &controllers.InviteController{})
    beego.Router("/list_pass", &controllers.ListPassController{})
	beego.Run()
}
