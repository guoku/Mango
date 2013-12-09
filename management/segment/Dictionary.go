package segment

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
)

type Brand struct {
	Freq int
	Name string
}
type Black struct {
	Word        string
	Freq        int
	Prob        float64
	Blacklisted bool
}

func (this *TrieTree) LoadDictionary(mgohost, mgodb, mgocol string) {
	conn, err := mgo.Dial(mgohost)
	if err != nil {
		log.Println("mongo连接错误")
		panic(err)
	}
	session := conn.DB(mgodb).C(mgocol)
	var brands []*Brand
	session.Find(bson.M{"freq": bson.M{"$gt": 30}}).All(&brands)
	for _, brand := range brands {
		if brand.Freq > 30 {
			this.Add(brand.Name, brand.Freq)
		}
	}
}

func (this *TrieTree) LoadBlackWords(mgohost, mgodb, mgocol string) {
	conn, err := mgo.Dial(mgohost)
	if err != nil {
		log.Println("mongo连接错误")
		panic(err)
	}
	session := conn.DB(mgodb).C(mgocol)
	var blacks []*Black
	session.Find(bson.M{"blacklisted": true}).All(&blacks)
	for _, black := range blacks {
		if black.Freq > 30 {
			this.AddBlackWord(black.Word)
		}
	}
}
