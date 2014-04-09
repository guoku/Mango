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
