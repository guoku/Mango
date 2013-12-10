package segment

import (
	"log"
	"testing"
)

func TestLoadData(t *testing.T) {
	//	LoadData()
	s := SplitTextToWords([]byte("我在中国China北京12389 America美国"))
	log.Println(TextSliceToString(s))
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
	//	tree.LoadDictionary("10.0.1.23", "words", "brands")
	//	tree.LoadBlackWords("10.0.1.23", "words", "dict_chi_eng")
	tree.Add("valentino", 10)
	tree.Add("valentino/华伦天奴", 20)
	tree.Add("red", 30)
	tree.Add("red valentino", 20)
	tree.Add("高档品牌", 30)
	tree.AddBlackWord("roland")
	tree.AddBlackWord("高档品")
	tree.AddBlackWord("高档")
	tree.AddBlackWord("意大利")
	tree.AddBlackWord("正品")
	tree.AddBlackWord("代购")
	tree.AddBlackWord("同款")
	tree.AddBlackWord("2013")
	t1 := "高档品100%重磅真丝怪料软垂 长袖收腰 春秋淑女连衣裙 原价3000"
	t2 := "高端特供 原单双T100%真丝针织重磅双面加厚 短袖高腰淑女连衣裙"
	t3 := "2013秋季新款ROLAND MOURET EbonyNO no 可爱猫咪针织裙娃娃领毛衣裙"
	t4 := "2013夏装 代购RED VALENTINO新款雪纺连衣裙韩版女装珍珠娃娃领百褶荷叶袖短袖裙子"
	t5 := "INao意大利正品代购 VALENTINO羊毛混纺迷你连衣裙名媛荣文蔚同款"

	t1c := tree.Cleanning(t1)
	log.Println(t1c)
	t2c := tree.Cleanning(t2)
	log.Println(t2c)
	t3c := tree.Cleanning(t3)
	log.Println(t3c)
	t4c := tree.Cleanning(t4)
	log.Println(t4c)
	log.Println(tree.Extract(t4c))
	t5c := tree.Cleanning(t5)
	log.Println(t5c)
	log.Println(tree.Extract(t5c))
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
