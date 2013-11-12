package main

import (
	"encoding/json"
	"fmt"
	"github.com/qiniu/iconv"
	"io/ioutil"
	"labix.org/v2/mgo"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

func Request() string {
	request, err := http.NewRequest("GET", "http://www.jd.com/ajaxservice.aspx?stype=SortJson", nil)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	request.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	ct, _ := iconv.Open("utf-8", "gb2312")
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	bd := string(body)
	bd = bd[25 : len(bd)-2]
	bd = ct.ConvString(bd)
	return bd

}

/*
以下三个struct是用来解析从京东抓取下来的数据的
*/
type Detail struct {
	U string
	N string
	I []string
}

type Subcategory struct {
	U string
	N string
	T string
	I []*Detail
}

type Category struct {
	Data []*Subcategory
}

type JDCategory struct {
	CID    int
	Name   string
	Link   string
	Parent int
}

func Convert(raw string) []*JDCategory {
	var category Category
	var jds []*JDCategory
	err := json.Unmarshal([]byte(raw), &category)
	if err != nil {
		fmt.Println(err.Error())

	}
	subcategory := category.Data
	for m, v := range subcategory {
		re := regexp.MustCompile("<[^>]*>")
		firstClass := re.ReplaceAllString(v.N, "")
		root := new(JDCategory)
		root.Link = "http://jd.com"
		root.Name = "jingdong"
		root.CID = 0
		jd := new(JDCategory)
		jd.Name = firstClass
		jd.Parent = root.CID
		jd.CID = 10 + m //根据id的位数就可以知道所属的层级
		jds = append(jds, root)
		jds = append(jds, jd)
		for k, sub := range v.I {
			var link string
			if strings.Contains(sub.U, "http") {
				link = sub.U
			} else {
				if strings.Contains(sub.U, "000") {
					t := sub.U
					link = "http://channel.jd.com/" + t[0:len(t)-4] + ".html"

				} else {
					link = "http://channel.jd.com/" + sub.U + ".html"
				}
			}
			jdsub := new(JDCategory)
			jdsub.Name = sub.N
			jdsub.Link = link
			jdsub.Parent = jd.CID
			jdsub.CID = 100 + k*10 + m
			jds = append(jds, jdsub)

			for i, det := range sub.I {
				tmp := strings.Split(det, "|")
				jdsubsub := new(JDCategory)
				jdsubsub.Name = tmp[1]
				jdsubsub.Parent = jdsub.CID
				jdsubsub.CID = 1000 + i*100 + k*10 + m
				if strings.Contains(tmp[0], "http") {
					jdsubsub.Link = tmp[0]
				} else {
					if jdsub.Link == "" {
						t := tmp[0]
						jdsubsub.Link = "http://channel.jd.com/" + t[0:len(t)-4] + ".html"
					} else {
						jdsubsub.Link = "http://channel.jd.com/" + tmp[0] + ".html"
					}
				}
				jds = append(jds, jdsubsub)
			}

		}

	}
	return jds
}

func Save(data []*JDCategory) {
	session, _ := mgo.Dial("localhost")
	c := session.DB("jd").C("category")
	for _, cat := range data {
		c.Insert(cat)
	}
	defer session.Close()

}

type ByID []*JDCategory

func (this ByID) Len() int {
	return len(this)
}
func (this ByID) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
func (this ByID) Less(i, j int) bool {
	return this[i].CID < this[j].CID
}

func main() {
	bd := Request()
	jdcats := Convert(bd)

	sort.Sort(ByID(jdcats))
	for _, c := range jdcats {
		d, _ := json.Marshal(c)
		fmt.Println(string(d))
	}
	fmt.Println(len(jdcats))
	Save(jdcats)
}
