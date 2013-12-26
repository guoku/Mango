package crawler

import (
	"Mango/management/utils"
	"bytes"
	"compress/zlib"
	"github.com/qiniu/log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

//返回值font是商品的页面，detail是商品的详情页面，instock表示下架与否
//这里下架与否的判断设计得比较不好,如果抓取正常，instock是未知的，只有进行解析后才知道结果
//而如果出现了err，则应该看instock与否，如果下架了，这个itemid就不需要保存了
func FetchItem(itemid string, shoptype string) (font, detail string, instock bool, err error) {
	log.Infof("start to fetch %s", itemid)
	font, err, detail = utils.Fetch(itemid, shoptype)
	if err != nil {
		log.Infof("%s failed", itemid)
		if err.Error() != "404" {
			//说明不是因为商品下架而导致的失败
			instock = true
			return
		} else {
			//商品的页面已经找不到了
			instock = false
			return
		}
	}
	log.Infof("%s fetched successed!", itemid)
	return
}

//如果返回错误，应该看是否下架，下架了就不保存了
func ParsePage(font, detail, itemid, shopid, shoptype string) (info *utils.Info, instock bool, err error) {
	info, missing, err := utils.Parse(font, detail, itemid, shopid, shoptype)
	log.Info("解析完毕")
	instock = true
	if err != nil {
		if missing {
			instock = false
			return
		} else if err.Error() == "聚划算" || err.Error() == "cattag" {
			//商品来自聚划算或者找不着了
			//直接丢弃，不予以保存
			instock = false
			return
		} else {
			//只是解析错误，出现了新情况，暂时不管先
			log.Error(err)
			return
		}
	} else {
		instock = info.InStock
		//即使解析出来的结果还是下架的，但这个结果要保存
		//因为淘宝的这种下架方式还可以查看页面，就是不能购买
		//且有货了之后可能会重新上架
		return
	}
}

func SaveFailed(itemid, shopid, shoptype string, mgofailed *mgo.Collection) {
	failed := utils.FailedPages{ItemId: itemid, ShopId: shopid, ShopType: shoptype, UpdateTime: time.Now().Unix(), InStock: true}
	_, err := mgofailed.Upsert(bson.M{"itemid": itemid}, bson.M{"$set": failed})
	if err != nil {
		log.Error(err)
	}
}

func SaveSuccessed(itemid, shopid, shoptype, font, detail string, parsed, instock bool, mgopages *mgo.Collection) {
	font = Compress(font)
	log.Info("压缩后的字符", font)
	detail = Compress(detail)
	successpage := utils.Pages{ItemId: itemid, ShopId: shopid, ShopType: shoptype, FontPage: font, UpdateTime: time.Now().Unix(), DetailPage: detail, Parsed: parsed, InStock: instock}
	_, err := mgopages.Upsert(bson.M{"itemid": itemid}, bson.M{"$set": successpage})
	if err != nil {
		log.Error(err)
	}
	log.Info("保存页面数据成功")
}

func Compress(data string) string {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write([]byte(data))
	w.Close()
	return string(b.Bytes())
}

func Decompress(data string) (string, error) {
	buff := []byte(data)
	b := bytes.NewBuffer(buff)
	r, err := zlib.NewReader(b)
	if err != nil {
		log.Error(err)
		return "", err
	}
	rbuf := new(bytes.Buffer)
	rbuf.ReadFrom(r)
	return string(rbuf.Bytes()), nil
}
