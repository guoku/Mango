package crawler

import (
	"testing"
)

func TestGetShopItems(t *testing.T) {
	GetShopItems("http://sxdq.m.tmall.com/")
	GetShopItems("http://shop33585474.m.taobao.com")
}
