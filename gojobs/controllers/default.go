package controllers

import (
    "Mango/gojobs/models"
    "Mango/gojobs/rpc"
    "github.com/astaxie/beego"
)

type MainController struct {
    beego.Controller
}

type ServiceInfo struct {
    Name  string
    Statu string
    Count int
}

func (this *MainController) Get() {
    service := []*ServiceInfo{}
    for k, v := range rpc.RegistedService {
        info := new(ServiceInfo)
        info.Name = k
        var result string
        v.Statu("", &result)
        if result == "已经启动" {
            info.Statu = "started"
        } else {
            info.Statu = "stoped"
        }
        service = append(service, info)
    }
    this.Data["serviceInfo"] = service
    this.Data["Tab"] = &models.Tab{TabName: "Index"}
    this.Layout = "layout.html"
    this.TplNames = "index.tpl"
}
