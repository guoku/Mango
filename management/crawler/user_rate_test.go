package crawler

import (
	"Mango/management/models"
	"fmt"
	"github.com/qiniu/log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"testing"
	"time"
)

func TestGetUserid(t *testing.T) {
	t.SkipNow()
	userId, err := GetUserid("http://shop70766218.taobao.com/")
	if err != nil {
		t.Fatal(err)
	}
	log.Info(userId)
}

func TestParse(t *testing.T) {
	t.SkipNow()
	userid, err := GetUserid("http://dumex.tmall.com/?spm=a1z0b.7.w3-18454208515.1.EWtFUW")
	if err != nil {
		t.Fatal(err)
	}
	Parse(userid)
}
func TestParseTaobao(t *testing.T) {
	t.SkipNow()
	userid, err := GetUserid("http://shop71839143.taobao.com/")
	if err != nil {
		t.Fatal(err)
	}
	Parse(userid)
}

func TestGetInfo(t *testing.T) {
	t.SkipNow()
	//GetInfo("http://shop71839143.taobao.com/")
	GetInfo("http://dumex.tmall.com/")
}

func TestFetch(t *testing.T) {
	mgosession, err := mgo.Dial("127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	c := mgosession.DB("mango").C("taobao_shops_depot")
	shopItem := models.ShopItem{}

	c.Find(nil).One(&shopItem)
	sid := shopItem.ShopInfo.Sid
	if sid == 0 {
		log.Info("shibai")
		return
	}
	log.Info(sid)

	link := fmt.Sprintf("http://shop%d.taobao.com", sid)
	shop, err := Fetch(link)
	if err != nil {
		log.Fatal(err)
	}
	/*
		err = c.Remove(bson.M{"shop_info.sid": sid})
		if err != nil {
			log.Error(err)
			t.Fatal(err)
		}
	*/
	shopItem.ShopInfo = shop
	shopItem.LastUpdatedTime = time.Now()
	shopItem.ExtendedInfo.Type = shop.ShopType
	err = c.Update(bson.M{"shop_info.sid": sid}, bson.M{"$set": shopItem})
	if err != nil {
		log.Error(err)
		log.Fatal(err)
	}
}
