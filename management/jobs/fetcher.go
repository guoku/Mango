package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

const (
	FIRST  string = "first"
	FAILED string = "failed"
	UPDATE string = "update"
)

//保存failed的商品结构
type IDStruct struct {
	Itemid string
	Shopid string
	From   string
}

//保持爬虫成功的商品结构
type DetailData struct {
	Salescount int
	Reviews    int
	From       string
	Shopid     string
	Itemid     string
	UpdateTime time.Time
}
type Crawler struct {
	ExecChans   chan bool      //设置可同时运行多少个的通道
	DoneChans   chan bool      // 任何一个爬取完成，都发送布尔值给这个通道，从而保证ExecChans里完成了的排出
	NextChans   chan bool      //如果爬完一个店铺，就给这个通道发送一个值，从而保证可以不停从mongo里提取任务
	FailChans   chan *IDStruct //爬取失败的id放到这里，由一个定时启动的线程来完成爬取\
	DetailChans chan *DetailData
	Client      *http.Client
	Db          *sql.DB
	Pnum        int //同时可以运行的协程总数
	Interval    time.Duration
	UserAgents  []string
	Collection  *mgo.Collection
	Parsed      *mgo.Collection
	Failed      *mgo.Collection
	reSales     *regexp.Regexp
	reCount     *regexp.Regexp
	reReviews   *regexp.Regexp
}

func (this *Crawler) Init() {
	this.Pnum = 10
	this.ExecChans = make(chan bool, this.Pnum)
	this.DoneChans = make(chan bool, 1)
	this.NextChans = make(chan bool, 1)
	this.FailChans = make(chan *IDStruct, 5)
	this.DetailChans = make(chan *DetailData, 500)
	this.Client = &http.Client{}
	var err error
	this.Db, err = sql.Open("mysql", "root:123456@/guokuer")
	if err != nil {
		panic(err.Error())
	}
	this.Interval = 3 * time.Second
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
	this.Collection = this.MongoInit("collection")
	this.Parsed = this.MongoInit("parser")
	this.Failed = this.MongoInit("failed")

}

func (this *Crawler) ParallelRequest(itemids []string, shopid string) {
	total := len(itemids)
	if total == 0 {
		return
	}
	go func() {
		for i := 0; i < total; i++ {
			//	time.Sleep(3 * time.Second)
			r := <-this.DoneChans
			<-this.ExecChans
			if !r {
				log.Printf("第 %s 项获取失败", i)
			}

		}
		this.MongoUpdate(shopid)
		this.NextChans <- true //这里表示这个店铺爬取完毕，去爬下一个店铺
	}()
	go this.MongoInsert()
	go this.FailedSave()
	flag, err := this.IsTmall(itemids[total-1])
	if err != nil {
		time.Sleep(200 * time.Second)
		if total > 1 {
			flag, err = this.IsTmall(itemids[1])
		} else {
			flag, err = this.IsTmall(itemids[total-1])
		}
		if err != nil {
			this.MongoFail(shopid) //记录这个店铺爬取失败
			return
		}
	}
	for i := 0; i < total; i++ {
		go this.Fetch(i, itemids[i], shopid, flag)
	}
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
func (this *Crawler) Fetch(i int, itemid string, shopid string, flag bool) {

	this.ExecChans <- true
	log.Printf("start to fetch item %d:%s", i, itemid)
	isOk := true
	var salescount int
	var reviews int
	//先抓取一个页面，再抓取详情页面
	var furl string
	var nurl string
	var Type string
	if !flag {
		furl = "http://a.m.taobao.com/i" + itemid + ".htm"
		nurl = "http://a.m.taobao.com/da" + itemid + ".htm"
		Type = "taobao.com"
	} else {
		Type = "tmall.com"
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
			ids := IDStruct{Itemid: itemid, Shopid: shopid, From: Type}
			this.FailChans <- &ids
			this.DoneChans <- isOk
			return
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err == nil {
		match := this.reSales.FindStringSubmatch(string(body))
		//log.Println(match)
		if len(match) == 0 {
			log.Printf("item %s 爬取下来的页面没有月销量，可能淘宝禁止了", itemid)
			ids := IDStruct{Itemid: itemid, Shopid: shopid, From: Type}
			this.FailChans <- &ids
			this.DoneChans <- isOk
			return
		}
		sc := this.reCount.FindStringSubmatch(match[0])
		salescount, _ = strconv.Atoi(sc[0])
		//log.Println(sc[0])
		isOk = true
	} else {
		log.Printf("item %s 从get请求里无法读取数据", itemid)
		log.Printf("got an err or in Fetch fpage")
		ids := IDStruct{Itemid: itemid, Shopid: shopid, From: Type}
		this.FailChans <- &ids
		this.DoneChans <- isOk
		return
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
			ids := IDStruct{Itemid: itemid, Shopid: shopid, From: Type}
			this.FailChans <- &ids
			this.DoneChans <- isOk
			return
		}

	}
	nbody, err := ioutil.ReadAll(nresp.Body)
	if err == nil {
		match := this.reReviews.FindStringSubmatch(string(nbody))
		//	log.Println(match)
		if len(match) == 0 {
			log.Printf("item %s 没有提供评论数据，可能淘宝禁止了", itemid)
			ids := IDStruct{Itemid: itemid, Shopid: shopid, From: Type}
			this.FailChans <- &ids
			this.DoneChans <- isOk
			return
		}
		sc := this.reCount.FindStringSubmatch(match[0])
		reviews, _ = strconv.Atoi(sc[0])
		//log.Println(sc[0])
		isOk = true
	} else {
		log.Printf("item %d fetch error", itemid)
		log.Printf("从评论页面的get请求里读取数据出现错误,%s", err.Error())
		ids := IDStruct{Itemid: itemid, Shopid: shopid, From: Type}
		this.FailChans <- &ids
		this.DoneChans <- isOk
		return
	}
	log.Printf("商品:%s ,评价:%d,销量:%d", itemid, reviews, salescount)
	now := time.Now()
	detail := DetailData{From: Type, Reviews: reviews, Salescount: salescount, Shopid: shopid, Itemid: itemid, UpdateTime: now}
	this.DetailChans <- &detail
	defer nresp.Body.Close()
	log.Printf("item  %d had fetched:%s", i, itemid)
	defer (func() {
		this.DoneChans <- isOk
	})()

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

type ShopItem struct {
	Date       time.Time
	Items_list []string
	Items_num  int
	Shop_id    int
	State      string
}

func (this *Crawler) LoadItems() (string, []string) {
	shopitem := new(ShopItem)
	this.Collection.Find(bson.M{"state": "posted"}).One(&shopitem)
	return strconv.Itoa(shopitem.Shop_id), shopitem.Items_list
}

func (this *Crawler) MongoFail(shopid string) {
	shop_id, _ := strconv.Atoi(shopid)
	this.Collection.Update(bson.M{"shop_id": shop_id}, bson.M{"$set": bson.M{"state": "failed"}})
}
func (this *Crawler) MongoUpdate(shopid string) {
	shop_id, _ := strconv.Atoi(shopid)
	this.Collection.Update(bson.M{"shop_id": shop_id}, bson.M{"$set": bson.M{"state": "fetched"}})
}
func (this *Crawler) FailedSave() {
	i := 0
	for data := range this.FailChans {
		if i > 200 {
			log.Println("错误页面超过200个，可能淘宝禁止了，先休眠5分钟")
			time.Sleep(5 * 60 * time.Second)
			i = 0
		}
		i = i + 1
		this.Failed.Insert(data)
	}
}
func (this *Crawler) MongoInsert() {

	for data := range this.DetailChans {
		this.Parsed.Insert(data)
	}
}

func (this *Crawler) FailedUpdate() (string, []string) {
	//把failed库里面的商品再爬一遍，因为这些之前爬取失败了
	var tmp *IDStruct
	this.Failed.Find(nil).One(&tmp)
	fmt.Println(tmp.Shopid, " get from failed shopid ")
	var again []*IDStruct
	this.Failed.Find(bson.M{"shopid": tmp.Shopid}).All(&again)
	var itemids []string
	for _, v := range again {
		itemids = append(itemids, v.Itemid)
	}

	change, err := this.Failed.RemoveAll(bson.M{"shopid": tmp.Shopid})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d 条记录已经被删除了\n", change.Removed)
	return tmp.Shopid, itemids

}

func (this *Crawler) Update() (string, []string) {
	//定期更新已爬取过的商品数据
	now = time.Now()
	//一个星期更新一次
	lastupdate := now.Add(-7 * 24 * time.Hours())
	var tmp *DetailData
	this.Parsed.Find(bson.M{"updatetime": {"$lte": lastupdate}}).One(&tmp)
	var again []*DetailData
	this.Parsed.Find(bson.M{"updatetime": {"$lte": lastupdate}, "shopid": tmp.Shopid}).All(&again)

}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	go func() {
		crawler := new(Crawler)
		crawler.Init()
		for {
			shopid, itemids := crawler.LoadItems()
			if len(itemids) == 0 {
				continue
			}

			go crawler.ParallelRequest(itemids, shopid)
			fmt.Println("start to run")
			<-crawler.NextChans
			log.Printf("店铺 %s 已经爬取完毕，休息100秒", shopid)
			time.Sleep(100 * time.Second)
			failedCrawler := new(Crawler)
			failedCrawler.Init()

			shopid, itemids = failedCrawler.FailedUpdate()
			if len(itemids) == 0 {
				runtime.Gosched()
				continue
			}
			go failedCrawler.ParallelRequest(itemids, shopid)
			fmt.Println("开始重新抓取之前抓取失败的商品")
			<-failedCrawler.NextChans

			time.Sleep(30 * time.Second)
		}
	}()
	select {}
}
