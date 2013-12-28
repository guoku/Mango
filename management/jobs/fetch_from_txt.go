package main

import (
	//"Mango/management/crawler"
	"Mango/management/crawler"
	"Mango/management/filter"
	"Mango/management/utils"
	"bufio"
	"flag"
	"fmt"
	"github.com/qiniu/log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
	"strings"
	"time"
)

const (
	MGOHOST string = "10.0.1.23"
	MGODB   string = "zerg"
	MANGO   string = "mango"
)

type out struct {
	InStock  bool
	EntityId string
	ShopId   string
	ItemId   string
	Price    float64
	Title    string
	Brand    string
}

var mgopages *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "pages")
var mgofailed *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "failed")
var mgominer *mgo.Collection = utils.MongoInit(MGOHOST, MGODB, "minerals")
var mgoMango *mgo.Collection = utils.MongoInit(MGOHOST, MANGO, "taobao_items_depot")

func main() {
	var startline int
	flag.IntVar(&startline, "start", 1, "从第几行开始")
	flag.Parse()
	tree := new(filter.TrieTree)
	tree.LoadDictionary("10.0.1.23", "words", "brands")
	tree.LoadBlackWords("10.0.1.23", "words", "dict_chi_eng")
	file, err := os.Open("taobao.txt")
	if err != nil {
		pwd, _ := os.Getwd()
		log.Info(pwd)
		log.Error(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	log.Info("scanner")
	w, err := os.Create("processed.txt")
	if err != nil {
		log.Error(err)
	}
	defer w.Close()
	wscanner := bufio.NewWriter(w)
	log.Info("wscanner")
	defer wscanner.Flush()
	i := 1
	for scanner.Scan() {
		if i < startline {
			i = i + 1
			continue
		}
		log.Info("processed line :", i)
		i = i + 1
		log.Info("for")
		t := strings.Split(scanner.Text(), "\t")
		log.Info(t)
		if len(t) < 2 {
			//continue
			return
		}
		//第一个是entity id，第二个以及后面的是商品的itemid
		for i := 1; i < len(t); i++ {
			o := new(out)
			shoptype := "taobao.com"
			itemid := t[i]
			istmall, _ := crawler.IsTmall(itemid)
			if istmall {
				shoptype = "tmall.com"
			}
			page, err, detail := crawler.Fetch(itemid, shoptype)

			instock := true
			if err != nil {
				if err.Error() != "404" {
					failed := crawler.FailedPages{ItemId: itemid, ShopType: shoptype, UpdateTime: time.Now().Unix(), InStock: true}

					err = mgofailed.Insert(&failed)
					if err != nil {
						log.Info(err.Error())
						mgofailed.Update(bson.M{"itemid": itemid}, bson.M{"$set": failed})
						continue
					}
				} else {
					instock = false
					o.InStock = instock
					o.EntityId = t[0]
					o.ItemId = t[i]
				}
			} else {
				shoplink := crawler.GetShopLink(page)
				log.Info("shoplink", shoplink)
				shopid, _ := crawler.GetShopid(shoplink)
				info, missing, err := crawler.Parse(page, detail, itemid, shopid, shoptype)
				if missing {
					instock = false
					info.InStock = false
				}
				brands := tree.Extract(info.Title)
				info.Brand = filter.Brandsprocess(brands)
				info.Title = tree.Cleanning(info.Title)
				crawler.Save(info, mgoMango)
				o.InStock = info.InStock
				instock = info.InStock
				o.EntityId = t[0]
				o.ItemId = t[i]
				o.Price = info.Price
				o.ShopId = shopid
				o.Title = info.Title
				o.Brand = info.Brand
				successpage := crawler.Pages{ItemId: itemid, ShopId: shopid, ShopType: shoptype, FontPage: page, UpdateTime: time.Now().Unix(), DetailPage: detail, Parsed: true, InStock: instock}
				err = mgopages.Insert(&successpage)
				if err != nil {
					log.Println(err.Error())
					mgopages.Update(bson.M{"itemid": itemid}, bson.M{"$set": successpage})
				}
			}
			//entitiesid,itemid,instock,title,price,brand
			s := fmt.Sprintf("%s\t%s\t%t\t%s\t%f\t%s", o.EntityId, o.ItemId, o.InStock, o.Title, o.Price, o.Brand)
			log.Info(s)
			log.Info("get from channel")
			wscanner.WriteString(s)
			wscanner.WriteString("\n")
			wscanner.Flush()
		}
	}

	log.Info("over")

}
