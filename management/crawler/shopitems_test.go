package crawler

import (
	"testing"
)

func TestGetShopItems(t *testing.T) {
	t.SkipNow()
	GetShopItems("http://sxdq.m.tmall.com/")
	GetShopItems("http://shop33585474.m.taobao.com")
}
