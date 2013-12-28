package crawler

import (
	"time"
)

type Info struct {
	Desc        string            `json:"desc"`
	Cid         int               `json:"cid"`
	Promprice   float64           `json:"promotion_price"`
	Price       float64           `json:"price"`
	Imgs        []string          `json:"item_imgs"`
	Count       int               `json:"monthly_sales_volume"`
	Reviews     int               `json:"reviews_count"`
	Nick        string            `json:"nick"`
	InStock     bool              `json:"in_stock"`
	Attr        map[string]string `json:"props"`
	Location    *Loc              `json:"location"`
	UpdateTime  int64             `json:"data_updated_time"`
	ItemId      int               `json:"num_iid"`
	Sid         int               `json:"sid"`
	Title       string            `json:"title"`
	Brand       string            `json:"brand"`
	ShopType    string            `json:"shop_type"`
	DetailUrl   string            `json:"detail_url"`
	GuokuItemid string            `json:"item_id"`
}

type Loc struct {
	State string
	City  string
}

type Pages struct {
	ShopId     string
	ItemId     string
	FontPage   string
	DetailPage string
	ShopType   string
	UpdateTime int64
	Parsed     bool
	InStock    bool //是否下架了
}

type FailedPages struct {
	ShopId     string
	ItemId     string
	ShopType   string
	UpdateTime int64
	InStock    bool
}

type ShopItem struct {
	Date       time.Time
	Items_list []string
	Items_num  int
	Shop_id    int
	State      string
}

var proxys []string = []string{
	"http://127.0.0.1:30048",
	"http://127.0.0.1:30049",
	"http://127.0.0.1:30050",
	"http://127.0.0.1:30051",
	"http://127.0.0.1:30052",
	"http://127.0.0.1:30053",
	"http://127.0.0.1:30054",
	"http://127.0.0.1:30055",
	"http://127.0.0.1:30056",
	"http://127.0.0.1:30057",
	"http://127.0.0.1:30058",
	"http://127.0.0.1:30059",
	"http://127.0.0.1:30060",
	"http://127.0.0.1:30061",
	"http://127.0.0.1:30062",
	"http://127.0.0.1:30063",
	"http://127.0.0.1:30064",
	"http://127.0.0.1:30065",
	"http://127.0.0.1:30066",
	"http://127.0.0.1:30067",
	"http://127.0.0.1:30068",
	"http://127.0.0.1:30069",
	"http://127.0.0.1:30070",
	"http://127.0.0.1:30071",
	"http://127.0.0.1:30072",
	"http://127.0.0.1:30073",
	"http://127.0.0.1:30074",
	"http://127.0.0.1:30075",
	"http://127.0.0.1:30076",
	"http://127.0.0.1:30077",
}
var states map[string]bool = map[string]bool{
	"北京":  true,
	"上海":  true,
	"天津":  true,
	"重庆":  true,
	"广东":  true,
	"江苏":  true,
	"山东":  true,
	"浙江":  true,
	"河北":  true,
	"山西":  true,
	"辽宁":  true,
	"吉林":  true,
	"河南":  true,
	"安徽":  true,
	"福建":  true,
	"江西":  true,
	"黑龙江": true,
	"湖南":  true,
	"湖北":  true,
	"海南":  true,
	"四川":  true,
	"贵州":  true,
	"云南":  true,
	"陕西":  true,
	"甘肃":  true,
	"青海":  true,
	"台湾":  true,
	"西藏":  true,
	"内蒙古": true,
	"广西":  true,
	"宁夏":  true,
	"新疆":  true,
	"香港":  true,
	"澳门":  true,
	"海外":  true,
}
