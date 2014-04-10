package routers

import (
    "Mango/gojobs/controllers"
    "github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
    beego.Router("/switcher", &controllers.SwitcherController{})
    beego.Router("/detail", &controllers.DetailController{})
}
