package main

import (
	"Mango/management/crawler"
	"Mango/management/models"
	"Mango/management/utils"
	"encoding/json"
	"fmt"
	"github.com/qiniu/log"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"
)

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
var mgoShop *mgo.Collection = utils.MongoInit(MGOHOST, MANGO, "taobao_shops_depot")

type Response struct {
	ItemId   string `json:"item_id"`
	TaobaoId string `json:"taobao_id"`
}

/*
func syncOnlineItems() {
	count := 1000
	offset := 0
	ic := MgoSession.DB(MgoDbName).C("taobao_items_depot")
	sc := MgoSession.DB(MgoDbName).C("taobao_shops_depot")
	for {
		resp, err := http.Get(fmt.Sprintf("http://114.113.154.47:8000/management/taobao/item/sync/?count=%d&offset=%d", count, offset))
		//resp, err := http.Get(fmt.Sprintf("http://114.113.154.47:8000/management/taobao/item/sync/?count=%d&offset=%d", count, offset))
		if err != nil {
			time.Sleep(time.Minute)
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			time.Sleep(time.Minute)
			continue
		}
		r := make([]Response, 0)
		json.Unmarshal(body, &r)
		log.Info(r)
		if len(r) == 0 {
			break
		}
		allNew := true
		for _, v := range r {
			log.Info("taobao_id", v.TaobaoId)
			iid, _ := strconv.Atoi(v.TaobaoId)
			item := models.TaobaoItem{}
			err := ic.Find(bson.M{"num_iid": int(iid)}).One(&item)
			if err != nil && err.Error() == "not found" {
				//log.Info("not found", iid)
				ti, te := taobaoclient.GetTaobaoItemInfo(int(iid))
				if te != nil {
					log.Info("error", te.Error())
					continue
				}
				item.ApiData = ti
				item.ApiDataReady = true
				item.NumIid = int(iid)
				shop := models.ShopItem{}
				se := sc.Find(bson.M{"shop_info.nick": ti.Nick}).One(&shop)
				if se != nil {
					if se.Error() == "not found" {
						ts, te := taobaoclient.GetTaobaoShopInfo(ti.Nick)
						if te != nil {
							log.Info("shop error", te.Error())
							continue
						}
						shop.ShopInfo = ts
						shop.CreatedTime = time.Now()
						shop.LastUpdatedTime = time.Now()
						shop.Status = "queued"
						shop.CrawlerInfo = &models.CrawlerInfo{Priority: 10, Cycle: 720}
						shop.ExtendedInfo = &models.TaobaoShopExtendedInfo{Type: "unknown", Orientational: false, CommissionRate: -1}
						se = sc.Insert(&shop)
					} else {
						continue
					}
				}
				item.Sid = shop.ShopInfo.Sid
				item.CreatedTime = time.Now()
				item.ItemId = v.ItemId
				ic.Insert(&item)
				log.Info("insert", item.NumIid)
				continue
			}
			if item.ItemId != "" {
				allNew = false
				log.Info("already exists", item.NumIid)
				break
			} else {
				ic.Update(bson.M{"num_iid": int(iid)}, bson.M{"$set": bson.M{"item_id": v.ItemId}})
				log.Info("update", item.NumIid)
			}
		}
		if !allNew {
			break
		}
		offset += count
	}
}
*/

func syncOnlineItems() {
	count := 1000
	offset := 0
	for {
		resp, err := http.Get(fmt.Sprintf("http://114.113.154.47:8000/management/taobao/item/sync/?count=%d&offset=%d", count, offset))

		if err != nil {
			log.Error(err)
			time.Sleep(time.Minute)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			time.Sleep(time.Minute)
			continue
		}
		r := make([]Response, 0)
		json.Unmarshal(body, &r)
		if len(r) == 0 {
			break
		}
		allNew := true
		for _, v := range r {
			log.Info("taobao_id", v.TaobaoId)
			num_iid, _ := strconv.Atoi(v.TaobaoId)
			item := models.TaobaoItem{}
			err := mgoMango.Find(bson.M{"num_iid": int(num_iid)}).One(&item)
			if err != nil && err.Error() == "not found" {
				log.Error(err)
				font, detail, shoptype, _, err := crawler.FetchWithOutType(v.TaobaoId)
				if err != nil {
					log.Error(err)
					continue
				}
				nick, err := crawler.GetShopNick(font)
				if err != nil {
					log.Error(err)
					continue
				}
				shop := models.ShopItem{}
				err = mgoShop.Find(bson.M{"shop_info.nick": nick}).One(&shop)
				if err != nil && err.Error() == "not found" {
					log.Info("店铺不存在，开始抓取店铺信息")
					log.Error(err)
					link, err := crawler.GetShopLink(font)
					if err != nil {
						log.Error(err)
						continue
					}
					sh, err := crawler.FetchShopDetail(link)
					if err != nil {
						log.Error(err)
						continue
					}
					log.Infof("%+v", sh)
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
				info, _, err := crawler.ParsePage(font, detail, v.TaobaoId, sid, shoptype)
				if err != nil {
					log.Error(err)
					continue
				}
				info.GuokuItemid = v.ItemId
				log.Infof("%+v", info)
				crawler.Save(info, mgoMango)
			}
			if item.ItemId != "" {
				allNew = false
				log.Info("already exists", item.NumIid)
				break
			} else {
				log.Info(item.ItemId)
				mgoMango.Update(bson.M{"num_iid": int(num_iid)}, bson.M{"$set": bson.M{"item_id": v.ItemId}})

			}
		}
		if !allNew {
			break
		}
		offset += count

	}

}

type CreateItemsResp struct {
	ItemId   string `json:"item_id"`
	EntityId string `json:"entity_id"`
	Status   string `json:"status"`
}

func uploadOfflineItems() {
	log.Info("start to uploadOfflineItems")
	cc := utils.MongoInit(MGOHOST, MANGO, "taobao_cats")
	ic := mgoMango
	readyCats := make([]models.TaobaoItemCat, 0)
	cc.Find(bson.M{"matched_guoku_cid": bson.M{"$gt": 0}}).All(&readyCats)
	for _, v := range readyCats {
		log.Info("start", v.ItemCat.Cid)
		items := make([]models.TaobaoItemStd, 0)
		ic.Find(bson.M{"cid": v.ItemCat.Cid, "uploaded": false, "item_imgs.0": bson.M{"$exists": true}, "score": bson.M{"$gt": 2}}).All(&items)
		log.Info("items length:", len(items))
		for j := range items {
			log.Info("deal with ", items[j].NumIid)
			if items[j].Title == "" {
				continue
			}
			params := url.Values{}
			utils.GetUploadItemParams(&items[j], &params, v.MatchedGuokuCid)
			resp, err := http.PostForm("http://114.113.154.47:8000/management/entity/create/offline/", params)
			//resp, err := http.PostForm("http://114.113.154.47:8000/management/entity/create/offline/", params)
			log.Infof("%+v", params)
			if err != nil {
				continue
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err)
			}
			log.Info(resp.Status)
			//log.Info(string(body))
			//	ioutil.WriteFile("err.html", body, 0666)

			r := CreateItemsResp{}
			json.Unmarshal(body, &r)
			//fmt.Printf("%x", body)
			log.Info(r)
			if r.Status == "success" {
				log.Info("status success")
				err = ic.Update(bson.M{"num_iid": items[j].NumIid}, bson.M{"$set": bson.M{"item_id": r.ItemId, "uploaded": true}})
				if err != nil {
					log.Info(err.Error())
				}
			} else if r.ItemId != "" {
				log.Info("itemid is none")
				err = ic.Update(bson.M{"num_iid": items[j].NumIid}, bson.M{"$set": bson.M{"item_id": r.ItemId, "uploaded": false}})
				if err != nil {
					log.Info(err.Error())
				}
			}
		}
	}
}

func uploadRefreshItems() {
	log.Info("start to uploadRefreshItems")
	cc := utils.MongoInit(MGOHOST, MANGO, "taobao_cats")
	ic := mgoMango
	readyCats := make([]models.TaobaoItemCat, 0)
	cc.Find(bson.M{"matched_guoku_cid": bson.M{"$gt": 0}}).All(&readyCats)
	for _, v := range readyCats {
		log.Info("start", v.ItemCat.Cid)
		items := make([]models.TaobaoItemStd, 0)
		ic.Find(bson.M{"cid": v.ItemCat.Cid, "refreshed": true, "item_imgs.0": bson.M{"$exists": true}, "score": bson.M{"$gt": 2}}).Sort("-refresh_time").All(&items)
		for j := range items {
			log.Info("deal with ", items[j].NumIid)
			if items[j].Title == "" {
				continue
			}
			params := url.Values{}
			utils.GetUploadItemParams(&items[j], &params, v.MatchedGuokuCid)
			resp, err := http.PostForm("http://114.113.154.47:8000/management/entity/create/offline/", params)
			//resp, err := http.PostForm("http://114.113.154.47:8000/management/entity/create/offline/", params)
			log.Infof("%+v", params)
			if err != nil {
				continue
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Error(err)
			}
			log.Info(resp.Status)
			//log.Info(string(body))
			//ioutil.WriteFile("err.html", body, 0666)

			r := CreateItemsResp{}
			json.Unmarshal(body, &r)
			//fmt.Printf("%x", body)
			log.Info(r.Status)
			if r.Status == "success" || r.Status == "updated" {
				log.Info("status success")
				err = ic.Update(bson.M{"num_iid": items[j].NumIid}, bson.M{"$set": bson.M{"item_id": r.ItemId, "refreshed": true, "refresh_time": time.Now()}})
				if err != nil {
					log.Info(err.Error())
				}
			} else if r.ItemId != "" {
				log.Info("itemid is none")
				err = ic.Update(bson.M{"num_iid": items[j].NumIid}, bson.M{"$set": bson.M{"item_id": r.ItemId}})
				if err != nil {
					log.Info(err.Error())
				}
			}
		}
	}
}
func main() {
	runtime.GOMAXPROCS(4)

	go func() {
		for {
			syncOnlineItems()
			time.Sleep(time.Hour)
		}
	}()

	go func() {
		for {
			uploadOfflineItems()
			time.Sleep(time.Hour)
		}
	}()

	go func() {
		for {
			uploadRefreshItems()
			time.Sleep(time.Hour)
		}
	}()
	select {}
}
