package jobs

import "fmt"

type Job interface {
    Start(arg string, result *string) error
    Stop(arg string, result *string) error
    Statu(arg string, result *string) error
}

type Hooker interface {
    run()
}
type Base struct {
    start bool
    Hook  Hooker
}

func (this *Base) Start(arg string, result *string) error {
    fmt.Println("执行start方法")
    if arg == START {
        if this.start {
            *result = START_STATU
            return nil
        }
        *result = START_STATU
        this.start = true
        fmt.Printf("此时%s\n", this.start)
        go this.Hook.run()
    }
    return nil
}

func (this *Base) Stop(arg string, result *string) error {
    if arg == STOP {
        this.start = false
        *result = STOP_STATU
    }
    return nil
}

func (this *Base) Statu(arg string, result *string) error {
    if this.start {
        *result = START_STATU
    } else {
        *result = STOP_STATU
    }
    return nil
}

func (this *Base) run() {
    fmt.Println("run father run method")
}
