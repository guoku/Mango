package jobs

const (
    START    = "start"
    STOP     = "stop"
    FETCHNEW = "jobs:fetchnew" //在redis里保存fetchnew任务成功爬取商品的集合名
)

const (
    STOP_STATU  = "已经停止"
    START_STATU = "已经启动"
)
