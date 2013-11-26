package main

import (
	"errors"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Pages struct {
	ShopId     string
	ItemId     string
	FontPage   string
	DetailPage string
	ShopType   string
	UpdateTime int64
	Parsed     bool
	InStock    bool //是否下架了
}

type FailedPages struct {
	ShopId     string
	ItemId     string
	ShopType   string
	UpdateTime int64
	InStock    bool
}

func main() {
	mgominer := mongoInit("minerals")
	mgopage := mongoInit("pages")
	mgofailed := mongoInit("failed")
	for {
		go refetch(mgopage, mgofailed)
		run(mgominer, mgopage, mgofailed)
	}
}

var proxys []string = []string{
	"http://127.0.0.1:30048",
	"http://127.0.0.1:30049",
	"http://127.0.0.1:30050",
	"http://127.0.0.1:30051",
	"http://127.0.0.1:30052",
	"http://127.0.0.1:30053",
	"http://127.0.0.1:30054",
	"http://127.0.0.1:30055",
	"http://127.0.0.1:30056",
	"http://127.0.0.1:30057",
	"http://127.0.0.1:30058",
	"http://127.0.0.1:30059",
	"http://127.0.0.1:30060",
	"http://127.0.0.1:30061",
	"http://127.0.0.1:30062",
	"http://127.0.0.1:30063",
	"http://127.0.0.1:30064",
	"http://127.0.0.1:30065",
	"http://127.0.0.1:30066",
	"http://127.0.0.1:30067"}

func getTransport() (transport *http.Transport) {
	length := len(proxys)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	proxy := proxys[r.Intn(length)]
	url_i := url.URL{}
	url_proxy, _ := url_i.Parse(proxy)
	transport = &http.Transport{Proxy: http.ProxyURL(url_proxy), ResponseHeaderTimeout: time.Duration(30) * time.Second}
	return
}

func isTmall(itemid string) (bool, error) {
	url := "http://a.m.taobao.com/i" + itemid + ".htm"
	request, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return false, err
	} else {
		finalURL := resp.Request.URL.String()
		if finalURL == url {
			return false, nil
		} else {
			return true, nil
		}
	}
	resp.Body.Close()
	return true, nil

}
func fetch(itemid string, shoptype string) (html string, err error, detail string) {
	url := ""
	detailurl := ""
	if shoptype == "tmall.com" {
		url = "http://a.m.tmall.com/i" + itemid + ".htm"
		detailurl = "http://a.m.tmall.com/da" + itemid + ".htm"
	} else {
		url = "http://a.m.taobao.com/i" + itemid + ".htm"
		detailurl = "http://a.m.taobao.com/da" + itemid + ".htm"
	}
	transport := getTransport()
	client := &http.Client{Transport: transport}
	req, err := http.NewRequest("GET", url, nil)
	useragent := userAgentGen()
	req.Header.Set("User-Agent", useragent)
	if err != nil {
		log.Print(err.Error())
		return "", err, ""
	}
	log.Printf("start to do request")
	resp, err := client.Do(req)
	log.Printf("request has been done")
	if err != nil {
		if resp == nil {
			log.Println("当proxy不可达时，resp为空")
		}
		time.Sleep(1 * time.Second)
		log.Println(err.Error())
		return "", err, ""
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		//fmt.Println(resp.Request.URL.String())
		resplink := resp.Request.URL.String()
		if strings.Contains(resplink, "h5") {
			html = ""
			detail = ""
			err = errors.New("taobao forbidden")
			log.Println("taobao forbidden")
			return
		}
		robots, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err.Error())
			return "", err, ""
		}
		html = string(robots)
	} else {
		log.Println(resp.StatusCode)
		html = ""
		err = errors.New(resp.Status)
		return html, err, ""
	}
	resp.Body.Close()
	req, err = http.NewRequest("GET", detailurl, nil)
	req.Header.Set("User-Agent", useragent)
	if err != nil {
		log.Print(err.Error())
		return "", err, ""
	}
	resp, err = client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return "", err, ""
	}
	if resp.StatusCode == 200 {
		//fmt.Println(resp.Request.URL.String())
		resplink := resp.Request.URL.String()
		if strings.Contains(resplink, "h5") {
			html = ""
			detail = ""
			err = errors.New("taobao forbidden")
			log.Println("taobao forbidden")
			return
		}
		robots, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err.Error())
			return "", err, ""
		}
		detail = string(robots)
	} else {
		log.Println(resp.StatusCode)
		html = ""
		err = errors.New(resp.Status)
		return html, err, ""
	}
	resp.Body.Close()
	return
}

func mongoInit(collectionName string) *mgo.Collection {
	conf, err := toml.LoadFile("config.toml")
	var mongoSetting *toml.TomlTree
	mongoSetting = conf.Get("mongodb").(*toml.TomlTree)
	log.Println(mongoSetting.Get("host").(string))
	session, err := mgo.Dial(mongoSetting.Get("host").(string))
	if err != nil {
		log.Println("严重错误")
		panic(err)
	}
	db := mongoSetting.Get("db").(string)
	log.Println(db)
	c := session.DB(mongoSetting.Get("db").(string)).C(collectionName)
	return c
}

type ShopItem struct {
	Date       time.Time
	Items_list []string
	Items_num  int
	Shop_id    int
	State      string
}

func loadItems(session *mgo.Collection) (string, []string) {
	log.Printf("start to load items")
	shopitem := new(ShopItem)
	session.Find(bson.M{"state": "posted"}).One(&shopitem)
	log.Printf("load shop %s", shopitem.Shop_id)
	return strconv.Itoa(shopitem.Shop_id), shopitem.Items_list
}

var UserAgents []string = []string{
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:24.0) Gecko/20100101 Firefox.24.0",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:22.0) Gecko/20100101 Firefox/22.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/28.0.1500.63 Safari/537.36",
	"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8.1.11) Gecko/20071127 Firefox/2.0.0.11",
	"Mozilla/5.0 (Macintosh; U; Intel Mac OS X; en; rv:1.8.1.6) Gecko/20070809 Camino/1.5.1",
	"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; Trident/6.0)"}

func userAgentGen() string {
	length := len(UserAgents)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return UserAgents[r.Intn(length)]
}

func run(mgominer, mgopage, mgofailed *mgo.Collection) {
	var threads int = 1
	var allowchan chan bool = make(chan bool, threads) //同一时刻不能有过多的请求，否则goagent都会受不了的
	log.Printf("start to run fetch")
	shopid, items := loadItems(mgominer)
	log.Printf("load items success")
	shoptype := "taobao.com"
	if len(items) == 0 {
		log.Printf("%s has no items", shopid)
		return
	} else {
		log.Printf("get %d items", len(items))
		it, err := isTmall(items[0])
		log.Printf("judge success")
		if err != nil {
			log.Printf("There is an error during judge")
			log.Println(err.Error())
			return
		}
		if it {
			shoptype = "tmall.com"
		}
		var wg sync.WaitGroup
		for _, itemid := range items {
			allowchan <- true
			wg.Add(1)
			go func(itemid string) {
				defer wg.Done()
				log.Printf("start to fetch %s", itemid)
				page, err, detail := fetch(itemid, shoptype)

				if err != nil {
					log.Printf("%s failed", itemid)
					failed := FailedPages{ItemId: itemid, ShopId: shopid, ShopType: shoptype, UpdateTime: time.Now().Unix(), InStock: true}
					err = mgofailed.Insert(&failed)
					if err != nil {
						log.Println(err.Error())
						mgofailed.Update(bson.M{"itemid": itemid}, bson.M{"$set": failed})
					}
				} else {
					log.Printf("%s success", itemid)
					successpage := Pages{ItemId: itemid, ShopId: shopid, ShopType: shoptype, FontPage: page, UpdateTime: time.Now().Unix(), DetailPage: detail, Parsed: false, InStock: true}
					err = mgopage.Insert(&successpage)
					if err != nil {
						log.Println(err.Error())
						mgopage.Update(bson.M{"itemid": itemid}, bson.M{"$set": successpage})
					}
				}
				<-allowchan

			}(itemid)

		}
		wg.Wait()
	}
	//确定所有数据都爬取完毕了，才对state进行更新，防止中途的停止导致数据丢失
	/*
		for i := 0; i < threads; i++ {
			allowchan <- true
		}
	*/
	close(allowchan)
	sid, _ := strconv.Atoi(shopid)
	err := mgominer.Update(bson.M{"shop_id": sid}, bson.M{"$set": bson.M{"state": "fetched", "date": time.Now()}})
	if err != nil {
		log.Println("update minerals state error")
		log.Println(err.Error())
	}
}

func refetch(mgopage, mgofailed *mgo.Collection) {
	//重新抓取失败的
	log.Printf("start to run refetch")
	var threads int = 6
	var failed *FailedPages
	err := mgofailed.Find(nil).One(&failed)
	if err != nil {
		return
	}
	var fails []*FailedPages
	mgofailed.Find(bson.M{"shopid": failed.ShopId}).All(&fails)
	var allowchan chan bool = make(chan bool, threads) //同一时刻不能有过多的请求，否则goagent都会受不了的
	var wg sync.WaitGroup
	for _, item := range fails {
		allowchan <- true
		info, err := mgofailed.RemoveAll(bson.M{"itemid": item.ItemId})
		if err != nil {
			log.Println(info.Removed)

			log.Println(err.Error())
		}
		wg.Add(1)
		go func(itemid string) {
			defer wg.Done()
			log.Printf("start to  refetch %s", item.ItemId)
			page, err, detail := fetch(item.ItemId, failed.ShopType)
			if err != nil {
				log.Printf("%s refetch failed", item.ItemId)
				newfail := FailedPages{ItemId: item.ItemId, ShopId: failed.ShopId, ShopType: failed.ShopType, UpdateTime: time.Now().Unix(), InStock: true}
				err = mgofailed.Insert(&newfail)
				if err != nil {
					log.Println(err.Error())
					mgofailed.Update(bson.M{"itemid": item.ItemId}, bson.M{"$set": newfail})
				}
			} else {
				log.Printf("%s refetch successed", item.ItemId)
				successpage := Pages{ItemId: item.ItemId, ShopId: failed.ShopId, ShopType: failed.ShopType, FontPage: page, DetailPage: detail, UpdateTime: time.Now().Unix(), Parsed: false, InStock: true}
				err = mgopage.Insert(&successpage)
				if err != nil {
					log.Println(err.Error())
					mgopage.Update(bson.M{"itemid": item.ItemId}, bson.M{"$set": successpage})
				}
			}
			<-allowchan

		}(item.ItemId)
	}
	wg.Wait()
	close(allowchan)

}
