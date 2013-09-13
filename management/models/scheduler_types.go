package models

import (
    "time"
    "github.com/jason-zou/taobaosdk/rest"
)

type ShopItem struct {
	ShopInfo        *rest.Shop   `bson:"shop_info"`
	Status          string      `bson:"status"`
	CreatedTime     time.Time   `bson:"created_time"`
	LastUpdatedTime time.Time   `bson:"last_updated_time"`
	LastCrawledTime time.Time   `bson:"last_crawled_time"`
}

type CrawlerData struct {
	Score  int  `bson:"score"`
	Volume int  `bson:"volume"`
}

type TaobaoItem struct {
	Sid                    int          `bson:"sid"`
	NumIid                 int          `bson:"num_iid"`
	ApiDataReady           bool         `bson:"api_data_ready"`
	CrawlerDataUpdatedTime time.Time    `bson:"crawler_data_updated_time"`
	ApiDataUpdatedTime     time.Time    `bson:"api_data_updated_time"`
	ApiData                *rest.Item    `bson:"api_data"`
	CrawlerData            *CrawlerData  `bson:"crawler_data"`
	CreatedTime            time.Time    `bson:"created_time"`
}

