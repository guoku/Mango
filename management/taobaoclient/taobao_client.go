package taobaoclient

import (
    "fmt"
    "github.com/jason-zou/taobaosdk"
    "github.com/jason-zou/taobaosdk/rest"
)

var GUOKU_MOBI_APP_INFO taobaosdk.AppInfo = taobaosdk.AppInfo{"21419640", "df91464ae934bacca326450f8ade67f7"}

func GetTaobaoShopInfo(nick string) (*rest.Shop, *taobaosdk.TopError) {
    r := rest.ShopGetRequest{}
    r.SetAppInfo(GUOKU_MOBI_APP_INFO.AppKey,
                 GUOKU_MOBI_APP_INFO.Secret)
    r.SetNick(nick)
    r.SetFields("sid,cid,nick,title,pic_path,created,modified,shop_score")
    resp, progErr, topErr := r.GetResponse()
    if progErr != nil {
        fmt.Println(progErr.Error())
        return nil, topErr
    }
    if topErr != nil || resp == nil {
        return nil, topErr
    }
    return resp.Shop, topErr
}

func GetTaobaoItemInfo(numIid int) (*rest.Item, *taobaosdk.TopError) {
    r := rest.ItemGetRequest{}
    r.SetAppInfo(GUOKU_MOBI_APP_INFO.AppKey,
                 GUOKU_MOBI_APP_INFO.Secret)
    r.SetNumIid(numIid)
	r.SetFields("detail_url,num_iid,title,nick,type,desc,cid,pic_url,num,list_time,delist_time,stuff_status,location,price,global_stock_type,item_imgs,item_img,skus,sku,props_name,prop_imgs,prop_img")
    resp, progErr, topErr := r.GetResponse()
    if progErr != nil {
        fmt.Println(progErr.Error())
        return nil, topErr
    }
    if topErr != nil || resp == nil {
        return nil, topErr
    }
    return resp.Item, topErr
}


func GetItemCatsInfo(parentCid int) ([]*rest.ItemCat, *taobaosdk.TopError) {
    r := rest.ItemCatsGetRequest{}
    r.SetAppInfo(GUOKU_MOBI_APP_INFO.AppKey,
                 GUOKU_MOBI_APP_INFO.Secret)
    r.SetParentCid(parentCid)
    r.SetFields("features,attr_key,attr_value,cid,parent_cid,name,is_parent,status,sort_order")
    resp, progErr, topErr := r.GetResponse()
    if progErr != nil {
        fmt.Println(progErr.Error())
        return nil, topErr
    }
    if topErr != nil || resp == nil {
        return nil, topErr
    }
    for _, v := range resp.ItemCats.ItemCatArray {
        fmt.Println(v.Cid)
    }
    return resp.ItemCats.ItemCatArray, topErr
}
