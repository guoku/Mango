package segment

import (
	"log"
	"testing"
)

func TestLoadData(t *testing.T) {
	result, _ := LoadData(1999, 2999)
	ToHTML(result, "result2.html")
	s := SplitTextToWords([]byte("我在中国China北京12389 America美国"))
	log.Println(TextSliceToString(s))
	a := []string{"abc"}
	r := []rune(a[0])
	log.Println(len(r))
}
func TestLoadDictionary(t *testing.T) {
	tree := new(TrieTree)
	//	tree.LoadDictionary("10.0.1.23", "words", "brands")
	tree.Add("苹果", 10)
	result := tree.Extract("三星外贸正品苹果/出口德国订单Schmidt原单 烤箱隔热手套/烘焙手套")
	log.Println(result)
	a := []rune("中国z")
	log.Println(a[0])
	log.Println(a[2])
}

func TestCleanning(t *testing.T) {
	tree := new(TrieTree)
	tree.LoadDictionary("10.0.1.23", "words", "brands")
	tree.LoadBlackWords("10.0.1.23", "words", "dict_chi_eng")
	tc := tree.Extract("韩国进口正品  吕洗发水/防脱深层修复 护发素400ml爱茉莉")
	log.Println(tc)
}
func TestExtract(t *testing.T) {
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
