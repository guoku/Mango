package segment

import (
	"github.com/qiniu/log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strings"
)

type Brand struct {
	Freq int
	Name string
}

type ValidBrand struct {
	Name  string
	Valid bool
}
type Black struct {
	Word        string
	Freq        int
	Prob        float64
	Blacklisted bool
	Deleted     bool
	Type        string
}

func (this *TrieTree) LoadNormal(mgohost, mgodb, mgocol string) {
	conn, err := mgo.Dial(mgohost)
	if err != nil {
		log.Info("mongo连接错误")
		panic(err)
	}
	session := conn.DB(mgodb).C(mgocol)

	var norms []*Black
	err = session.Find(bson.M{"blacklisted": false, "deleted": false, "freq": bson.M{"$gt": 100}}).All(&norms)
	if err != nil {
		log.Info(err.Error())
	}
	for _, v := range norms {
		text := strings.ToLower(v.Word)
		text = strings.TrimSpace(text)
		this.AddNormal(text)
	}
}
func (this *TrieTree) LoadDictionary(mgohost, mgodb, mgocol string) {
	conn, err := mgo.Dial(mgohost)
	if err != nil {
		log.Info("mongo连接错误")
		panic(err)
	}
	session := conn.DB(mgodb).C(mgocol)
	var brands []*Brand
	session.Find(bson.M{"freq": bson.M{"$gt": 200}}).All(&brands)
	var bm map[string]int = make(map[string]int)
	for _, brand := range brands {
		if brand.Freq > 30 {
			name := strings.ToLower(brand.Name)
			name = strings.TrimSpace(name)
			name = strings.Replace(name, " / ", "/", 1)
			bm[name] = brand.Freq
		}
	}
	var valids []*ValidBrand
	session.Find(bson.M{"valid": true}).All(&valids)
	for _, valid := range valids {
		name := strings.ToLower(valid.Name)
		name = strings.TrimSpace(name)
		name = strings.Replace(name, " / ", "/", 1)
		bm[name] = 9999
	}
	for k, v := range bm {
		this.Add(k, v)
	}
	defer conn.Clone()
}

func (this *TrieTree) LoadBlackWords(mgohost, mgodb, mgocol string) {
	conn, err := mgo.Dial(mgohost)
	if err != nil {
		log.Info("mongo连接错误")
		panic(err)
	}
	session := conn.DB(mgodb).C(mgocol)
	var blacks []*Black
	session.Find(bson.M{"blacklisted": true}).All(&blacks)
	for _, black := range blacks {
		if black.Freq > 100 {
			b := strings.ToLower(black.Word)
			b = strings.TrimSpace(b)
			this.AddBlackWord(b)
		}
	}
	defer conn.Clone()
}
