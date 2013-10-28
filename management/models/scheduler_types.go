package models

import (
	"github.com/jason-zou/taobaosdk/rest"
	"time"
)

type CrawlerInfo struct {
	Priority int `bson:"priority"`
	Cycle    int `bson:"cycle"`
}

type TaobaoShopExtendedInfo struct {
	Type           string  `bson:"type"`
	Orientational  bool    `bson:"orientational"`
	CommissionRate float32 `bson:"commission_rate"`
}

type ShopItem struct {
	ShopInfo        *rest.Shop              `bson:"shop_info"`
	Status          string                  `bson:"status"`
	CreatedTime     time.Time               `bson:"created_time"`
	LastUpdatedTime time.Time               `bson:"last_updated_time"`
	LastCrawledTime time.Time               `bson:"last_crawled_time"`
	CrawlerInfo     *CrawlerInfo            `bson:"crawler_info"`
	ExtendedInfo    *TaobaoShopExtendedInfo `bson:"extended_info"`
}

type CrawlerData struct {
	Score  int `bson:"score"`
	Volume int `bson:"volume"`
}

type ScoreInfo struct {
    Likes int `bson:"likes"`
    IsSelection bool `bson:"is_selection"`
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
}
