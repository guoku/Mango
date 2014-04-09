package controllers

import (
    "Mango/gojobs/jobs"
    myrpc "Mango/gojobs/rpc"
    "fmt"
    "github.com/astaxie/beego"
    "net/rpc"
)

type SwitcherController struct {
    beego.Controller
}

func (this *SwitcherController) Post() {
    name := this.GetString("serviceName")
    action := this.GetString("actionName")
    var result string

    if _, ok := myrpc.RegistedService[name]; ok {
        rpcServer := beego.AppConfig.String("rpc::server")
        rpcPort := beego.AppConfig.String("rpc::port")
        client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%s", rpcServer, rpcPort))
        if err != nil {
            panic(err)
        }
        if action == "start" {
            err = client.Call(fmt.Sprintf("%s.%s", name, "Start"), jobs.START, &result)
        } else {
            err = client.Call(fmt.Sprintf("%s.%s", name, "Stop"), jobs.STOP, &result)
        }
        if err != nil {
            panic(err)
        }
        fmt.Println(result)
        this.Redirect("/", 302)
    } else {
        this.Redirect("/error", 303)
    }
}

func (this *SwitcherController) Get() {
    name := this.GetString("serviceName")
    rpcServer := beego.AppConfig.String("rpc::server")
    rpcPort := beego.AppConfig.String("rpc::port")
    client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%s", rpcServer, rpcPort))
    if err != nil {
        panic(err)
    }
    if _, ok := myrpc.RegistedService[name]; ok {
        var result string
        err = client.Call(fmt.Sprintf("%s.%s", name, "Start"), jobs.START, &result)
        if err != nil {
            panic(err)
        }
        if result == "已经启动" {
            this.Data["status"] = true

        } else {
            this.Data["status"] = false
        }
        this.Layout = "layout.html"
        this.TplNames = "index.tpl"
    } else {
        this.Redirect("/error", 303)
    }
}
