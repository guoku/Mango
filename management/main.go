package main

import (
	"Mango/management/controllers"
	"Mango/management/models"
	//"errors"
	//"flag"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"labix.org/v2/mgo"
)

func init() {
	/*
		var env string
		flag.StringVar(&env, "env", "debug", "program environment")
		flag.Parse()
		if env != "prod" && env != "staging" && env != "debug" {
			panic(errors.New("Wrong Environment Flag Value. Should be 'debug', 'staging' or 'prod'"))
		}
		beego.AppConfigPath = fmt.Sprintf("conf/%s.conf", env)
	*/
	beego.ParseConfig()
	mysqlUser := beego.AppConfig.String("mysqluser")

	mysqlPass := beego.AppConfig.String("mysqlpass")
	mysqlProtocol := beego.AppConfig.String("mysqlprotocol")
	mysqlHost := beego.AppConfig.String("mysqlhost")
	mysqlPort := beego.AppConfig.String("mysqlport")
	mongoHost := beego.AppConfig.String("mongohost")
	orm.RegisterDriver("mysql", orm.DR_MySQL)
	orm.RegisterDataBase("default", "mysql", fmt.Sprintf("%s%s@%s(%s%s)/guokuer?charset=utf8", mysqlUser, mysqlPass, mysqlProtocol, mysqlHost, mysqlPort), 30)
	//orm.RegisterDataBase("default", "mysql", "root@unix(/tmp/mysql.sock)/guokuer?charset=utf8", 30)
	orm.RegisterModel(new(models.User), new(models.UserProfile))
	orm.RegisterModel(new(models.RegisterInvitation))
	orm.RegisterModel(new(models.Permission))
	orm.RegisterModel(new(models.PasswordInfo))
	orm.RegisterModel(new(models.PasswordPermission))
	orm.RegisterModel(new(models.MPKey))
	orm.RegisterModel(new(models.MPApiToken))
	orm.RunCommand()
	//orm.Debug = true
	beego.HttpTLS, _ = beego.AppConfig.Bool("usehttps")
	beego.HttpCertFile = "server.crt"
	beego.HttpKeyFile = "server.key"
	beego.SessionOn = true
	/*if beego.SessionProvider == "redis" {
		beego.SessionSavePath = beego.AppConfig.String("redispath")
	}*/
	//beego.UseFcgi = true
	if beego.SessionProvider == "redis" {
		beego.SessionSavePath = beego.AppConfig.String("redispath")
	}
	//beego.UseFcgi = true
	session, err := mgo.Dial(mongoHost)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	controllers.MgoSession = session
	controllers.MgoDbName = beego.AppConfig.String("mongodbname")
	controllers.OfflineMode, _ = beego.AppConfig.Bool("offline")
	fmt.Println(beego.SessionGCMaxLifetime)
}

func main() {
	beego.Router("/", &controllers.IndexController{})
	beego.Router("/register", &controllers.RegisterController{})
	beego.Router("/list_users", &controllers.ListUsersController{})
	beego.Router("/login", &controllers.LoginController{})
	beego.Router("/logout", &controllers.LogoutController{})
	beego.Router("/invite", &controllers.InviteController{})
	beego.Router("/list_pass", &controllers.ListPassController{})
	beego.Router("/admin_view", &controllers.AdminPermissionController{})
	beego.Router("/user_profile", &controllers.UserProfileController{})
	beego.Router("/add_pass", &controllers.AddPassController{})
	beego.Router("/edit_pass/:id([0-9]+)", &controllers.EditPassController{})
	beego.Router("/delete_pass/:id([0-9]+)", &controllers.DeletePassController{})
	beego.Router("/edit_pass_permission/:id([0-9]+)", &controllers.EditPassPermissionController{})
	beego.Router("/scheduler/list_shops", &controllers.ShopListController{})
	beego.Router("/scheduler/shop_detail/taobao/", &controllers.TaobaoShopDetailController{})
	beego.Router("/scheduler/update_taobaoshop_info", &controllers.UpdateTaobaoShopController{})
	beego.Router("/scheduler/add_shop", &controllers.AddShopController{})
	beego.Router("/scheduler/api/add_shop", &controllers.AddShopFromApiController{})
	beego.Router("/scheduler/api/get_shop_from_queue", &controllers.GetShopFromQueueController{})
	beego.Router("/scheduler/item_detail/taobao/", &controllers.TaobaoItemDetailController{})
	beego.Router("/sync/shop", &controllers.SyncShopController{})

    if controllers.OfflineMode {
        beego.Router("/dict_manage/", &controllers.DictManagerController{})
        beego.Router("/dict_manage/blacklist/", &controllers.BlacklistManager{})
        beego.Router("/dict_manage/blacklist/update/", &controllers.BlacklistUpdateController{})
        beego.Router("/dict_manage/blacklist/delete/", &controllers.BlacklistDeleteController{})
        beego.Router("/dict_manage/blacklist/add/", &controllers.BlacklistAddController{})
        beego.Router("/dict_manage/brands/", &controllers.BrandsManageController{})
        beego.Router("/dict_manage/brands/update/", &controllers.BrandsUpdateController{})
        beego.Router("/dict_manage/brands/add/", &controllers.BrandsAddController{})
        beego.Router("/dict_manage/brands/delete/", &controllers.BrandsDeleteController{})
        beego.Router("/commodity/category/", &controllers.CategoryController{})
        beego.Router("/commodity/add_online_items/", &controllers.CreateOnlineItemsController{})
        beego.Router("/commodity/category_manage/", &controllers.CategoryManageController{})
        beego.Router("/commodity/category_manage/add_taobao_category/", &controllers.AddMatchedCategoryController{})
        beego.Router("/scheduler/api/send_taobao_items", &controllers.SendItemsController{})
        beego.Router("/scheduler/api/send_item_detail", &controllers.SendItemDataController{})
    }
	beego.Run()
}
