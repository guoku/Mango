package old_guoku_models

type AuthUser struct{
}

type BaseEntity struct {
    Id int `orm:"auto;index"`
    EntityHash string `orm:"unique;index"`
    Brand string `orm:"null"`
    Title string
    Weight int
}

type BaseItem struct {
    Source string
    Entity *BaseEntity `orm:"rel(fk)"`
    Weigth int
}

type BaseTaobaoItem struct {
    TaobaoId string
    TaobaoCategoryId int
    Item *BaseItem `orm:"rel(fk)"`
    Title string
    ShopNick string
}

type GuokuEntityLike struct {
   EntityId int
   UserId int
   CreatedTime time.Time 
}

