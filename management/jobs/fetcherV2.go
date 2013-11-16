package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"runtime"

	"io/ioutil"
	"strconv"
	"time"
)

const (
	TASKFIRST      string = "first"
	TASKFAILED     string = "failed"
	TASKUPDATE     string = "update"
	PNUM           int    = 10
	MGOPARSER      string = "parser"
	MGOSOURCE      string = "collection"
	MGOFAILED      string = "failed"
	FAILEDCHANBUF  int    = 200
	SUCCESSCHANBUF int    = 500
	MAXFAILED      int    = 200
)

//保存failed的商品结构
type FailedStruct struct {
	ItemId string
	ShopId string
	From   string
}

type SuccessStruct struct {
	Salescount int
	Reviews    int
	From       string
	ItemId     string
	ShopId     string
	UpdateTime time.Time
}

type Crawler struct {
	ExecChans     chan bool //设置可同时运行多少个
	DoneChans     chan bool //爬取完成就发送一个值，使得ExecChans可以继续装载
	NextShopChans chan bool //爬取下一家店
	UpdateChans   chan *SuccessStruct
	FailedChans   chan *FailedStruct
	SuccessChans  chan *SuccessStruct
	Client        *http.Client
	UserAgents    []string
	reSales       *regexp.Regexp
	reReviews     *regexp.Regexp
	reCount       *regexp.Regexp
	MgoSource     *mgo.Collection
	MgoParsed     *mgo.Collection
	MgoFailed     *mgo.Collection
}

func (this *Crawler) Init() {
	this.ExecChans = make(chan bool, PNUM)
	this.DoneChans = make(chan bool)
	this.NextShopChans = make(chan bool)
	this.FailedChans = make(chan *FailedStruct, FAILEDCHANBUF)
	this.SuccessChans = make(chan *SuccessStruct, SUCCESSCHANBUF)
	this.UpdateChans = make(chan *SuccessStruct, SUCCESSCHANBUF)
	this.Client = &http.Client{}

	this.reSales = regexp.MustCompile("月&nbsp; 销&nbsp; 量：\\d+")
	this.reCount = regexp.MustCompile("\\d+")
	this.reReviews = regexp.MustCompile("评价\\( \\d+")
	this.UserAgents = []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.8; rv:24.0) Gecko/20100101 Firefox.24.0",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:22.0) Gecko/20100101 Firefox/22.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/28.0.1500.63 Safari/537.36",
		"Mozilla/5.0(iPad; U; CPU iPhone OS 3_2 like Mac OS X; en-us) AppleWebKit/531.21.10 (KHTML, like Gecko) Version/4.0.4 Mobile/7B314 Safari/531.21.10",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8.1.11) Gecko/20071127 Firefox/2.0.0.11",
		"Mozilla/5.0 (Macintosh; U; Intel Mac OS X; en; rv:1.8.1.6) Gecko/20070809 Camino/1.5.1",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; Trident/6.0)"}

	this.MgoSource = this.MongoInit(MGOSOURCE)
	this.MgoFailed = this.MongoInit(MGOFAILED)
	this.MgoParsed = this.MongoInit(MGOPARSER)
}

func (this *Crawler) MongoInit(collectionName string) *mgo.Collection {

	conf, err := toml.LoadFile("tbcrawler.toml")
	var mongoSetting *toml.TomlTree
	mongoSetting = conf.Get("mongodb").(*toml.TomlTree)
	session, err := mgo.Dial(mongoSetting.Get("host").(string))
	if err != nil {
		panic(err)
	}
	c := session.DB(mongoSetting.Get("db").(string)).C(mongoSetting.Get(collectionName).(string))
	return c
}

func (this *Crawler) UserAgentLit() string {
	length := len(this.UserAgents)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return this.UserAgents[r.Intn(length)]
}
func (this *Crawler) IsTmall(itemid string) (bool, error) {
	//判断这个商品是淘宝上的还是天猫上的
	url := "http://a.m.taobao.com/i" + itemid + ".htm"
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("User-Agent", this.UserAgentLit())
	resp, err := this.Client.Do(request)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("%s.\n Sleeping....", err.Error())
		return false, err
	}
	finalURL := resp.Request.URL.String()
	if finalURL == url {
		return false, nil
	} else {
		return true, nil
	}

}

type ShopItem struct {
	Date       time.Time
	Items_list []string
	Items_num  int
	Shop_id    int
	State      string
}

func (this *Crawler) LoadItems() (string, []string) {
	shopitem := new(ShopItem)
	this.MgoSource.Find(bson.M{"state": "posted"}).One(&shopitem)
	return strconv.Itoa(shopitem.Shop_id), shopitem.Items_list
}

func (this *Crawler) ShopStatuUpdate(shopid string) {

	shop_id, _ := strconv.Atoi(shopid)
	this.MgoSource.Update(bson.M{"shop_id": shop_id}, bson.M{"$set": bson.M{"state": "fetched"}})
}

func (this *Crawler) SuccessDataSave() {
	for data := range this.SuccessChans {
		this.MgoParsed.Insert(data)
	}
}

func (this *Crawler) FailedDataSave() {
	i := 0
	for data := range this.FailedChans {
		if i > MAXFAILED {
			log.Printf("抓取错误超过%d个，可能淘宝禁止了，先休眠五分钟")
			time.Sleep(5 * 60 * time.Second)
			i = 0
		}
		i = i + 1
		this.MgoFailed.Insert(data)
	}
}

func (this *Crawler) UpdatedDataSave() {
	for data := range this.UpdateChans {

		this.MgoParsed.Update(bson.M{"shopid": data.ShopId}, bson.M{"$set": bson.M{"salescount": data.Salescount, "reviews": data.Reviews, "updatetime": time.Now()}})
	}
}
func (this *Crawler) ParallelRequest(itemids []string, shopid string, task string, from string) {
	total := len(itemids)
	if total == 0 {
		return
	}
	if task == TASKFIRST || task == TASKFAILED {
		if from == "" {
			isTmall, err := this.IsTmall(itemids[total-1])
			if err != nil {
				return
			} else {
				if isTmall {
					from = "tmall.com"
				} else {
					from = "taobao.com"
				}
			}
		}
		go func() {
			for i := 0; i < total; i++ {
				<-this.DoneChans
				<-this.ExecChans
			}
			this.ShopStatuUpdate(shopid)
			this.NextShopChans <- true
		}()
		go this.SuccessDataSave()
		go this.FailedDataSave()
		go this.UpdatedDataSave()
		for i := 0; i < total; i++ {
			go this.Fetch(i, itemids[i], shopid, from, task)
		}
	}
}
func (this *Crawler) Fetch(i int, itemid string, shopid string, from string, task string) {

	this.ExecChans <- true
	log.Printf("start to fetch item %d:%s", i, itemid)
	isOk := true
	var salescount int
	var reviews int
	//先抓取一个页面，再抓取详情页面
	var furl string
	var nurl string
	if !(from == "tmall.com") {
		furl = "http://a.m.taobao.com/i" + itemid + ".htm"
		nurl = "http://a.m.taobao.com/da" + itemid + ".htm"
	} else {
		furl = "http://a.m.tmall.com/i" + itemid + ".htm"
		nurl = "http://a.m.tmall.com/da" + itemid + ".htm"
	}
	request, _ := http.NewRequest("GET", furl, nil)
	request.Header.Set("User-Agent", this.UserAgentLit())
	resp, err := this.Client.Do(request)
	if err != nil {
		log.Printf(err.Error())
		client := &http.Client{}
		resp, err = client.Do(request)
		if err != nil {
			log.Printf("item %s fetch error", itemid)
			if task == TASKUPDATE {
				return
			} else {
				failed := FailedStruct{ItemId: itemid, ShopId: shopid, From: from}
				this.FailedChans <- &failed
				this.DoneChans <- isOk
				return
			}
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err == nil {
		match := this.reSales.FindStringSubmatch(string(body))
		//log.Println(match)
		if len(match) == 0 {
			log.Printf("item %s 爬取下来的页面没有月销量，可能淘宝禁止了", itemid)

			if task == TASKUPDATE {
				return
			} else {
				failed := FailedStruct{ItemId: itemid, ShopId: shopid, From: from}
				this.FailedChans <- &failed
				this.DoneChans <- isOk
				return
			}

		}
		sc := this.reCount.FindStringSubmatch(match[0])
		salescount, _ = strconv.Atoi(sc[0])
		//log.Println(sc[0])
		isOk = true
	} else {
		log.Printf("item %s 从get请求里无法读取数据", itemid)
		log.Printf("got an err or in Fetch fpage")

		if task == TASKUPDATE {
			return
		} else {
			failed := FailedStruct{ItemId: itemid, ShopId: shopid, From: from}
			this.FailedChans <- &failed
			this.DoneChans <- isOk
			return
		}
	}
	nrequest, _ := http.NewRequest("GET", nurl, nil)
	nrequest.Header.Set("User-Agent", this.UserAgentLit())
	nresp, err := this.Client.Do(nrequest)
	if err != nil {
		log.Printf(err.Error())
		client := &http.Client{}
		nresp, err = client.Do(nrequest)
		if err != nil {
			log.Printf("item %s get 请求出错", itemid)
			log.Printf(err.Error())

			if task == TASKUPDATE {
				return
			} else {
				failed := FailedStruct{ItemId: itemid, ShopId: shopid, From: from}
				this.FailedChans <- &failed
				this.DoneChans <- isOk
				return
			}

		}

	}
	nbody, err := ioutil.ReadAll(nresp.Body)
	if err == nil {
		match := this.reReviews.FindStringSubmatch(string(nbody))
		//	log.Println(match)
		if len(match) == 0 {
			log.Printf("item %s 没有提供评论数据，可能淘宝禁止了", itemid)
			if task == TASKUPDATE {
				return
			} else {
				failed := FailedStruct{ItemId: itemid, ShopId: shopid, From: from}
				this.FailedChans <- &failed
				this.DoneChans <- isOk
				return

			}
		}
		sc := this.reCount.FindStringSubmatch(match[0])
		reviews, _ = strconv.Atoi(sc[0])
		//log.Println(sc[0])
		isOk = true
	} else {
		log.Printf("item %d fetch error", itemid)
		log.Printf("从评论页面的get请求里读取数据出现错误,%s", err.Error())
		if task == TASKUPDATE {
			return
		} else {
			failed := FailedStruct{ItemId: itemid, ShopId: shopid, From: from}
			this.FailedChans <- &failed
			this.DoneChans <- isOk
			return

		}
	}
	log.Printf("商品:%s ,评价:%d,销量:%d", itemid, reviews, salescount)
	now := time.Now()
	detail := SuccessStruct{From: from, Reviews: reviews, Salescount: salescount, ShopId: shopid, ItemId: itemid, UpdateTime: now}
	if task == TASKUPDATE {
		this.UpdateChans <- &detail
	} else {
		this.SuccessChans <- &detail
	}
	if task == TASKFAILED {
		this.MgoFailed.RemoveAll(bson.M{"itemid": itemid})
	}
	defer nresp.Body.Close()
	log.Printf("item  %d had fetched:%s", i, itemid)
	defer (func() {
		this.DoneChans <- isOk
	})()

}

func (this *Crawler) FailedUpdate() (string, []string, string) {
	var tmp *FailedStruct
	this.MgoFailed.Find(nil).One(&tmp)
	fmt.Println(tmp.ShopId, " get from failed shopid")
	var again []*FailedStruct
	this.MgoFailed.Find(bson.M{"shopid": tmp.ShopId}).All(&again)
	var itemids []string
	for _, v := range again {
		itemids = append(itemids, v.ItemId)

	}
	return tmp.ShopId, itemids, tmp.From
}

func (this *Crawler) ItemsUpdate() (string, []string, string) {
	var tmp *SuccessStruct
	now := time.Now()
	lastupdate := now.Add(-7 * 24 * time.Hour)
	this.MgoParsed.Find(bson.M{"updatetime": bson.M{"$lte": lastupdate}}).One(&tmp)
	var again []*SuccessStruct
	this.MgoParsed.Find(bson.M{"updatetime": bson.M{"$lte": lastupdate}, "shopid": tmp.ShopId}).All(&again)
	var itemids []string
	for _, v := range again {
		itemids = append(itemids, v.ItemId)
	}
	return tmp.ShopId, itemids, tmp.From
}

func (this *Crawler) NewFetch(task string) {

	this.Init()
	var shopid string
	var itemids []string
	var from string = ""
	if task == TASKFIRST {

		shopid, itemids = this.LoadItems()
	} else if task == TASKFAILED {
		shopid, itemids, from = this.FailedUpdate()
	}
	runtime.Gosched()
	log.Printf("任务类型 %s", task)
	go this.ParallelRequest(itemids, shopid, task, from)
	log.Println("开始抓取数据")
	<-this.NextShopChans

	log.Println("店铺 %s 已经爬取完毕,休息100秒", shopid)

}
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	ticker := time.NewTicker(time.Second * 100)
	go func() {
		crawler := new(Crawler)
		for _ = range ticker.C {
			crawler.NewFetch("first")
		}
	}()

	fticker := time.NewTicker(time.Second * 120)
	go func() {
		crawler := new(Crawler)
		for _ = range fticker.C {
			crawler.NewFetch("failed")
		}
	}()

	ntciker := time.NewTicker(time.Second * 100)
	go func() {
		crawler := new(Crawler)
		for _ = range ntciker.C {
			crawler.NewFetch("update")
		}
	}()

	select {}
}
