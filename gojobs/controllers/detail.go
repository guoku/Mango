package controllers

import (
    "Mango/gojobs/jobs"
    "Mango/gojobs/models"
    "Mango/gojobs/rpc"
    "fmt"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/orm"
)

type DetailController struct {
    beego.Controller
}

func (this *DetailController) Get() {
    q := this.GetString("q")
    page, err := this.GetInt("p")
    if err != nil {
        page = 1
    }
    serviceName := this.GetString("serviceName")
    if serviceName == "" {
        for k, _ := range rpc.RegistedService {
            serviceName = k
            break
        }
        if serviceName == "" {
            this.Redirect("/", 302)
        }
    }
    if q != "" {
        this.filter(q, page, serviceName)
        return
    }
    info := new(ServiceInfo)
    info.Name = serviceName
    service := rpc.RegistedService[serviceName]
    var result string
    service.Statu("", &result)
    if result == "已经启动" {
        info.Statu = "started"
    } else {
        info.Statu = "stoped"
    }
    num := jobs.SCard(fmt.Sprintf("jobs:%s", serviceName))
    info.Count = int(num)
    numOnePage := 100
    o := orm.NewOrm()
    total, _ := o.QueryTable("crawler_logs").Count()
    this.Data["Paginator"] = models.NewSimplePaginator(int(page), int(total), numOnePage, this.Input())
    qs := o.QueryTable("crawler_logs").OrderBy("-time").Offset(numOnePage * int(page-1)).Limit(numOnePage)
    var clogs []*models.CrawlerLogs
    qs.All(&clogs)
    this.Data["Info"] = info
    this.Data["LogInfo"] = clogs
    this.Layout = "layout.html"
    this.TplNames = "detail.tpl"
    this.Data["Tab"] = &models.Tab{TabName: "Detail"}
}

func (this *DetailController) filter(q string, page int64, serviceName string) {
    info := new(ServiceInfo)
    info.Name = serviceName
    service := rpc.RegistedService[serviceName]
    var result string
    service.Statu("", &result)
    if result == "已经启动" {
        info.Statu = "started"
    } else {
        info.Statu = "stoped"
    }
    num := jobs.SCard(fmt.Sprintf("jobs:%s", serviceName))
    info.Count = int(num)
    numOnePage := 100
    o := orm.NewOrm()
    total, _ := o.QueryTable("crawler_logs").Filter("log_type", q).Count()
    this.Data["Paginator"] = models.NewSimplePaginator(int(page), int(total), numOnePage, this.Input())
    qs := o.QueryTable("crawler_logs").Filter("log_type", q).OrderBy("-time").Offset(numOnePage * int(page-1)).Limit(numOnePage)
    var clogs []*models.CrawlerLogs
    qs.All(&clogs)
    this.Data["Info"] = info
    this.Data["LogInfo"] = clogs
    this.Layout = "layout.html"
    this.TplNames = "detail.tpl"
    this.Data["Tab"] = &models.Tab{TabName: "Detail"}

}
