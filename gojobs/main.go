package main

import (
    _ "Mango/gojobs/routers"
    _ "Mango/gojobs/rpc"
    "github.com/astaxie/beego"
)

func main() {
    beego.Run()
}
