package tmp

import (
    "Mango/gojobs/jobs"
    "fmt"
    "net"
    "net/http"
    "net/rpc"
    "time"
)

type Watcher struct {
    start bool
}

var C chan bool = make(chan bool)

func (w *Watcher) GetInfo(arg int, result *string) error {
    if arg == 1 {
        if w.start {
            *result = "已经启动了"
            return nil
        }
        *result = "开始启动"
        w.start = true
        go w.test()
    }

    if arg == 2 {
        w.start = false
        *result = "已经停止"
    }
    return nil
}

func (w *Watcher) test() {
    defer func() {
        w.start = false
    }()
    for {
        if w.start == false {
            return
        }
        fmt.Println("test run")
        time.Sleep(1 * time.Second)

    }
}
