package models

import (
    "fmt"
    "github.com/astaxie/beego"
    "github.com/astaxie/beego/orm"
    _ "github.com/go-sql-driver/mysql"
    "html/template"
    "net/url"
    "time"
)

func init() {
    orm.RegisterModel(new(CrawlerLogs))
    user := beego.AppConfig.String("log::mysqluser")
    pasw := beego.AppConfig.String("log::mysqlpass")
    sqlUrl := beego.AppConfig.String("log::mysqlurl")
    sqlPort := beego.AppConfig.String("log::mysqlport")
    db := beego.AppConfig.String("log::mysqldb")
    orm.RegisterDataBase("default", "mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", user, pasw, sqlUrl, sqlPort, db))

}

type CrawlerLogs struct {
    Id      int `orm:"auto;index"`
    Level   string
    LogType string
    File    string
    Line    int
    Time    time.Time `orm:"auto_now_add"`
    Reason  string    `orm:"type(text)"`
}
type SimplePaginator struct {
    HasPrev     bool
    HasNext     bool
    CurrentPage int
    TotalPages  int
    OtherParams template.URL
    PrevPage    int
    NextPage    int
}

func NewSimplePaginator(currentPage int, total int, numInOnePage int, params url.Values) *SimplePaginator {
    paginator := &SimplePaginator{}
    paginator.CurrentPage = currentPage
    paginator.TotalPages = total / numInOnePage
    params.Del("p")
    paginator.OtherParams = template.URL(params.Encode())
    if total%numInOnePage > 0 {
        paginator.TotalPages += 1
    }
    if paginator.CurrentPage > 1 {
        paginator.HasPrev = true
        paginator.PrevPage = paginator.CurrentPage - 1
    }
    if paginator.CurrentPage < paginator.TotalPages {
        paginator.HasNext = true
        paginator.NextPage = paginator.CurrentPage + 1
    }
    return paginator
}

type Tab struct {
    TabName string
}

func (this *Tab) IsIndex() bool {
    return this.TabName == "Index"
}

func (this *Tab) IsDetail() bool {
    return this.TabName == "Detail"
}
