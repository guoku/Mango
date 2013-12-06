package segment

import (
	"testing"
)

func TestAdd(t *testing.T) {
	tree := new(TrieTree)
	tree.Add("我是/中国人", 1)
	tree.Add("我在中间", 2)
	s, exist := tree.Search("我是")
	if exist {
		if s.Words == "我是/中国人" {
			t.Log("通过")
		} else {
			t.Fatal("存在，但是返回值不对")
		}
	} else {
		t.Fatal("查找不存在，不通过")
	}
	s2, exist2 := tree.Search("我是/中国人")
	if exist2 {
		if s2.Words == "我是/中国人" {
			t.Log("通过")
		} else {
			t.Fatal("存在，但是返回值不对")
		}
	} else {
		t.Fatal("查找不存在，不通过")
	}
	s3, exist3 := tree.Search("中国人")
	if exist3 {
		if s3.Words == "我是/中国人" {
			t.Log("通过")
		} else {
			t.Fatal("存在，但是返回值不对")
		}
	} else {
		t.Fatal("查找不存在，不通过")
	}
	s4, exist4 := tree.Search("我在中间")
	if exist4 {
		if s4.Words == "我在中间" {
			t.Log("通过")
		} else {
			t.Fatal("存在，但是返回值不对")
		}
	} else {
		t.Fatal("查找不存在，不通过")
	}
	_, exist5 := tree.Search("我不在北京")
	if exist5 {
		t.Fatal("查找不存在的词成功，词典查询有错误")
	}
}
