package filter

import (
	//"encoding/json"
	"fmt"
	"github.com/qiniu/log"
	//"io/ioutil"
	//"net/http"
	//"sort"
    "strings"
	"Mango/management/segment"
	"testing"
)

func TestFilterBrand(t *testing.T) {
	tree := new(TrieTree)
	tree.LoadDictionary("10.0.1.23", "words", "brands")
	tree.LoadBlackWords("10.0.1.23", "words", "dict_chi_eng")
	var sego *segment.GuokuSegmenter = new(segment.GuokuSegmenter)
	sego.LoadDictionary()
	texts := sego.Segment("L‘occitane J.L.YTOURNEL Maybelline/美宝莲tuleste market三星hello kitty China&HongKong 苹果外贸原单进口，幸福小天使个性定制日本德国进口I我 am Here'89 89韩国进口正品[【】]hello@kitty打折 吕洗发水/防脱深层修复 护发素400ml爱茉莉")
	fmt.Println(texts)
    fmt.Println(strings.Join(SegSliceToSegString(texts), "||"))
	texts := sego.Segment("春季潮男衬衫")
	log.Info(texts)
	brand := tree.FilterBrand(texts)
	log.Info(brand)
	clean := tree.Filtrate(texts)
	log.Info(clean)
	log.Info(tree.Cleanning("L‘occitane J.L.YTOURNEL Maybelline/美宝莲tuleste market三星hello kitty China&HongKong 苹果外贸原单进口，幸福小天使个性定制日本德国进口I我 am Here'89 89韩国进口正品[【】]hello@kitty打折 吕洗发水/防脱深层修复 护发素400ml爱茉莉"))
}
/*
func TestLoadData(t *testing.T) {
	t.SkipNow()
	result, _ := LoadData(940000, 2999)
	ToHTML(result, "result3.html")
	s := SplitTextToWords([]byte("我在中国China北京12389 America美国"))
	log.Println(TextSliceToString(s))
	a := []string{"abc"}
	r := []rune(a[0])
	log.Println(len(r))
}
func TestLoadDictionary(t *testing.T) {
	t.SkipNow()
	tree := new(TrieTree)
	//	tree.LoadDictionary("localhost", "words", "brands")
	tree.Add("苹果", 10)
	result := tree.Extract("三星外贸正品苹果/出口德国订单Schmidt原单 烤箱隔热手套/烘焙手套")
	log.Println(result)
	a := []rune("中国z")
	log.Println(a[0])
	log.Println(a[2])
}
func TestProcess(t *testing.T) {
	//sql := `use guoku_12_09;`
	//var R []*Result
	t.SkipNow()
	tree := new(TrieTree)
	tree.LoadDictionary("10.0.1.23", "words", "brands")
	tree.LoadBlackWords("10.0.1.23", "words", "dict_chi_eng")
	for i := 0; i < 10000000; i = i + 1000 {
		link := "http://114.113.154.47:8000/management/entity/without/title/sync/?offset=%d&count=%d"
		link = fmt.Sprintf(link, i, 1000)
		resp, err := http.Get(link)
		var rs []*Result
		if err != nil {
			panic(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		var entities []Entity
		err = json.Unmarshal(body, &entities)
		if err != nil {
			log.Println(err.Error())
			return
		}
		var result []*Result
		for _, ent := range entities {
			if len(ent.Titles) == 0 {
				continue
			}
			title := ent.Titles[0]
			//	log.Println("\n")
			//	log.Println("原始的标题：", title)
			brands := tree.Extract(title)
			//	log.Println("抽取出来的品牌名：", brands)
			title = tree.Cleanning(title)
			//	log.Println("清理的标题：", title)
			result := Result{Id: ent.ID, Title: ent.Titles[0], Brands: brands, CleanTitle: title, Category: ent.Category}
			rs = append(rs, &result)
		}
		log.Println("执行一次")
		for _, r := range result {
			sqlstr := `update base_entity set brand="%s",title="%s" where id=%d;` + "\n"
			b := Brandsprocess(r.Brands)
			s := fmt.Sprintf(sqlstr, b, r.CleanTitle, r.Id)
			//sql = sql + s
			ioutil.WriteFile("result.sql", []byte(s), 0666)
			log.Println(s)
		}
		//R = append(R, rs...)
		resp.Body.Close()
	}
	//sort.Sort(ById(R))
}
func TestCleanning(t *testing.T) {
	t.SkipNow()
	tree := new(TrieTree)
	tree.LoadDictionary("10.0.1.23", "words", "brands")
	tree.LoadBlackWords("10.0.1.23", "words", "dict_chi_eng")
	tc := tree.Cleanning("I我 am Here'89 89韩国进口正品[【】]hello@kitty打折 吕洗发水/防脱深层修复 护发素400ml爱茉莉")
	log.Println(tc)
	log.Println(fmt.Sprintf("<td><br>%s</td>", tc))
	ioutil.WriteFile("test", []byte(tc), 0777)
}
func TestExtract(t *testing.T) {
	t.SkipNow()
	tree := new(TrieTree)
	tree.Add("qiaobang/乔邦", 10)
	tree.Add("电热", 20)
	s, exist := tree.Search("电热")
	for _, v := range s {
		if !exist {
			t.Fatal("不存在")
		} else {
			if v.Words != "电热" {
				t.Fatal("不存在")
			}
		}
	}
	result := tree.Extract("电热乔邦,热水壶水壶")
	log.Println(result)
	if len(result) != 2 {
		t.Fatal("返回的结果数量不对")
	}
	log.Println(result)
}
func TestAdd(t *testing.T) {
	t.SkipNow()
	tree := new(TrieTree)
	tree.Add("我是/中国人", 1)
	tree.Add("我在中间", 2)

	s, exist := tree.Search("我是")
	for _, v := range s {
		if exist {
			if v.Words == "我是/中国人" {
				t.Log("通过")
			} else {
				t.Fatal("存在，但是返回值不对")
			}
		} else {
			t.Fatal("查找不存在，不通过")
		}
	}
	s2, exist2 := tree.Search("我是/中国人")
	for _, v := range s2 {
		if exist2 {
			if v.Words == "我是/中国人" {
				t.Log("通过")
			} else {
				t.Fatal("存在，但是返回值不对")
			}
		} else {
			t.Fatal("查找不存在，不通过")
		}
	}
	s3, exist3 := tree.Search("中国人")
	for _, v := range s3 {
		if exist3 {
			if v.Words == "我是/中国人" {
				t.Log("通过")
			} else {
				t.Fatal("存在，但是返回值不对")
			}
		} else {
			t.Fatal("查找不存在，不通过")
		}
	}
	s4, exist4 := tree.Search("我在中间")
	for _, v := range s4 {
		if exist4 {
			if v.Words == "我在中间" {
				t.Log("通过")
			} else {
				t.Fatal("存在，但是返回值不对")
			}
		} else {
			t.Fatal("查找不存在，不通过")
		}
	}
	_, exist5 := tree.Search("我不在北京")
	if exist5 {
		t.Fatal("查找不存在的词成功，词典查询有错误")
	}
}
*/
