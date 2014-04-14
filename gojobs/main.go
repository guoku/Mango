package main

import (
    "Mango/gojobs/controllers"
    _ "Mango/gojobs/routers"
    _ "Mango/gojobs/rpc"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/orm"
)

func main() {
    orm.RunCommand()
    beego.Router("/", &controllers.MainController{})
    beego.Router("/switcher", &controllers.SwitcherController{})
    beego.Router("/detail", &controllers.DetailController{})
    beego.Run()
}
