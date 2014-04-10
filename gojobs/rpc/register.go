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
    RegistedService["fetchnew"] = fetchnew
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
