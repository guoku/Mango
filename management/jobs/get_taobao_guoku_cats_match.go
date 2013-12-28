package main

import (
    "fmt"
    "io"
    "os"
    "Mango/management/models"
    "labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)
var MgoSession *mgo.Session
var MgoDbName string = "mango"
func init() {
    session, err := mgo.Dial("10.0.2.200")
    if err != nil {
        panic(err)
    }
    MgoSession = session
}

func generateMatchFile() {
    c := MgoSession.DB(MgoDbName).C("taobao_cats")
    cats := make([]models.TaobaoItemCat, 0)
    c.Find(bson.M{"matched_guoku_cid" : bson.M{"$gt" : 0}}).All(&cats)
    f, err := os.Create("cats_matching.txt")
    if err != nil {
        fmt.Println(err)
        return
    }
    for i := range cats {
        n, err := io.WriteString(f, fmt.Sprintf("%d\t%d\n", cats[i].ItemCat.Cid, cats[i].MatchedGuokuCid))
        if err != nil {
            fmt.Println(n, err)
        }
    }
    f.Close()
}

func main() {
    generateMatchFile()
}
