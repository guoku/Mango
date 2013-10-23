package revision

/*
  To change shop statu to "queued" whose statu is "crawling" and
  out of date for 3 days
*/

import (
	"Mango/management/models"

	"fmt"
	"github.com/pelletier/go-toml"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"time"
)

var MgoSession *mgo.Session

func init() {
	conf, err := toml.LoadFile("conf/config.toml")

	var mongoSetting *toml.TomlTree
	mongoSetting = conf.Get("mongodb.test").(*toml.TomlTree)
	session, err := mgo.Dial(mongoSetting.Get("host").(string))
	if err != nil {
		panic(err)
	}
	MgoSession = session

}

func change() {
	defer MgoSession.Close()
	c := MgoSession.DB("test").C("taobao_shops_depot")
	shops := make([]models.ShopItem, 100)
	//c.Update(bson.M{"status":"crawling"},bson.M{"$set":bson.M{"status":"queued"})
	c.Find(bson.M{"status": "crawling"}).All(&shops)
	for _, shop := range shops {
		lastupdatetime := shop.LastUpdatedTime
		now := time.Now()
		diff := now.Sub(lastupdatetime)
		fmt.Println(diff.Hours())
		if diff.Hours() > 3*24 {
			c.Update(bson.M{"_id": shop.ObjectId}, bson.M{"$set": bson.M{"status": "queued"}})
			log.Print("update one shop statu to queued")
		}
	}

}

func Run_statu_revision() {
	log.Print("statu change is running")
	change()
	ticker := time.NewTicker(time.Second * 3600)
	go func() {
		for t := range ticker.C {
			fmt.Println(t)
			change()
		}
	}()
	time.Sleep(time.Hour * 24 * 3)
	ticker.Stop()
	log.Print("now the change status(crawling to queued) jop is over")
}
