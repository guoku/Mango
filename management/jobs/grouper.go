package main


import (
    "time"
)

func main() {
    for {
        scanTaobaoItems()
        time.Sleep(1 * time.Minute)
    }
}
