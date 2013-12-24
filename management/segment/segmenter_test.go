package segment

import (
	"fmt"
	"testing"
	"time"
	//    "strings"
	"Mango/management/models"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

func TestSegment(t *testing.T) {
	var seg GuokuSegmenter
	seg.LoadDictionary()
	sess, err := mgo.Dial("10.0.1.23")
	if err != nil {
		t.Fatal("mongo error")
	}
	startTime := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	c := sess.DB("mango").C("taobao_items_depot")
	items := make([]models.TaobaoItemStd, 0)
	c.Find(bson.M{"data_updated_time": bson.M{"$gt": startTime}}).Skip(1000).Limit(2000).All(&items)
	fmt.Println("len", len(items))
	for i := 0; i < len(items); i++ {
		fmt.Println("origin:", items[i].Title)
		fmt.Println("segment:", seg.Segment(items[i].Title))
	}
	fmt.Println("hello kitty I love you 宠物硅胶防滑高跟Bomll＆Unite宝路联合MIUI小米The body shop美体小铺")
	fmt.Println("segment:", seg.Segment("hello kitty I love you 宠物硅胶防滑高跟Bomll＆Unite宝路联合MIUI小米The body shop美体小铺Tiger/虎牌"))

}
