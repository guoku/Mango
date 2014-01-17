package crawler

import (
	"github.com/qiniu/log"
	"testing"
)

func TestFetchSpec(t *testing.T) {
	t.SkipNow()
	font, detail, _, _, _ := FetchItem("17046431987", "taobao.com")
	info, _, _ := ParsePage(font, detail, "17046431987", "", "taobao.com")
	log.Infof("%+v", info)
}
