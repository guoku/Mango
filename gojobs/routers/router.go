package routers

import (
    "Mango/gojobs/controllers"
    "github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
    beego.Router("/start", &controllers.StartController{})
    beego.Router("/end", &controllers.EndController{})
}
