package controllers

import (
	"Mango/management/models"
	"encoding/json"
	"fmt"
	"github.com/qiniu/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	//"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const CommodityCodeName = "manage_commodity"

type CommodityController struct {
	UserSessionController
}

func (this *CommodityController) Prepare() {
	this.UserSessionController.Prepare()
	user := this.Data["User"].(*models.User)
	this.Data["Tab"] = &models.Tab{TabName: "Commodity"}
	log.Info("pre", this.Input())
	if !CheckPermission(user.Id, CommodityCodeName) {
		this.Abort("401")
		return
	}
}

type CategoryController struct {
	CommodityController
}

func getSubCats(parentCid int) []int {
	cc := MgoSession.DB(MgoDbName).C("taobao_cats")
	subCats := make([]models.TaobaoItemCat, 0)
	cc.Find(bson.M{"item_cat.parent_cid": parentCid}).All(&subCats)
	subCids := make([]int, 0)
	for _, v := range subCats {
		subCids = append(subCids, v.ItemCat.Cid)
		subCids = append(subCids, getSubCats(v.ItemCat.Cid)...)
	}
	return subCids
}

func getCatsPath(cid int) []*models.TaobaoItemCat {
	cc := MgoSession.DB(MgoDbName).C("taobao_cats")
	tcid := cid
	path := make([]*models.TaobaoItemCat, 0)
	for {
		log.Info("path:", tcid)
		if tcid == 0 {
			break
		}
		cat := models.TaobaoItemCat{}
		err := cc.Find(bson.M{"item_cat.cid": tcid}).One(&cat)
		if err != nil || cat.ItemCat == nil {
			break
		}
		path = append(path, &cat)
		tcid = cat.ItemCat.ParentCid
	}
	rpath := make([]*models.TaobaoItemCat, 0)
	for i := len(path) - 1; i >= 0; i-- {
		rpath = append(rpath, path[i])
	}
	return rpath
}

func (this *CategoryController) Get() {
	cc := MgoSession.DB(MgoDbName).C("taobao_cats")
	query := this.GetString("q")
	if query != "" {
		rq := bson.RegEx{query, "i"}
		cats := make([]models.TaobaoItemCat, 0)
		cc.Find(bson.M{"item_cat.name": rq}).Sort("-item_num").All(&cats)
		this.Data["SearchCats"] = cats
		this.Data["IsSearch"] = true
	} else {
		this.Data["IsSearch"] = false
		num, err := this.GetInt("taobao_cid")
		if err != nil {
			num = 0
		}
		cid := int(num)
		category := models.TaobaoItemCat{}
		if cid != 0 {
			err = cc.Find(bson.M{"item_cat.cid": cid}).One(&category)
			if err != nil || category.ItemCat == nil {
				this.Abort("404")
				return
			}
		}
		num, err = this.GetInt("p")
		if err != nil {
			num = 1
		}
		p := int(num)
		ic := MgoSession.DB(MgoDbName).C("taobao_items_depot")
		//subCids := getSubCats(cid)
		//subCids = append(subCids, cid)
		directSubCats := make([]models.TaobaoItemCat, 0)
		cc.Find(bson.M{"item_cat.parent_cid": cid}).Sort("-item_num").All(&directSubCats)
		items := make([]models.TaobaoItemStd, 0)
		var total int = 0
		//ic.Find(bson.M{"api_data_ready" : true, "api_data.cid" : bson.M{"$in" : subCids}}).Sort("-score").Skip(int((p-1) * NumInOnePage)).Limit(int(NumInOnePage)).All(&items)
		if cid != 0 && !category.ItemCat.IsParent {
			//ic.Find(bson.M{"api_data.cid" : cid}).Sort("-score").Skip(int((p-1) * NumInOnePage)).Limit(int(NumInOnePage)).All(&items)
			ic.Find(bson.M{"cid": cid}).Sort("-score").Skip(int((p - 1) * NumInOnePage)).Limit(int(NumInOnePage)).All(&items)
			/*m := bson.M{}
			  ic.Find(bson.M{"api_data.cid" : cid, "api_data_ready" : true}).Sort("-score").Skip(int((p-1) * NumInOnePage)).Explain(m)
			  fmt.Printf("Explain: %#v\n",m) */
			//ic.Find(bson.M{"api_data.cid" : cid, "api_data_ready" : true}).Skip(int((p-1) * NumInOnePage)).Limit(int(NumInOnePage)).All(&items)
			//total, _ = ic.Find(bson.M{"api_data_ready" : true, "api_data.cid" : bson.M{"$in" : subCids}}).Count()
			total, _ = ic.Find(bson.M{"cid": cid}).Count()
		}
		paginator := models.NewSimplePaginator(p, total, NumInOnePage, this.Input())
		this.Data["Items"] = items
		this.Data["DirectSubCats"] = directSubCats
		if len(directSubCats) == 0 {
			this.Data["HasSubCats"] = false
		} else {
			this.Data["HasSubCats"] = true
		}
		this.Data["CatsPath"] = getCatsPath(cid)
		this.Data["Paginator"] = paginator
		this.Data["Cid"] = cid
	}
	log.Info(this.Data["Tab"])
	this.Layout = DefaultLayoutFile
	this.TplNames = "categories.tpl"
}

type CreateOnlineItemsController struct {
	CommodityController
}

type CreateItemsResp struct {
	ItemId   string `json:"item_id"`
	EntityId string `json:"entity_id"`
	Status   string `json:"status"`
}

func (this *CreateOnlineItemsController) Post() {
	taobaoIds := this.GetString("taobao_ids")
	cid, _ := this.GetInt("cid")
	taobaoIdArr := strings.Split(taobaoIds, ",")
	taobaoCat := models.TaobaoItemCat{}
	cc := MgoSession.DB(MgoDbName).C("taobao_cats")
	cc.Find(bson.M{"item_cat.cid": cid}).One(&taobaoCat)
	if taobaoCat.MatchedGuokuCid == 0 {
		this.Abort("404")
		return
	}

	ic := MgoSession.DB(MgoDbName).C("taobao_items_depot")
	for _, v := range taobaoIdArr {
		taobaoId, _ := strconv.Atoi(v)
		item := models.TaobaoItemStd{}
		err := ic.Find(bson.M{"num_iid": taobaoId}).One(&item)
		if err != nil {
			continue
		}
		params := url.Values{}
		params.Add("taobao_id", v)
		params.Add("cid", strconv.Itoa(item.Cid))
		params.Add("taobao_title", item.Title)
		params.Add("taobao_shop_nick", item.Nick)
		params.Add("taobao_price", fmt.Sprintf("%f", item.Price))
		itemImgs := item.ItemImgs
		if itemImgs != nil && len(itemImgs) > 0 {
			params.Add("chief_image_url", itemImgs[0])
			for i, _ := range itemImgs {
				params.Add("image_url", itemImgs[i])
			}
		}
		params.Add("category_id", strconv.Itoa(taobaoCat.MatchedGuokuCid))
		resp, err := http.PostForm("http://api.guoku.com:10080/management/entity/create/offline/", params)
		if err != nil {
			this.Abort("404")
			return
		}
		body, _ := ioutil.ReadAll(resp.Body)
		r := CreateItemsResp{}
		json.Unmarshal(body, &r)
		log.Info(r)
		if r.Status == "success" {
			ic.Update(bson.M{"num_iid": taobaoId}, bson.M{"$set": bson.M{"item_id": r.ItemId}})
		}
	}
	this.Data["json"] = map[string]string{"status": "succeeded"}
	this.ServeJson()
}

type CategoryManageController struct {
	CommodityController
}

func (this *CategoryManageController) Get() {
	update := this.GetString("update")
	log.Info(this.Input())
	gc := MgoSession.DB(MgoDbName).C("guoku_cats")
	cc := MgoSession.DB(MgoDbName).C("taobao_cats")
	gcg := MgoSession.DB(MgoDbName).C("guoku_cat_groups")
	log.Info("update", update)
	if update == "" {
		guokuCats := make([]models.GuokuCat, 0)
		gc.Find(nil).All(&guokuCats)
		for i, _ := range guokuCats {
			cc.Find(bson.M{"matched_guoku_cid": guokuCats[i].CategoryId}).All(&guokuCats[i].MatchedTaobaoCats)
		}
		this.Data["GuokuCats"] = guokuCats
		this.Layout = DefaultLayoutFile
		this.TplNames = "categories_manage.tpl"
	} else {
		resp, err := http.Get("http://114.113.154.47:8000/management/category/sync/")
		if err != nil {
			log.Info(err.Error())
			this.Redirect("/commodity/category_manage", 302)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Info(err.Error())
			this.Redirect("/commodity/category_manage", 302)
			return
		}
		guokuCatsGroups := make([]models.GuokuCatGroup, 0)
		err = json.Unmarshal(body, &guokuCatsGroups)
		if err != nil {
			log.Info(err.Error())
			this.Redirect("/commodity/category_manage", 302)
			return
		}
		gcg.RemoveAll(nil)
		gc.RemoveAll(nil)
		for _, v := range guokuCatsGroups {
			err = gcg.Insert(&v)
			if err != nil {
				log.Info(err)
			}
			for _, c := range v.Content {
				c.GroupId = v.GroupId
				err = gc.Insert(&c)
				if err != nil {
					log.Info(err)
				}
			}
		}
		this.Redirect("/commodity/category_manage", 302)
	}
}

type AddMatchedCategoryController struct {
	CommodityController
}

func (this *AddMatchedCategoryController) Post() {
	gcid, _ := this.GetInt("guoku_cid")
	tcid, _ := this.GetInt("taobao_cid")
	cc := MgoSession.DB(MgoDbName).C("taobao_cats")
	cc.Update(bson.M{"item_cat.cid": tcid}, bson.M{"$set": bson.M{"matched_guoku_cid": gcid}})
	this.Redirect("/commodity/category_manage/", 302)
}
