package models

import (
	"github.com/jason-zou/taobaosdk/rest"
	"labix.org/v2/mgo/bson"
	"time"
)

type CrawlerInfo struct {
	Priority int `bson:"priority"`
	Cycle    int `bson:"cycle"`
}

type ShopScoreInfo struct {
	TotalLikes      int       `bson:"total_likes"`
	TotalSelections int       `bson:"total_selections"`
	TotalItems      int       `bson:"total_items"`
	UpdatedTime     time.Time `bson:"updated_time"`
}
type TaobaoShopExtendedInfo struct {
	Type           string   `bson:"type"`
	Orientational  bool     `bson:"orientational"`
	CommissionRate float32  `bson:"commission_rate"`
	Original       bool     `bson:"original"`
	SingleTail     bool     `bson:"single_tail"` //原单
	Gifts          []string `bson:"gifts"`
	Commission     bool     `bson:"commission"`
}

type ShopItem struct {
	ShopInfo        *rest.Shop              `bson:"shop_info"`
	Status          string                  `bson:"status"`
	CreatedTime     time.Time               `bson:"created_time"`
	LastUpdatedTime time.Time               `bson:"last_updated_time"`
	LastCrawledTime time.Time               `bson:"last_crawled_time"`
	CrawlerInfo     *CrawlerInfo            `bson:"crawler_info"`
	ExtendedInfo    *TaobaoShopExtendedInfo `bson:"extended_info"`
	ScoreInfo       *ShopScoreInfo          `bson:"score_info"`
}

type CrawlerData struct {
	Score  int `bson:"score"`
	Volume int `bson:"volume"`
}

type ScoreInfo struct {
	Likes       int       `bson:"likes"`
	IsSelection bool      `bson:"is_selection"`
	UpdatedTime time.Time `bson:"updated_time"`
}

type TaobaoItem struct {
	Sid                    int          `bson:"sid"`
	NumIid                 int          `bson:"num_iid"`
	ApiDataReady           bool         `bson:"api_data_ready"`
	CrawlerDataUpdatedTime time.Time    `bson:"crawler_data_updated_time"`
	ApiDataUpdatedTime     time.Time    `bson:"api_data_updated_time"`
	ApiData                *rest.Item   `bson:"api_data"`
	CrawlerData            *CrawlerData `bson:"crawler_data"`
	CreatedTime            time.Time    `bson:"created_time"`
	ScoreInfo              *ScoreInfo   `bson:"score_info"`
	Score                  float64      `bson:"score"`
	ScoreUpdatedTime       time.Time    `bson:"score_updated_time"`
	ItemId                 string       `bson:"item_id"`
	Uploaded               bool         `bson:"uploaded"`
}

type TaobaoItemStd struct {
	NumIid           int               `bson:"num_iid" json:"num_iid"`
	Title            string            `bson:"title" json:"title"`
	Nick             string            `bson:"nick" json:"nick"`
	Desc             string            `bson:"desc" json:"desc"`
	Cid              int               `bson:"cid" json:"cid"`
	Sid              int               `bson:"sid" son:"sid"`
	Price            float32           `bson:"price" json:"price"`
	Location         *rest.Location    `bson:"location" json:"location"`
	PromotionPrice   float32           `bson:"promotion_price" json:"promotion_price"`
	ItemImgs         []string          `bson:"item_imgs" json:"item_imgs"`
	ShopType         string            `bson:"shop_type" json:"shop_type"`
	ReviewsCount     int               `bson:"reviews_count" json:"reviews_count"`
	MonthlySalesNum  int               `bson:"monthly_sales_num" json:"monthly_sales_num"`
	Props            map[string]string `bson:"props" json:"props"`
	InStock          bool              `bson:"in_stock" json:"in_stock"`
	CreatedTime      time.Time         `bson:"created_time"`
	DataUpdatedTime  time.Time         `bson:"data_updated_time"`
	ScoreInfo        *ScoreInfo        `bson:"score_info"`
	Score            float64           `bson:"score"`
	ScoreUpdatedTime time.Time         `bson:"score_updated_time"`
	ItemId           string            `bson:"item_id"`
	Uploaded         bool              `bson:"uploaded"`
	GroupId          int               `bson:"group_id"`
	Refreshed        bool              `bson:"refreshed"` //对一个商品再次抓取后，设置为false，上传后设为true
	RefreshTime      time.Time         `bson:"refresh_time"`
	DetailUrl        string            `bson:"detail_url"`
}

type TaobaoProp struct {
	TaobaoId int    `bson:"taobao_id"`
	Name     string `bson:"name"`
	Type     string `bson:"type"`
}

type TaobaoItemCat struct {
	Id              bson.ObjectId `bson:"_id"`
	ItemCat         *rest.ItemCat `bson:"item_cat"`
	ItemNum         int           `bson:"item_num"`
	SelectionNum    int           `bson:"selection_num"`
	MatchedGuokuCid int           `bson:"matched_guoku_cid"`
	UpdatedTime     time.Time     `bson:"updated_time"`
}

type GuokuCat struct {
	CategoryId        int              `json:"category_id" bson:"category_id"`
	IconSmall         string           `json:"category_icon_small" bson:"icon_small"`
	Title             string           `json:"category_title" bson:"title"`
	IconLarge         string           `json:"category_icon_large" bson:"icon_large"`
	GroupId           int              `json:"-" bson:"group_id"`
	MatchedTaobaoCats []*TaobaoItemCat `json:"-" bson:"-"`
}

type GuokuCatGroup struct {
	Status        int         `json:"status" bson:"status"`
	Content       []*GuokuCat `json:"content" bson:"-"`
	CategoryCount int         `json:"category_count" bson:"category_count"`
	GroupId       int         `json:"group_id" bson:"group_id"`
	Title         string      `json:"title" bson:"title"`
}

//垃圾词
type DictWord struct {
	Word        string  `bson:"word"`
	Freq        int     `bson:"freq"`
	Prob        float64 `bson:"prob"`
	Weight      float64 `bson:"weight"`
	Blacklisted bool    `bson:"blacklisted"`
	Deleted     bool    `bson:"deleted"`
	Type        string  `bson:"type"`
}

//品牌词
type BrandsWord struct {
	Name    string `bson:"name"`
	Freq    int    `bson:"freq"`
	Type    string `bson:"type"`
	Deleted bool   `bson:"deleted"`
	Valid   bool   `bson:"valid"`
}
