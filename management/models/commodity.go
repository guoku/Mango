package models

type ItemGroup struct {
    GroupId int `bson:"group_id"`
    Status string `bson:"status"`
    TaobaoCid int   `bson:"taobao_cid"`
    Vector map[string]float64   `bson:"vector"`
    VectorFreq map[string]int   `bson:"vector_freq"`
    DelegateId int  `bson:"delegate_id"`
    NumItem int `bson:"num_item"`
    AveragePrice float32    `bson:"average_price"`
}

type ItemExtendedInfo struct {
}
