package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
)

func main() {
    file, err := os.Open("taobao.txt")
    if err != nil {
        panic(err)
    }
    session, err := mgo.Dial("10.0.1.23")
    if err != nil {
        panic(err)
    }
    c := session.DB("words").C("brands")
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        brands := strings.Split(line, "\t")
        for _, brand := range brands {
            info, _ := c.Upsert(bson.M{"name": brand}, bson.M{"$set" : bson.M{"valid" : true}})
            fmt.Println(brand, info)
        }
    }
}
