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
    "http://10.0.1.23:30048",
    "http://10.0.1.23:30049",
    "http://10.0.1.23:30050",
    "http://10.0.1.23:30051",
    "http://10.0.1.23:30052",
    "http://10.0.1.23:30053",
    "http://10.0.1.23:30054",
    "http://10.0.1.23:30055",
    "http://10.0.1.23:30056",
    "http://10.0.1.23:30057",
    "http://10.0.1.23:30058",
    "http://10.0.1.23:30059",
    "http://10.0.1.23:30060",
    "http://10.0.1.23:30061",
    "http://10.0.1.23:30062",
    "http://10.0.1.23:30063",
    "http://10.0.1.23:30064",
    "http://10.0.1.23:30065",
    "http://10.0.1.23:30066",
    "http://10.0.1.23:30067",
    "http://10.0.1.23:30068",
    "http://10.0.1.23:30069",
    "http://10.0.1.23:30070",
    "http://10.0.1.23:30071",
    "http://10.0.1.23:30072",
    "http://10.0.1.23:30073",
    "http://10.0.1.23:30074",
    "http://10.0.1.23:30075",
    "http://10.0.1.23:30076",
    "http://10.0.1.23:30077",
    "http://10.0.1.23:30078",
    "http://10.0.1.23:30079",
    "http://10.0.1.23:30080",
    "http://10.0.1.23:30081",
    "http://10.0.1.23:30082",
    "http://10.0.1.23:30083",
    "http://10.0.1.23:30084",
    "http://10.0.1.23:30085",
    "http://10.0.1.23:30086",
    "http://10.0.1.23:30087",

    "http://10.0.1.100:30048",
    "http://10.0.1.100:30049",
    "http://10.0.1.100:30050",
    "http://10.0.1.100:30089",
    "http://10.0.1.100:30052",
    "http://10.0.1.100:30053",
    "http://10.0.1.100:30054",
    "http://10.0.1.100:30055",
    "http://10.0.1.100:30056",
    "http://10.0.1.100:30057",
    "http://10.0.1.100:30058",
    "http://10.0.1.100:30059",
    "http://10.0.1.100:30060",
    "http://10.0.1.100:30061",
    "http://10.0.1.100:30062",
    "http://10.0.1.100:30063",
    "http://10.0.1.100:30064",
    "http://10.0.1.100:30065",
    "http://10.0.1.100:30066",
    "http://10.0.1.100:30067",
    "http://10.0.1.100:30068",
    "http://10.0.1.100:30069",
    "http://10.0.1.100:30070",
    "http://10.0.1.100:30071",
    "http://10.0.1.100:30072",
    "http://10.0.1.100:30073",
    "http://10.0.1.100:30074",
    "http://10.0.1.100:30075",
    "http://10.0.1.100:30076",
    "http://10.0.1.100:30077",
    "http://10.0.1.100:30078",
    "http://10.0.1.100:30079",
    "http://10.0.1.100:30080",
    "http://10.0.1.100:30081",
    "http://10.0.1.100:30082",
    "http://10.0.1.100:30083",
    "http://10.0.1.100:30084",
    "http://10.0.1.100:30085",
    "http://10.0.1.100:30086",
    "http://10.0.1.100:30087",
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
