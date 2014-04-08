package main

import (
    _ "Mango/gojobs/routers"
    _ "Mango/gojobs/tmp"
    "fmt"
    "github.com/astaxie/beego"
    "net/rpc"
    "time"
)

func main() {
    client, err := rpc.DialHTTP("tcp", "127.0.0.1:2301")
    if err != nil {
        panic(err)
    }

    var result string
    err = client.Call("Watcher.GetInfo", 1, &result)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(result)
    time.Sleep(2 * time.Second)
    err = client.Call("Watcher.GetInfo", 2, &result)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(result)
    beego.Run()
}
