package controllers

import (
    "fmt"
    "github.com/astaxie/beego"
    "net/rpc"
)

type StartController struct {
    beego.Controller
}

func (this *StartController) Post() {
    start, _ := this.GetInt("start")
    var result string
    if start == 1 {
        client, err := rpc.DialHTTP("tcp", "127.0.0.1:2301")
        if err != nil {
            panic(err)
        }

        err = client.Call("Watcher.GetInfo", 1, &result)
        if err != nil {
            panic(err)
        }
        fmt.Println(result)
    }
    this.Redirect("/", 302)
}
