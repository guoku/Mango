package main

import (
    _ "Mango/gojobs/routers"
    _ "Mango/gojobs/rpc"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/orm"
)

func main() {
    orm.RunCommand()
    beego.Run()
}
