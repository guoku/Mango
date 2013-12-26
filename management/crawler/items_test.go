package crawler

import (
	"Mango/management/utils"
	"fmt"
	"github.com/qiniu/log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strconv"
	"sync"
	"testing"
	"time"
)

const THREADSNUM int = 20
const (
	MGOHOST string = "10.0.1.23"
	MGODB   string = "zerg"
	TAOBAO  string = "taobao.com"
	TMALL   string = "tmall.com"
	MANGO   string = "mango"
)

var mgopages *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "pages")
var mgofailed *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "failed")
var mgominer *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "minerals")
var mgoMango *mgo.Collection = utils.MongoInit(MGOHOST, MANGO, "taobao_items_depot")

func TestFetchIItem(t *testing.T) {
	t.SkipNow()
	var allowchan chan bool = make(chan bool, THREADSNUM)
	log.Printf("\n\nStart to run fetch")
	shoptype := TAOBAO
	shopitem := new(utils.ShopItem)
	minerals := utils.MongoInit(MGOHOST, MGODB, "minerals")
	minerals.Find(bson.M{"state": "posted"}).Sort("-date").One(shopitem)
	shopid := strconv.Itoa(shopitem.Shop_id)
	items := shopitem.Items_list
	if len(items) == 0 {
		return
	}
	istmall, err := utils.IsTmall(items[0])
	if err != nil {
		log.Error(err)
		t.Fatal(err)
		return
	}
	if istmall {
		shoptype = TMALL
	}

	var wg sync.WaitGroup
	for _, itemid := range items {
		allowchan <- true
		wg.Add(1)
		go func(itemid string) {
			defer wg.Done()
			defer func() { <-allowchan }()
			font, detail, instock, err := FetchItem(itemid, shoptype)
			if err != nil {
				if instock {
					SaveFailed(itemid, shopid, shoptype, mgofailed)
				}
			} else {
				info, instock, err := ParsePage(font, detail, itemid, shopid, shoptype)
				if err != nil {
					if instock {
						SaveSuccessed(itemid, shopid, shoptype, font, detail, false, instock, mgopages)
					}
				} else {
					//保存解析结果到mongo
					err := utils.Save(info, mgoMango)
					fmt.Printf("%+v", info)
					parsed := false
					if err != nil {
						log.Error(err)
						parsed = false
					} else {
						parsed = true
					}
					SaveSuccessed(itemid, shopid, shoptype, font, detail, parsed, instock, mgopages)
				}
			}
		}(itemid)
	}
	wg.Wait()
	close(allowchan)
	sid, _ := strconv.Atoi(shopid)
	err = mgominer.Update(bson.M{"shop_id": sid}, bson.M{"$set": bson.M{"state": "fetched", "date": time.Now()}})
	if err != nil {
		log.Println("update minerals state error")
		log.Println(err.Error())
	}

}
