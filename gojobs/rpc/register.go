package rpc

import (
    "Mango/gojobs/jobs"
    "fmt"
    "net"
    "net/http"
    "net/rpc"
)

var RegistedService map[string]jobs.Job = make(map[string]jobs.Job)

func init() {
    //所有的job都必须按照如下两句注册到RPC服务器里面去
    fetchnew := new(jobs.FetchNew)
    fetchnew.Hook = fetchnew
    RegistedService["fetchnew"] = fetchnew

    syncShop := new(jobs.SyncShop)
    syncShop.Hook = syncShop
    RegistedService["syncshop"] = syncShop

    statuUpdate := new(jobs.StatuUpdate)
    statuUpdate.Hook = statuUpdate
    RegistedService["statuupdate"] = statuUpdate

    syncNewItem := new(jobs.SyncNewItem)
    syncNewItem.Hook = syncNewItem
    RegistedService["syncnewitem"] = syncNewItem

    fetchfailed := new(jobs.Fetchfailed)
    fetchfailed.Hook = fetchfailed
    RegistedService["fetchfailed"] = fetchfailed

    syncRefreshItem := new(jobs.SyncRefreshItem)
    syncRefreshItem.Hook = syncRefreshItem
    RegistedService["syncrefreshitem"] = syncRefreshItem

    syncOnlineItem := new(jobs.SyncOnlineItem)
    syncOnlineItem.Hook = syncOnlineItem
    RegistedService["synconlineitem"] = syncOnlineItem

    syncSelection := new(jobs.SyncSelection)
    syncSelection.Hook = syncSelection
    RegistedService["syncselection"] = syncSelection

    for k, v := range RegistedService {
        rpc.RegisterName(k, v)
    }
    rpc.HandleHTTP()
    ls, err := net.Listen("tcp", ":2301")
    if err != nil {
        panic(err)
    }
    fmt.Println("正在监听2301端口")
    go http.Serve(ls, nil)
}
