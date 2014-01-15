package crawler

import (
	"Mango/management/models"
	"Mango/management/utils"
	"encoding/json"
	"fmt"
	"github.com/qiniu/log"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"

	"testing"
	"time"
)

type Response struct {
	ItemId   string `json:"item_id"`
	TaobaoId string `json:"taobao_id"`
}

var mgoShop *mgo.Collection = utils.MongoInit(MGOHOST, MANGO, "taobao_shops_depot")

func TestFetchWithouttype(t *testing.T) {
	t.SkipNow()
	count := 1
	offset := 0
	resp, err := http.Get(fmt.Sprintf("http://114.113.154.47:8000/management/taobao/item/sync/?count=%d&offset=%d", count, offset))

	if err != nil {
		log.Error(err)
		time.Sleep(time.Minute)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		time.Sleep(time.Minute)
		return
	}

	r := make([]Response, 0)
	json.Unmarshal(body, &r)
	log.Infof("%+v", r)
	if len(r) == 0 {
		return
	}

	for _, v := range r {
		log.Info("taobao_id", v.TaobaoId)
		num_iid, _ := strconv.Atoi(v.TaobaoId)
		item := models.TaobaoItem{}
		err := mgoMango.Find(bson.M{"num_iid": int(num_iid)}).One(&item)
		if err != nil && err.Error() == "not found" {
			log.Error(err)
			font, detail, shoptype, _, err := FetchWithOutType(v.TaobaoId)
			if err != nil {
				log.Error(err)
				return
			}
			nick, err := GetShopNick(font)
			if err != nil {
				log.Error(err)
				return
			}
			shop := models.ShopItem{}
			err = mgoShop.Find(bson.M{"shop_info.nick": nick}).One(&shop)
			if err != nil && err.Error() == "not found" {
				log.Error(err)
				link, _ := GetShopLink(font)
				sh, _ := FetchShopDetail(link)
				shop.ShopInfo = sh
				shop.CreatedTime = time.Now()
				shop.LastUpdatedTime = time.Now()
				shop.Status = "queued"
				shop.CrawlerInfo = &models.CrawlerInfo{Priority: 10, Cycle: 720}
				shop.ExtendedInfo = &models.TaobaoShopExtendedInfo{Type: shoptype, Orientational: false, CommissionRate: -1}
				mgoShop.Insert(&shop)
			}
			log.Infof("%+v", shop)
			sid := strconv.Itoa(shop.ShopInfo.Sid)
			info, _, err := ParsePage(font, detail, v.TaobaoId, sid, shoptype)
			if err != nil {
				log.Error(err)
				return
			}
			log.Infof("%+v", info)
			Save(info, mgoMango)
		}

	}
}

func TestFetchWeb(t *testing.T) {
	//	FetchWeb("http://detail.tmall.com/item.htm?id=19864856561")
	FetchWeb("http://item.taobao.com/item.htm?id=35940684948")
}
