## 各个job的介绍

fetcher.go ：淘宝抓取评论和销售数量  
fetcher.go : 设想是增加一个定时更新评论和销售数量的功能  
statu_update.go : 把爬取过的店铺按照一定的时间，更新为queued  
jdcategory.go : 提取京东的分类目录  
supervisor_listener.py : 定时重启supervisor的爬虫组node  
tbcrawler.toml : fetcher 使用的mongo配置  
get_taobao_cats.go : 统计每个类目下对应的商品数量  
scorer.go : 给每个商品打分  
sync_items.go :把爬取下来的item，当分数高于一定程度，就传到线上  
api_crawler.go : 获取淘宝API信息