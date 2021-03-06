package crawler

import (
	"Mango/management/models"
	"Mango/management/utils"
	"fmt"
	"github.com/qiniu/log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"testing"
	"time"
)

func TestGetUserid(t *testing.T) {
	t.SkipNow()
	userId, err := GetUserid("http://shop33634329.taobao.com")
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
	ParseShop(userid)

}
func TestParseTaobao(t *testing.T) {
	t.SkipNow()
	userid, err := GetUserid("http://shop104286230.taobao.com/")
	if err != nil {
		t.Fatal(err)
	}
	ParseShop(userid)
}

func TestGetInfo(t *testing.T) {
	t.SkipNow()
	//GetInfo("http://shop71839143.taobao.com/")
	GetShopInfo("http://dumex.tmall.com/")
}

func TestFetch(t *testing.T) {
	t.SkipNow()
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
	shop, err := FetchShopDetail(link)
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

func TestDecompress(t *testing.T) {
	t.SkipNow()
	mgopage := utils.MongoInit("10.0.1.23", "zerg", "pages")
	var pages *Pages
	mgopage.Find(bson.M{"itemid": "18634001513"}).One(&pages)
	font, _ := Decompress(pages.FontPage)
	log.Info(font)

}

func TestFetchWtype(t *testing.T) {
	t.SkipNow()
	font, _, _, _, err := FetchWithOutType("19864856561")
	if err != nil {
		log.Error(err)
		t.Fatal(err)
	}
	nick, err := GetShopNick(font)
	log.Info("nick is ", nick)
	if err != nil {
		log.Error(err)
		t.Fatal(err)
	}
	link, _ := GetShopLink(font)
	log.Info("link is ", link)
	sh, _ := FetchShopDetail(link)
	log.Infof("%+v", sh)
	shop := models.ShopItem{}
	shop.ShopInfo = sh
	log.Infof("%+v", shop.ShopInfo)
	log.Infof("%+v", shop.ShopInfo.ShopScore)
}
