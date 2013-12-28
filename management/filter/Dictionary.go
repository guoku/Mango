package filter

import (
	"github.com/qiniu/log"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	//"regexp"
	"Mango/management/models"
	"Mango/management/segment"
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

func FromText(mgohost, mgodb, mgocol, filename string) {
	conn, err := mgo.Dial(mgohost)
	if err != nil {
		log.Info("mongo连接错误")
		panic(err)
	}
	session := conn.DB(mgodb).C(mgocol)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Info("mongo连接错误")
		panic(err)
	}
	content := string(data)
	lines := strings.Split(content, "\n")
	var valids []*ValidBrand
	for _, line := range lines {
		log.Info(line)
		line = strings.TrimSpace(line)
		if line != "" {
			v := ValidBrand{Name: line, Valid: true}
			valids = append(valids, &v)

		}
	}
	err = session.Insert(&valids)
	if err != nil {
		log.Info("mongo插入错误")
		panic(err)
	}
	conn.Close()

}

/*
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
*/
//加载品牌词
func (this *TrieTree) LoadDictionary(mgohost, mgodb, mgocol string) {
	conn, err := mgo.Dial(mgohost)
	defer conn.Clone()
	if err != nil {
		log.Info("mongo连接错误")
		panic(err)
	}
	session := conn.DB(mgodb).C(mgocol)
	brands := make([]models.BrandsWord, 0)
	session.Find(bson.M{"$or": []bson.M{bson.M{"valid": true}, bson.M{"deleted": false, "freq": bson.M{"$gt": 50}}}}).All(&brands)
	//	var brands []*Brand
	//	session.Find(bson.M{"freq": bson.M{"$gt": 1}, "deleted": bson.M{"$ne": true}}).All(&brands)
	//	re := regexp.MustCompile("^\\pP+|\\pP+$")
	var sego *segment.GuokuSegmenter = new(segment.GuokuSegmenter)
	sego.LoadDictionary()
	for _, brand := range brands {
		if brand.Freq > 30 {
			/*
				name := strings.ToLower(brand.Name)
				name = strings.TrimSpace(name)
				//	name = re.ReplaceAllString(name, "")
				name = strings.Replace(name, " / ", "/", 1)
				bm[name] = brand.Freq
			*/
			sname := sego.Segment(brand.Name)
			for _, s := range sname {
				this.Add(brand.Name, s, brand.Freq)
			}
		}
	}
	var valids []*ValidBrand
	session.Find(bson.M{"valid": true}).All(&valids)
	for _, valid := range valids {
		/*
			name := strings.ToLower(valid.Name)
			name = strings.TrimSpace(name)
			//name = re.ReplaceAllString(name, "")
			name = strings.Replace(name, " / ", "/", 1)
			bm[name] = 9999
		*/
		sname := sego.Segment(valid.Name)
		for _, s := range sname {
			this.Add(valid.Name, s, 9999)
		}
	}
}

func (this *TrieTree) LoadBlackWords(mgohost, mgodb, mgocol string) {
	conn, err := mgo.Dial(mgohost)
	defer conn.Clone()
	if err != nil {
		log.Info("mongo连接错误")
		panic(err)
	}
	session := conn.DB(mgodb).C(mgocol)
	blackwords := make([]models.DictWord, 0)
	session.Find(bson.M{"blacklisted": true, "deleted": bson.M{"$ne": true}}).All(&blackwords)
	for _, black := range blackwords {
		if black.Freq > 50 {
			//log.Info(sk)
			this.AddBlackWord(black.Word)

		}
	}
}
