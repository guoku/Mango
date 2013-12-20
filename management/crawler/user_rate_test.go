package crawler

import (
	"github.com/qiniu/log"
	"testing"
)

func TestGetUserid(t *testing.T) {
	userId, err := GetUserid("http://shop70766218.taobao.com/")
	if err != nil {
		t.Fatal(err)
	}
	log.Info(userId)
}

func TestParse(t *testing.T) {
	t.SkipNow()
	userid, err := GetUserid("http://dumex.tmall.com/?spm=a1z0b.7.w3-18454208515.1.EWtFUW")
	if err != nil {
		t.Fatal(err)
	}
	Parse(userid)
}
func TestParseTaobao(t *testing.T) {
	//t.SkipNow()
	userid, err := GetUserid("http://shop71839143.taobao.com/")
	if err != nil {
		t.Fatal(err)
	}
	Parse(userid)
}

func TestGetInfo(t *testing.T) {
	//GetInfo("http://shop71839143.taobao.com/")
	GetInfo("http://dumex.tmall.com/")
}
