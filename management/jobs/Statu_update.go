package jobs

// Statu_update updates shop statu to "queued" whose statu is "finised"
import (
	"Mango/management/models"

	"log"
	"time"

	"github.com/pelletier/go-toml"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

func update() {
	conf, err := toml.LoadFile("conf/config.toml")
	var mongoSetting *toml.TomlTree
	mongoSetting = conf.Get("mongodb.test").(*toml.TomlTree)
	session, err := mgo.Dial(mongoSetting.Get("host").(string))
	if err != nil {
		panic(err)
	}
	defer session.Close()
	c := session.DB("mango").C("taobao_shops_depot")
	shops := make([]models.ShopItem, 100)
	//c.Update(bson.M{"status":"crawling"},bson.M{"$set":bson.M{"status":"queued"})
	c.Find(bson.M{"status": "finished"}).All(&shops)
	for _, shop := range shops {
		lastupdatetime := shop.LastCrawledTime
		now := time.Now()
		diff := now.Sub(lastupdatetime)
		cycle := shop.CrawlerInfo.Cycle
		if diff.Hours() > float64(cycle) {
			c.Update(bson.M{"shop_info.sid": shop.ShopInfo.Sid}, bson.M{"$set": bson.M{"status": "queued"}})
			log.Print("update one shop statu to queued")
		}
	}

}

func Update_statu() {
	log.Print("statu update is running")
	update()
	ticker := time.NewTicker(time.Hour * 1)

	for t := range ticker.C { //无限循环
		log.Print(t)
		update()
	}

}
