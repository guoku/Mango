package main

import (
    "fmt"
    "io"
    "os"
    "labix.org/v2/mgo"
)

type Word struct {
    Word string
    Freq int
}

func main() {
    session, err := mgo.Dial("10.0.1.23")
    if err != nil {
        panic(err)
    }
    words := make([]Word, 0)
    wc := session.DB("words").C("dict")
    wc.Find(nil).Sort("-freq").All(&words)
    f, _ := os.Create("words.txt")
    for _, v := range words {
        io.WriteString(f, fmt.Sprintf("%s\t%d\n", v.Word, v.Freq))
    }
    f.Close()
}
