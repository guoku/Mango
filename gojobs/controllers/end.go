package controllers

import (
    "fmt"
    "github.com/astaxie/beego"
    "net/rpc"
)

type EndController struct {
    beego.Controller
}

func (this *EndController) Post() {
    start, _ := this.GetInt("end")
    var result string
    if start == 2 {
        client, err := rpc.DialHTTP("tcp", "127.0.0.1:2301")
        if err != nil {
            panic(err)
        }

        err = client.Call("Watcher.GetInfo", 2, &result)
        if err != nil {
            panic(err)
        }
        fmt.Println(result)
    }
    this.Redirect("/", 302)
}
