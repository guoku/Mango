package controllers

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
	"Mango/management/models"
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
    this.Data["Tab"] = &models.Tab{TabName : "Commodity"}
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
        fmt.Println("path:", tcid)
        if tcid == 0 {
            break
        }
        cat := models.TaobaoItemCat{}
        cc.Find(bson.M{"item_cat.cid": tcid}).One(&cat)
        path = append(path, &cat)
        tcid = cat.ItemCat.ParentCid
    }
    rpath := make([]*models.TaobaoItemCat, 0)
    for i := len(path) -1 ; i >= 0; i-- {
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
        cc.Find(bson.M{"item_cat.name" : rq}).Sort("-item_num").All(&cats)
        this.Data["SearchCats"] = cats
        this.Data["IsSearch"] = true
    } else {
        this.Data["IsSearch"] = false
        num, err := this.GetInt("taobao_cid")
        if err != nil {
            num = 0
        }
        cid := int(num)
        num, err = this.GetInt("p")
        if err != nil {
            num = 1
        }
        p := int(num)
        ic := MgoSession.DB(MgoDbName).C("raw_taobao_items_depot")
        subCids := getSubCats(cid)
        subCids = append(subCids, cid)
        directSubCats := make([]models.TaobaoItemCat, 0)
        cc.Find(bson.M{"item_cat.parent_cid": cid}).Sort("-item_num").All(&directSubCats)
        items := make([]models.TaobaoItem, 0)
        var total int
        fmt.Println("start query")
            //ic.Find(bson.M{"api_data_ready" : true, "api_data.cid" : bson.M{"$in" : subCids}}).Sort("-score").Skip(int((p-1) * NumInOnePage)).Limit(int(NumInOnePage)).All(&items)
        ic.Find(bson.M{"api_data.cid" : cid, "api_data_ready" : true}).Sort("-score").Skip(int((p-1) * NumInOnePage)).Limit(int(NumInOnePage)).All(&items)
        //ic.Find(bson.M{"api_data.cid" : cid, "api_data_ready" : true}).Skip(int((p-1) * NumInOnePage)).Limit(int(NumInOnePage)).All(&items)
            //total, _ = ic.Find(bson.M{"api_data_ready" : true, "api_data.cid" : bson.M{"$in" : subCids}}).Count()
        total, _ = ic.Find(bson.M{"api_data.cid" : cid, "api_data_ready" : true}).Count()
        fmt.Println("finish query")
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
    }
    fmt.Println(this.Data["Tab"])
    this.Layout = DefaultLayoutFile
    this.TplNames = "categories.tpl"
}

type CategoryManageController struct {
    CommodityController
}

func (this *CategoryManageController) Get() {
    update := this.GetString("update")
    gc := MgoSession.DB(MgoDbName).C("guoku_cats")
    cc := MgoSession.DB(MgoDbName).C("taobao_cats")
    gcg := MgoSession.DB(MgoDbName).C("guoku_cat_groups")
    if update == "" {
        guokuCats := make([]models.GuokuCat, 0)
        gc.Find(nil).All(&guokuCats)
        for i, _ := range guokuCats {
            cc.Find(bson.M{"matched_guoku_cid" : guokuCats[i].CategoryId}).All(&guokuCats[i].MatchedTaobaoCats)
        }
        this.Data["GuokuCats"] = guokuCats
        this.Layout = DefaultLayoutFile
        this.TplNames = "categories_manage.tpl"
    } else {
        resp, err := http.Get("http://api.guoku.com:10080/management/category/sync/")
        if err != nil {
            fmt.Println(err.Error())
            this.Redirect("/commodity/category_manage", 301)
            return
        }
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            fmt.Println(err.Error())
            this.Redirect("/commodity/category_manage", 301)
            return
        }
        guokuCatsGroups := make([]models.GuokuCatGroup, 0)
        err = json.Unmarshal(body, &guokuCatsGroups)
        if err != nil {
            fmt.Println(err.Error())
            this.Redirect("/commodity/category_manage", 301)
            return
        }
        gcg.RemoveAll(nil)
        gc.RemoveAll(nil)
        for _, v := range guokuCatsGroups {
            gcg.Insert(&v)
            for _, c := range v.Content {
                c.GroupId = v.GroupId
                gc.Insert(&c)
            }
        }
        this.Redirect("/commodity/category_manage", 301)
    }
}

type AddMatchedCategoryController struct {
    CommodityController
}

func (this *AddMatchedCategoryController) Post() {
    gcid, _ := this.GetInt("guoku_cid")
    tcid, _ := this.GetInt("taobao_cid")
    cc := MgoSession.DB(MgoDbName).C("taobao_cats")
    cc.Update(bson.M{"item_cat.cid" : tcid}, bson.M{"$set" : bson.M{"matched_guoku_cid" : gcid}})
    this.Redirect("/commodity/category_manage", 301)
}
