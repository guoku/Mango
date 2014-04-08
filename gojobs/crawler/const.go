package crawler

const (
    IMG_POSTFIX string = "_\\d+x\\d+.*\\.jpg|_b\\.jpg"
    HTTP_ERR           = "http err"
    FETCH_ERR          = "fetch err" //通过这个类型的错误，就能找出所有爬取出错的itemid了，因为不是所有出错的地方都能够获取到itemid
    PARSE_ERR          = "parse err"
    SAVE_ERR           = "save err"
    POST_ERR           = "post err"
    SHOP_ERR           = "shop err"
)
